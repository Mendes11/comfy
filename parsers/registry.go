package parsers

import (
	"reflect"
	"regexp"
)

type ParserOpts struct {
	DateFormat string
}

// ParseFunc converts a string to the target's value, and sets it inplace
type ParseFunc func(s string, target reflect.Value, opts ParserOpts) error

var (
	parsersByKindRegistry           = map[reflect.Kind]ParseFunc{}
	parsersByTypeNameRegistry       = map[string]ParseFunc{}
	parsersByTypeNameRegexpRegistry = map[*regexp.Regexp]ParseFunc{}
)

type RegistryType int

const (
	RegistryTypeKind = iota
	RegistryTypeName
	RegistryTypeRegexp
)

// GetForType a parser for the given reflect.Type
func GetForType(value reflect.Type) (parser ParseFunc, exists bool) {
	typeName := value.Name()
	parser, exists = parsersByTypeNameRegistry[typeName]
	if exists {
		return
	}
	for k, v := range parsersByTypeNameRegexpRegistry {
		if k.MatchString(typeName) {
			return v, true
		}
	}
	parser, exists = parsersByKindRegistry[value.Kind()]
	return
}

// Set a parser based on the key K.
// It'll replace in case of the key being already set.
// The function accepts K of type reflect.Kind or string:
//
//  1. When reflect.Kind, it registers that kind into the parsers registry. And comfy checks the field Kind against it.
//
//  2. When string, it registers in the type names registry. Comfy will compare the field's type in string format against the registry.
//
//     - Eg: time.Time{} => "time.Time"
//
//     - This is useful for when you want to give a parser to a type you can't give it ComfyParser interface (a type you don't control)
func Set[K reflect.Kind | string](registry RegistryType, key K, parser ParseFunc) {
	switch k := any(key).(type) {
	case reflect.Kind:
		parsersByKindRegistry[k] = parser
	case string:
		parsersByTypeNameRegistry[k] = parser
	}
}

// setValue ensures we correctly convert the value to the user type
func setValue(target reflect.Value, value reflect.Value) {
	if value.Type() != target.Type() {
		value = value.Convert(target.Type())
	}
	target.Set(value)
}
