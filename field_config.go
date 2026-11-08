package comfy

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/Mendes11/comfy/parsers"
)

const DefaultFileSuffix = "_FILE"

var DefaultFieldConfig = FieldConfig{
	FileSuffix:            DefaultFileSuffix,
	AutoFileLookupEnabled: true,
	NoPrefix:              false,
	Format:                "",
}

// FieldConfig holds how fields are going to be loaded and looked up
type FieldConfig struct {
	CustomEnvName         *string // set by the comfy tag
	FileSuffix            string
	NoPrefix              bool
	Format                string
	AutoFileLookupEnabled bool
}

func NewFieldConfig() FieldConfig {
	return DefaultFieldConfig
}

func (f FieldConfig) HasCustomEnvName() bool {
	return f.CustomEnvName != nil && *f.CustomEnvName != ""
}

func (f FieldConfig) AsParserOpts() parsers.ParserOpts {
	return parsers.ParserOpts{
		DateFormat: f.Format,
	}
}

type tagConfigName string

const (
	tagConfNoPrefix   tagConfigName = "no-prefix"
	tagConfFormat     tagConfigName = "format"
	tagConfFileSuffix tagConfigName = "file-suffix"
)

type tagConfigApply func(*FieldConfig, string) error

// tagAppliers maps every possible Comfy tag settings to a function that modifies FieldConfig
var tagAppliers = map[tagConfigName]tagConfigApply{
	tagConfNoPrefix:   noPrefixApply,
	tagConfFormat:     formatApply,
	tagConfFileSuffix: fileSuffixApply,
}

var noPrefixApply = func(c *FieldConfig, value string) error {
	c.NoPrefix = true
	return nil
}

var formatApply = func(c *FieldConfig, value string) error {
	c.Format = value
	return nil
}

var fileSuffixApply = func(c *FieldConfig, value string) error {
	c.FileSuffix = value
	return nil
}

func fieldTagConfigs(structField reflect.StructField) (FieldConfig, error) {
	conf := NewFieldConfig()
	tagValuesStr, exists := structField.Tag.Lookup("comfy")
	if !exists {
		return conf, nil
	}
	tagValues := strings.Split(tagValuesStr, ",")

	if len(tagValues) == 0 {
		emptyName := ""
		conf.CustomEnvName = &emptyName
		return conf, nil
	}

	conf.CustomEnvName = &tagValues[0]
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
