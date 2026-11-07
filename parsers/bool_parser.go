package parsers

import (
	"reflect"
	"strings"
)

func init() {
	parsersByKindRegistry[reflect.Bool] = parseBool
}

func parseBool(s string, target reflect.Value, opts ParserOpts) error {
	value := strings.ToLower(s) == "true"
	setValue(target, reflect.ValueOf(value))
	return nil
}
