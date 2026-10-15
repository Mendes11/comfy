package parsers

import (
	"reflect"
	"strconv"
)

func init() { parserRegistry[reflect.Int] = IntParser }

var IntParser = func(s string, target reflect.Value) error {
	v, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	target.Set(reflect.ValueOf(v))
	return nil
}
