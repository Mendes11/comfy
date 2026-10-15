package comfy

import (
	"reflect"
	"strconv"
)

// ParseFunc converts a string to the target's value, and sets it inplace
type ParseFunc func(s string, target reflect.Value) error

var parserRegistry = map[reflect.Kind]ParseFunc{}

// Parser interface parses a given string inplace.
type ComfyParser interface {
	ComfyParse(s string) error
}

var parserType = reflect.TypeOf((*ComfyParser)(nil)).Elem()

// parse will convert a string to its target value inplace.
//
// When it doesn't have a builtin parser, or it doesn't implement ComfyParser, it returns a MissingParserError.
//
// If the parse fails, then TypeMismatchError is returned.
func parse(fieldName, value string, target reflect.Value) error {
	nonPtrTarget := target
	if target.Type().Kind() == reflect.Pointer {
		nonPtrTarget = target.Elem()
	}
	if parser, ok := parserRegistry[nonPtrTarget.Kind()]; ok {
		err := parser(value, target)
		if err != nil {
			err = TypeMismatchError{
				FieldName:  fieldName,
				FieldValue: value,
				TargetType: target.Type(),
				Err:        err,
			}
		}
		return err
	}

	if target.Type().Implements(parserType) {
		ret := target.MethodByName("ComfyParser").Call([]reflect.Value{reflect.ValueOf(value)})
		err := ret[0].Interface().(error)
		if err != nil {
			err = TypeMismatchError{
				FieldName:  fieldName,
				FieldValue: value,
				TargetType: target.Type(),
				Err:        err,
			}
		}
		return err
	}

	return MissingParserError{TargetType: target.Type()}
}

// Builtin Parsers

var intParser ParseFunc = func(s string, target reflect.Value) error {
	v, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	target.Set(reflect.ValueOf(v))
	return nil
}

var stringParser ParseFunc = func(s string, target reflect.Value) error {
	target.Set(reflect.ValueOf(s))
	return nil
}

func init() {
	parserRegistry[reflect.String] = stringParser
	parserRegistry[reflect.Int] = intParser
}
