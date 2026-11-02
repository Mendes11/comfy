package parsers

import "reflect"

func init() { parserRegistry[reflect.String] = StringParser }

var StringParser = func(s string, target reflect.Value) error {
	setValue(target, reflect.ValueOf(s))
	return nil
}

// setValue ensures we correctly convert the value to the user type
func setValue(target reflect.Value, value reflect.Value) {
	if value.Type() != target.Type() {
		value = value.Convert(target.Type())
	}
	target.Set(value)
}
