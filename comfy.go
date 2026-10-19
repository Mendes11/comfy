package comfy

import (
	"fmt"
	"os"
	"reflect"

	"github.com/Mendes11/comfy/dotenv"
)

const DEFAULT_ENVIRONMENT_ENV = "APP_ENV"

var appEnvName = DEFAULT_ENVIRONMENT_ENV

func SetEnv(env string) {
	appEnvName = env
}

func Environment() string {
	return os.Getenv(appEnvName)
}

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

func setup() error {
	if err := dotenv.LoadDotEnvs(Environment); err != nil {
		return err
	}
	return nil
}
