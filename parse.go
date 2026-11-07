package comfy

import (
	"fmt"
	"reflect"
	"strings"

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
	valueType := nonPtrTarget.Type()

	if implementsComfyParser(nonPtrTarget) {
		ret := nonPtrTarget.Addr().MethodByName("ComfyParse").Call([]reflect.Value{reflect.ValueOf(value)})
		if !ret[0].IsNil() {
			err := ret[0].Interface().(error)
			err = TypeMismatchError{
				FieldName:  fieldName,
				FieldValue: value,
				TargetType: target.Type(),
				Err:        err,
			}
			return err
		}
		return nil

	} else if valueType.Kind() == reflect.Slice {
		values := strings.Split(value, ",")
		for i, v := range values {
			sliceValue := reflect.New(valueType.Elem()).Elem()
			if err := parse(fmt.Sprintf("%s[%d]", fieldName, i), v, sliceValue); err != nil {
				return err
			}
			nonPtrTarget.Set(reflect.Append(nonPtrTarget, sliceValue))
		}
		return nil
	} else if parser, ok := parsers.Get(valueType); ok {
		err := parser(value, target)
		if err != nil {
			err = TypeMismatchError{
				FieldName:  fieldName,
				FieldValue: value,
				TargetType: valueType,
				Err:        err,
			}
		}
		return err
	}

	return MissingParserError{TargetType: target.Type()}
}

func implementsComfyParser(value reflect.Value) bool {
	return value.Type().Implements(parserType) || value.Addr().Type().Implements(parserType)
}
