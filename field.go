package comfy

import (
	"reflect"
)

// Field represents a single configurable field loaded by Comfy
type Field struct {
	structField reflect.StructField
	value       reflect.Value
	isPtr       bool

	Parent *Field
	Config FieldConfig
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
		Parent:      parent,
		Config:      conf,
	}, nil
}

// Load will load this field from the configuration source, and returns true when the configuration exists and was loaded.
// When no error, but also no configuration matched for this field, it returns false.
func (f Field) Load() (loaded bool, err error) {
	if !f.structField.IsExported() {
		return false, nil
	}
	name := f.EnvName()

	// Allocate a new value to be replaced only if the field is loaded
	var tmpValue reflect.Value
	if f.isPtr {
		tmpValue = reflect.New(f.value.Type().Elem()).Elem()
	} else {
		tmpValue = reflect.New(f.value.Type()).Elem()
	}

	// Route the field to its appropriate loading function
	if tmpValue.Kind() == reflect.Struct {
		tmpField, err := NewField(f.structField, tmpValue, f.Parent)
		if err != nil {
			return false, err
		}
		loaded, err = parseStruct(tmpField)
		if err != nil {
			return false, err
		}

	} else if envValue, exists, err := loadEnvVar(name, true); exists || err != nil {
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
	if f.Parent != nil {
		p = f.Parent.AsEnvPrefix()
	}
	if !f.Config.noPrefix {
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
	if f.Parent != nil {
		prefix = f.Parent.AsEnvPrefix()
	}
	envName := f.baseEnvName()

	if prefix != "" && envName != "" {
		envName = prefix + "_" + envName
	}

	return envName
}

func (f Field) baseEnvName() string {
	if f.Config.envName != nil {
		return *f.Config.envName
	}
	return convertToEnvName(f.structField.Name)
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
