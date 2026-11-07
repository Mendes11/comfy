package comfy

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/Mendes11/comfy/parsers"
)

const DefaultFileSuffix = "_FILE"

// FieldConfig holds all tag configurations for a single struct field
type FieldConfig struct {
	customEnvName *string
	fileSuffix    string
	noPrefix      bool
	format        string
}

func (f FieldConfig) HasCustomEnvName() bool {
	return f.customEnvName != nil && *f.customEnvName != ""
}

func (f FieldConfig) AsParserOpts() parsers.ParserOpts {
	return parsers.ParserOpts{
		DateFormat: f.format,
	}
}

type tagConfigName string

const (
	tagConfNoPrefix tagConfigName = "no-prefix"
	tagConfFormat   tagConfigName = "format"
)

type tagConfigApply func(*FieldConfig, string) error

// tagAppliers maps every possible Comfy tag settings to a function that modifies FieldConfig
var tagAppliers = map[tagConfigName]tagConfigApply{
	tagConfNoPrefix: noPrefixApply,
	tagConfFormat:   formatApply,
}

var noPrefixApply = func(c *FieldConfig, value string) error {
	c.noPrefix = true
	return nil
}

var formatApply = func(c *FieldConfig, value string) error {
	c.format = value
	return nil
}

func fieldTagConfigs(structField reflect.StructField) (FieldConfig, error) {
	conf := FieldConfig{
		fileSuffix: DefaultFileSuffix,
	}
	tagValuesStr, exists := structField.Tag.Lookup("comfy")
	if !exists {
		return conf, nil
	}
	tagValues := strings.Split(tagValuesStr, ",")

	if len(tagValues) == 0 {
		emptyName := ""
		conf.customEnvName = &emptyName
		return conf, nil
	}

	conf.customEnvName = &tagValues[0]
	for _, tagConf := range tagValues[1:] {
		tagConfValues := strings.Split(tagConf, "=")
		tagConfName := tagConfValues[0]
		tagConfValue := ""
		if len(tagConfValues) == 2 {
			tagConfValue = tagConfValues[0]
		}
		applier, exists := tagAppliers[tagConfigName(tagConfName)]
		if !exists {
			return conf, fmt.Errorf("unsupported comfy tag configuration %s", tagConf)
		}
		err := applier(&conf, tagConfValue)
		if err != nil {
			return conf, err
		}
	}
	return conf, nil
}
