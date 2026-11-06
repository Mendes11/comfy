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
	name  string
	value reflect.Value
	isPtr bool

	parent *Field
	config FieldConfig
}

func (f Field) SetValue(v reflect.Value) {
	if v.Type() != f.value.Type() {
		v = v.Convert(f.value.Type())
	}
	f.value.Set(v)
}

func (f Field) Config() FieldConfig {
	return f.config
}

// NewField returns an instance of a Comfy Field
func NewField(name string, value reflect.Value, conf FieldConfig, parent *Field) (Field, error) {
	return Field{
		name:   name,
		value:  value,
		isPtr:  value.Kind() == reflect.Pointer,
		parent: parent,
		config: conf,
	}, nil
}

func NewFieldFromStructField(sField reflect.StructField, value reflect.Value, parent *Field) (Field, error) {
	conf, err := fieldTagConfigs(sField)
	if err != nil {
		return Field{}, fmt.Errorf("comfy.NewFieldFromStructField fieldTagConfigs: %w", err)
	}
	return Field{
		name:   sField.Name,
		value:  value,
		isPtr:  value.Kind() == reflect.Pointer,
		parent: parent,
		config: conf,
	}, nil
}

// Load will load this field from the configuration source, and returns true when the configuration exists and was loaded.
// When no error, but also no configuration matched for this field, it returns false.
func (f Field) Load() (loaded bool, err error) {
	name := f.EnvName()

	// Create a copy value to serve as the parsing target
	// This temporary value is used to make sure optional fields can keep their nil value when nothing loaded,
	// because we have to have an allocated value when parsing the field and assigning a value to it.
	// But, if there are no fields loaded in that branch, we don't assign this copy to the original
	copyValue := f.value
	if f.isPtr && f.value.IsNil() {
		copyValue = reflect.New(f.value.Type().Elem()).Elem()
	} else if f.isPtr {
		copyValue = copyValue.Elem()
	}

	// Route the field to its appropriate loading function
	if copyValue.Kind() == reflect.Struct && !copyValue.Addr().Type().Implements(parserType) {
		// Struct Type
		// Check if is of FileField type
		if fileFieldMatcher.MatchString(f.value.Type().Name()) {
			filePath := os.Getenv(f.FileEnvName())
			if filePath == "" {
				return false, fmt.Errorf("comfy expects %s variable to be set", f.FileEnvName())
			}
			copyValue.Addr().MethodByName("Setup").Call([]reflect.Value{reflect.ValueOf(f), reflect.ValueOf(filePath)})
			loaded = true
		} else {
			// Load the struct fields
			tmpField, err := NewField(f.name, copyValue, f.config, f.parent)
			if err != nil {
				return false, err
			}
			loaded, err = loadStructFields(tmpField)
			if err != nil {
				return false, err
			}
		}
	} else if envValue, exists, err := loadEnvVar(name, name+f.config.fileSuffix); exists || err != nil {
		if err != nil {
			return false, err
		}
		if err := parse(f.name, envValue, copyValue); err != nil {
			return false, err
		}
		loaded = true
	}

	if loaded {
		if f.isPtr {
			f.value.Set(copyValue.Addr())
		} else {
			f.value.Set(copyValue)
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

// EnvName returns the environment name this field will look for.
// When a custom name is set, it will use that name without any prefix
func (f Field) EnvName() string {
	if f.config.HasCustomEnvName() {
		return *f.config.customEnvName
	}
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
	if f.config.customEnvName != nil {
		return *f.config.customEnvName
	}
	return ConvertToEnvName(f.name)
}
