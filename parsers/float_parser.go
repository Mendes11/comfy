package parsers

import (
	"reflect"
	"strconv"
)

func init() {
	parsersByKindRegistry[reflect.Float32] = floatParserFactory[float32](32)
	parsersByKindRegistry[reflect.Float64] = floatParserFactory[float64](64)
}

type Floats interface {
	float32 | float64
}

func floatParserFactory[K Floats](bitSize int) ParseFunc {
	return func(s string, target reflect.Value, opts ParserOpts) error {
		v, err := strconv.ParseFloat(s, bitSize)
		if err != nil {
			return err
		}
		vType := K(v)
		setValue(target, reflect.ValueOf(vType))
		return nil
	}
}
