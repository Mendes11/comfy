package parsers

import "reflect"

// ParseFunc converts a string to the target's value, and sets it inplace
type ParseFunc func(s string, target reflect.Value) error

var parserRegistry = map[reflect.Kind]ParseFunc{}

func Get(value reflect.Value) (ParseFunc, bool) {
	parser, exists := parserRegistry[value.Kind()]
	return parser, exists
}
