package comfy

import (
	"fmt"
	"reflect"
	"strings"
)

// FieldConfig holds all configurations of a single struct field
type FieldConfig struct {
	envName  *string
	noPrefix bool
}

type tagConfigName string

const (
	tagConfNoPrefix tagConfigName = "no-prefix"
)

type tagConfigApply func(*FieldConfig) error

// tagAppliers maps every possible Comfy tag settings to a function that modifies FieldConfig
var tagAppliers = map[tagConfigName]tagConfigApply{
	tagConfNoPrefix: noPrefixApply,
}

var noPrefixApply = func(c *FieldConfig) error {
	c.noPrefix = true
	return nil
}

func fieldTagConfigs(structField reflect.StructField) (FieldConfig, error) {
	conf := FieldConfig{}
	tagValuesStr, exists := structField.Tag.Lookup("comfy")
	if !exists {
		return conf, nil
	}
	tagValues := strings.Split(tagValuesStr, ",")

	if len(tagValues) == 0 {
		emptyName := ""
		conf.envName = &emptyName
		return conf, nil
	}

	conf.envName = &tagValues[0]
	for _, tagConf := range tagValues[1:] {
		applier, exists := tagAppliers[tagConfigName(tagConf)]
		if !exists {
			return conf, fmt.Errorf("unsupported comfy tag configuration %s", tagConf)
		}
		err := applier(&conf)
		if err != nil {
			return conf, err
		}
	}
	return conf, nil
}
