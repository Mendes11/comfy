package comfy

import (
	"fmt"
	"reflect"

	"github.com/Mendes11/comfy/dotenv"
)

type Comfy[T any] struct {
	Config *T `comfy:",no-prefix"`
}

// New instance of a Comfy loader
func New[T any](conf *T) *Comfy[T] {
	return &Comfy[T]{Config: conf}
}

// Load the configuration values
func (c *Comfy[T]) Load() error {
	if err := setup(); err != nil {
		return fmt.Errorf("comfy.Load setup failed: %w", err)
	}
	confValue := reflect.ValueOf(c.Config)
	if confValue.Kind() != reflect.Pointer {
		return ErrExpectedPtr
	}
	if confValue.Elem().Kind() != reflect.Struct {
		return ErrExpectedStructType
	}

	f, err := NewField("Config", confValue.Elem(), FieldConfig{noPrefix: true}, nil)
	if err != nil {
		return fmt.Errorf("comfy.Comfy#Load failed: %w", err)
	}
	_, err = f.Load()
	if err != nil {
		return fmt.Errorf("comfy.Comfy#Load failed to load field: %w", err)
	}
	return nil
}

// Load the values into the given configuration variable.
// conf must be a pointer
func Load[T any](conf *T) error {
	return New(conf).Load()
}

// LoadConfig will return a new configuration object based on the type parameter T.
// If an environment default is set, it'll initialize the values based on it.
func LoadConfig[T any]() (T, error) {
	conf := *new(T)
	if envDefault, exists := environments[AppEnvironment()]; exists {
		conf = envDefault.(T)
	}
	c := New(&conf)
	err := c.Load()
	return *c.Config, err
}

func setup() error {
	if err := dotenv.LoadDotEnvs(AppEnvironment); err != nil {
		return err
	}
	return nil
}
