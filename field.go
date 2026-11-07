package comfy

import (
	"fmt"
	"os"
	"reflect"
	"regexp"
)

var fileFieldMatcher = regexp.MustCompile(`^FileField\[.*?\]$`)

// NewField returns an instance of a Comfy Field
func NewField(name string, value reflect.Value, conf FieldConfig, parent *Field) (Field, error) {
	if !(value.CanSet() || (value.Kind() == reflect.Pointer && value.Elem().CanSet())) {
		return Field{}, fmt.Errorf("Value of %s must be settable", fieldPath(name, parent))
	}
	return Field{
		name:   name,
		value:  value,
		isPtr:  value.Kind() == reflect.Pointer,
		parent: parent,
		config: conf,
	}, nil
}

// Field represents a single configurable field loaded by Comfy
type Field struct {
	name  string
	value reflect.Value
	isPtr bool

	parent *Field
	// When this is of type Struct, we store all its fields here.
	children []Field
	config   FieldConfig
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

func (f *Field) addChildren(children Field) {
	f.children = append(f.children, children)
}

// Load will load this field from the configuration source, and returns true when the configuration exists and was loaded.
// When no error, but also no configuration matched for this field, it returns false.
func (f Field) Load() (loaded bool, err error) {
	// Load the fields bottom to top
	for _, c := range f.children {
		loaded, err = c.Load()
		if err != nil {
			return loaded, fmt.Errorf("%w", err)
		}
	}
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
	// Check if is of FileField type
	if fileFieldMatcher.MatchString(f.value.Type().Name()) {
		// TODO: Can this become a parser? Now that I have the regexp registry? Or even ComfyParser interface, instead of Setup?
		filePath := os.Getenv(f.FileEnvName())
		if filePath == "" {
			return false, fmt.Errorf("comfy expects %s variable to be set", f.FileEnvName())
		}
		copyValue.Addr().MethodByName("Setup").Call([]reflect.Value{reflect.ValueOf(f), reflect.ValueOf(filePath)})
		loaded = true

	} else if envValue, exists, err := loadEnvVar(name, name+f.config.fileSuffix); exists || err != nil {
		if err != nil {
			return false, err
		}
		if err := parse(f.name, envValue, copyValue, f.config); err != nil {
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

func isStruct(value reflect.Value) bool {
	return value.Kind() == reflect.Struct || (value.Kind() == reflect.Pointer && value.Elem().Kind() == reflect.Struct)
}

func fieldPath(name string, parent *Field) string {
	path := ""
	if parent != nil {
		path = fieldPath(parent.name, parent.parent)
	}
	if path != "" {
		return fmt.Sprintf("%s.%s", path, name)
	}
	return name
}

// newFieldsGraph loads all fields and their nested fields from a root struct value
// value has to be a reflection of a Struct field or of a pointer of one.
// It'll exclude from loading as a field:
// - Private fields of a Field instance of a struct
// - Attributes of a Field that implements the ComfyParser interface
func newFieldsGraph(name string, value reflect.Value, fConf FieldConfig, parent *Field) (field Field, err error) {
	field, err = NewField(name, value, fConf, parent)
	if err != nil {
		return field, fmt.Errorf("comfy.newFieldsGraph failed to mount field for %s: %w", name, err)
	}
	// Nil values won't resolve as structs. Their .Kind() resolves to Invalid.
	// So I allocate a value to be able to resolve the struct fields in case of a Struct
	// and be able to go down the children and load them as well
	if value.Kind() == reflect.Pointer && value.IsNil() {
		value = reflect.New(value.Type().Elem()).Elem()
	}
	if isStruct(value) && !implementsComfyParser(value) {
		if value.Kind() == reflect.Pointer {
			value = value.Elem()
		}
		for structField, childValue := range value.Fields() {
			if !structField.IsExported() {
				continue
			}
			conf, err := fieldTagConfigs(structField)
			if err != nil {
				return field, fmt.Errorf("comfy.newFieldsGraph failed to load tags for %s.%s: %w", name, structField.Name, err)
			}
			child, err := newFieldsGraph(structField.Name, childValue, conf, &field)
			if err != nil {
				return field, err
			}
			field.addChildren(child)
		}
	}

	return field, nil
}
