package comfy_test

import (
	"reflect"
	"testing"

	"github.com/Mendes11/comfy"
)

type ComfyField1 struct {
	name   string
	conf   comfy.FieldConfig
	called bool
}

func (c *ComfyField1) ComfySetup(name string, fc comfy.FieldConfig) error {
	c.called = true
	c.name = name
	c.conf = fc
	return nil
}

func TestComfyFieldSetup(t *testing.T) {
	t.Run("setup is called on field initialization", func(t *testing.T) {
		obj := &ComfyField1{}
		fc := comfy.NewFieldConfig()
		_, err := comfy.NewField("test", reflect.ValueOf(obj), fc, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !obj.called {
			t.Fatalf("Expected ComfyField#ComfySetup to be called")
		}
		if obj.name != "test" {
			t.Fatalf("Expected obj.name to be 'test', got %s", obj.name)
		}
	})
}

func TestFieldTags(t *testing.T) {
	t.Run("Field.Load returns error when field is required", func(t *testing.T) {
		type ConfigWithRequired struct {
			Var string `comfy:",required"`
		}
		_, err := comfy.LoadConfig[ConfigWithRequired]()
		if err == nil {
			t.Fatalf("expected comfy to fail")
		}
		t.Setenv("VAR", "test")
		_, err = comfy.LoadConfig[ConfigWithRequired]()
		if err != nil {
			t.Fatalf("expected to not fail, got: %v", err)
		}
	})
	t.Run("required field for optional fields", func(t *testing.T) {
		type Config struct {
			Var *string `comfy:",required"`
		}
		_, err := comfy.LoadConfig[Config]()
		if err == nil {
			t.Fatal("expected comfy to fail")
		}
		t.Setenv("VAR", "test")
		_, err = comfy.LoadConfig[Config]()
		if err != nil {
			t.Fatalf("expected to not fail, got %v", err)
		}
	})
	t.Run("required field doesn't fail if default value exists", func(t *testing.T) {
		type Config struct {
			Var1 string  `comfy:",required"`
			Var2 *string `comfy:",required"`
		}
		comfy.SetAppEnvironment("development")
		t.Cleanup(func() {
			comfy.UnsetEnvironmentDefault("testing")
		})
		comfy.SetEnvironmentDefault("development", Config{})
		_, err := comfy.LoadConfig[Config]()
		if err == nil {
			t.Fatal("expected to fail")
		}
		comfy.SetEnvironmentDefault("development", Config{Var1: "test"})
		_, err = comfy.LoadConfig[Config]()
		if err == nil {
			t.Fatal("expected to fail")
		}
		str := "testing"
		comfy.SetEnvironmentDefault("development", Config{Var1: "test", Var2: &str})
		_, err = comfy.LoadConfig[Config]()
		if err != nil {
			t.Fatalf("expected to not fail, got %v", err)
		}
	})
}
