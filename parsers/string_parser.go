package parsers

import "reflect"

func init() {
	parsersByKindRegistry[reflect.String] = parseString
}

func parseString(s string, target reflect.Value, opts ParserOpts) error {
	setValue(target, reflect.ValueOf(s))
	return nil
}
