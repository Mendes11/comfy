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
