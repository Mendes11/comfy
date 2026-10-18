package comfy

import (
	"fmt"
	"os"
	"reflect"
	"regexp"
)

var fileFieldMatcher = regexp.MustCompile(`^FileField\[.*?\]$`)

// Field represents a single configurable field loaded by Comfy
type Field struct {
	structField reflect.StructField
	value       reflect.Value
	isPtr       bool

	parent *Field
	config FieldConfig
}

// NewField returns an instance of a Comfy Field
func NewField(structField reflect.StructField, value reflect.Value, parent *Field) (Field, error) {
	conf, err := fieldTagConfigs(structField)
	if err != nil {
		return Field{}, err
	}
	return Field{
		structField: structField,
		value:       value,
		isPtr:       value.Kind() == reflect.Pointer,
		parent:      parent,
		config:      conf,
	}, nil
}

// Load will load this field from the configuration source, and returns true when the configuration exists and was loaded.
// When no error, but also no configuration matched for this field, it returns false.
func (f Field) Load() (loaded bool, err error) {
	if !f.structField.IsExported() {
		return false, nil
	}
	name := f.EnvName()

	// Allocate a new value to be replaced
	var tmpValue reflect.Value
	if f.isPtr {
		tmpValue = reflect.New(f.value.Type().Elem()).Elem()
	} else {
		tmpValue = reflect.New(f.value.Type()).Elem()
	}

	// Route the field to its appropriate loading function
	if tmpValue.Kind() == reflect.Struct {
		if fileFieldMatcher.MatchString(f.structField.Type.Name()) {
			filePath := os.Getenv(f.FileEnvName())
			if filePath == "" {
				return false, fmt.Errorf("comfy expects %s variable to be set", f.FileEnvName())
			}
			tmpValue.Addr().MethodByName("Setup").Call([]reflect.Value{reflect.ValueOf(f), reflect.ValueOf(filePath)})
			loaded = true
		} else {
			tmpField, err := NewField(f.structField, tmpValue, f.parent)
			if err != nil {
				return false, err
			}
			loaded, err = parseStruct(tmpField)
			if err != nil {
				return false, err
			}
		}
	} else if envValue, exists, err := loadEnvVar(name, name+f.config.fileSuffix); exists || err != nil {
		if err != nil {
			return false, err
		}
		if err := parse(f.structField.Name, envValue, tmpValue); err != nil {
			return false, err
		}
		loaded = true
	}

	if loaded {
		if f.isPtr {
			f.value.Set(tmpValue.Addr())
		} else {
			f.value.Set(tmpValue)
		}
	}
	return loaded, nil
}

// AsEnvPrefix returns a string to be used by this field's children as their prefix.
func (f Field) AsEnvPrefix() string {
	p := ""
	if f.parent != nil {
		p = f.parent.AsEnvPrefix()
	}
	if !f.config.noPrefix {
		envName := f.baseEnvName()
		if p != "" {
			p += "_"
		}
		p += envName
	}
	return p
}

func (f Field) EnvName() string {
	prefix := ""
	if f.parent != nil {
		prefix = f.parent.AsEnvPrefix()
	}
	envName := f.baseEnvName()

	if prefix != "" && envName != "" {
		envName = prefix + "_" + envName
	}

	return envName
}

func (f Field) FileEnvName() string {
	return f.EnvName() + f.config.fileSuffix
}

func (f Field) baseEnvName() string {
	if f.config.envName != nil {
		return *f.config.envName
	}
	return ConvertToEnvName(f.structField.Name)
}

func parseStruct(field Field) (bool, error) {
	loaded := false

	for structField, fieldValue := range field.value.Fields() {
		child, err := NewField(structField, fieldValue, &field)
		if err != nil {
			return false, err
		}
		childLoaded, err := child.Load()
		if err != nil {
			return false, err
		}
		loaded = loaded || childLoaded
	}
	return loaded, nil
}
