package comfy

import (
	"fmt"
	"os"
	"reflect"

	"github.com/Mendes11/comfy/dotenv"
)

const DefaultEnvironmentEnv = "APP_ENV"

var (
	appEnvName   = DefaultEnvironmentEnv
	manualAppEnv = ""
)

// SetAppEnvName changes the [DefaultEnvironmentEnv] comfy uses to know what's the current system environment
func SetAppEnvName(name string) {
	appEnvName = name
}

// SetAppEnvironment configures the application environment for loading its defaults and variables
// This is optional, use it only if you have different app environments.
// By default, the application loads the environment from APP_ENV, or what's set on [SetAppEnvName]
func SetAppEnvironment(env string) {
	manualAppEnv = env
}

// AppEnvironment retrieves the application environment if it's set
// It'll first resolve any environment set through [SetAppEnvironment], and when it's not set,
// it uses the app environment env var
func AppEnvironment() string {
	if manualAppEnv != "" {
		return manualAppEnv
	}
	return os.Getenv(appEnvName)
}

// Load the values into the given configuration variable.
// conf must be a pointer
func Load(conf any) error {
	if err := setup(); err != nil {
		return fmt.Errorf("comfy.Load setup failed: %w", err)
	}

	confValue := reflect.ValueOf(conf)
	if confValue.Kind() != reflect.Pointer {
		return ErrExpectedPtr
	}
	if confValue.Elem().Kind() != reflect.Struct {
		return ErrExpectedStructType
	}
	for structField, value := range confValue.Elem().Fields() {
		f, err := NewField(structField, value, nil)
		if err != nil {
			return err
		}
		_, err = f.Load()
		if err != nil {
			return err
		}
	}
	return nil
}

// LoadConfig will return a new configuration object based on the type parameter T.
// If an environment default is set, it'll initialize the values based on it.
func LoadConfig[T any]() (T, error) {
	conf := *new(T)
	if err := setup(); err != nil {
		return conf, fmt.Errorf("comfy.LoadConfig setup failed: %w", err)
	}
	if envDefault, exists := environments[AppEnvironment()]; exists {
		fmt.Println("env exists")
		conf = envDefault.(T)
	}
	if err := Load(&conf); err != nil {
		return conf, fmt.Errorf("comfy.LoadConfig failed: %w", err)
	}
	return conf, nil
}

func setup() error {
	if err := dotenv.LoadDotEnvs(AppEnvironment); err != nil {
		return err
	}
	return nil
}
