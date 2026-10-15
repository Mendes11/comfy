package parsers

import "reflect"

func init() { parserRegistry[reflect.String] = StringParser }

var StringParser = func(s string, target reflect.Value) error {
	target.Set(reflect.ValueOf(s))
	return nil
}
