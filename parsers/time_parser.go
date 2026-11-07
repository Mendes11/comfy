package parsers

import (
	"fmt"
	"reflect"
	"time"
)

func init() {
	Set(RegistryTypeName, "time.Time", parseTime)
}

func parseTime(s string, target reflect.Value, opts ParserOpts) error {
	t, err := time.Parse(opts.DateFormat, s)
	if err != nil {
		return fmt.Errorf("comfy.parseTime failed for format %s: %w", opts.DateFormat, err)
	}
	setValue(target, reflect.ValueOf(t))
	return nil
}
