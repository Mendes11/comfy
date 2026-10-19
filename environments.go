package comfy

import (
	"fmt"
	"reflect"
)

var environments = map[string]any{}

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
