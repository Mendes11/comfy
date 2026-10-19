package comfy

import (
	"fmt"
	"os"
	"reflect"
)

var environments = map[string]any{}

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

// SetEnvironmentDefault assigns a default value to be used when the app loads configurations under that app environment.
// eg: SetEnvironmentDefault("staging", StagingConfigs)
// Use this in with [LoadConfig] to have the defaults loaded
func SetEnvironmentDefault(name string, conf any) {
	if err := validateEnvironment(conf); err != nil {
		panic(err)
	}
	environments[name] = conf
}

func validateEnvironment(conf any) error {
	for _, v := range environments {
		if reflect.TypeOf(v).Name() != reflect.TypeOf(v).Name() {
			return fmt.Errorf("%T should match the other environments' type %T", conf, v)
		}
	}
	return nil
}
