package comfy

import (
	"reflect"

	"github.com/Mendes11/comfy/parsers"
)

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
	if parser, ok := parsers.Get(nonPtrTarget); ok {
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
