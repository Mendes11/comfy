package parsers

import (
	"reflect"
	"strconv"
)

func init() {
	parserRegistry[reflect.Int8] = intParserFactory[int8](8)
	parserRegistry[reflect.Int16] = intParserFactory[int16](16)
	parserRegistry[reflect.Int32] = intParserFactory[int32](32)
	parserRegistry[reflect.Int] = intParserFactory[int](32)
	parserRegistry[reflect.Int64] = intParserFactory[int64](64)

	// Uints
	parserRegistry[reflect.Uint8] = uintParserFactory[uint8](8)
	parserRegistry[reflect.Uint16] = uintParserFactory[uint16](16)
	parserRegistry[reflect.Uint32] = uintParserFactory[uint32](32)
	parserRegistry[reflect.Uint] = uintParserFactory[uint](32)
	parserRegistry[reflect.Uint64] = uintParserFactory[uint64](64)
}

type Ints interface {
	int | int8 | int16 | int32 | int64
}

type UInts interface {
	uint | uint8 | uint16 | uint32 | uint64
}

func intParserFactory[K Ints](bitSize int) func(string, reflect.Value) error {
	return func(s string, target reflect.Value) error {
		v, err := strconv.ParseInt(s, 10, bitSize)
		if err != nil {
			return err
		}
		vType := K(v)
		target.Set(reflect.ValueOf(vType))
		return nil
	}
}

func uintParserFactory[K UInts](bitSize int) func(string, reflect.Value) error {
	return func(s string, target reflect.Value) error {
		v, err := strconv.ParseUint(s, 10, bitSize)
		if err != nil {
			return err
		}
		vType := K(v)
		target.Set(reflect.ValueOf(vType))
		return nil
	}
}
