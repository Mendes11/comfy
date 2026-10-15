package comfy

import (
	"reflect"
)

func Load(conf any) error {
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
