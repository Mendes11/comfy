package comfy

import (
	"fmt"
	"reflect"

	"github.com/Mendes11/comfy/dotenv"
)

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
