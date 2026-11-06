package parsers

import (
	"reflect"
	"strings"
)

func init() {
	parserRegistry[reflect.Bool] = parseBool
}

func parseBool(s string, target reflect.Value) error {
	value := strings.ToLower(s) == "true"
	setValue(target, reflect.ValueOf(value))
	return nil
}
