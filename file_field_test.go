package comfy

import (
	"os"
	"reflect"
	"testing"
	"time"
)

func testSetupFileVariable(t *testing.T, value string) *os.File {
	testDir := t.TempDir()
	f, err := os.CreateTemp(testDir, "envfile")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = f.Write([]byte(value))
	_ = f.Close()
	return f
}

func testSetupChangeFileVariable(path string, value string) error {
	f, err := os.OpenFile(path, os.O_TRUNC|os.O_WRONLY, 0o666)
	if err != nil {
		return err
	}
	_, _ = f.Write([]byte(value))
	_ = f.Close()
	return nil
}

func TestFileField(t *testing.T) {
	t.Run("fileField loads when file changes", func(t *testing.T) {
		f := testSetupFileVariable(t, "testFromFile")
		fileField := NewFileField[string](
			Field{
				structField: reflect.StructField{Name: "test"},
			},
			f.Name(),
		)
		valBefore, err := fileField.Value()
		if err != nil {
			t.Fatal(err)
		}
		if err := testSetupChangeFileVariable(f.Name(), "newValue"); err != nil {
			t.Fatal(err)
		}
		newValue, err := fileField.Value()
		if err != nil {
			t.Fatal(err)
		}

		if valBefore != "testFromFile" {
			t.Fatalf("expected first read to be 'testFromFile', got '%s'", valBefore)
		}
		if newValue != "newValue" {
			t.Fatalf("expected second read to be 'newValue', got '%s'", newValue)
		}
	})

	t.Run("fileField loads and parse from path", func(t *testing.T) {
		f := testSetupFileVariable(t, "1234")
		fileField := NewFileField[int](
			Field{
				structField: reflect.StructField{Name: "test"},
			},
			f.Name(),
		)
		val, err := fileField.Value()
		if err != nil {
			t.Fatal(err)
		}
		if val != 1234 {
			t.Fatalf("expected 1234, got %d", val)
		}
	})
	t.Run("fileField subscription to change", func(t *testing.T) {
		f := testSetupFileVariable(t, "1234")
		fileField := NewFileField[int](
			Field{
				structField: reflect.StructField{Name: "test"},
			},
			f.Name(),
		)

		oldValue, newValue := 0, 0
		err := fileField.SubscribeToChange("test", func(id string, old int, newV int) {
			oldValue = old
			newValue = newV
		})
		if err != nil {
			t.Fatal(err)
		}
		if err := testSetupChangeFileVariable(f.Name(), "4321"); err != nil {
			t.Fatal(err)
		}
		time.Sleep(2 * time.Second)
		if oldValue != 1234 || newValue != 4321 {
			t.Fatalf("expected oldValue to be 1234, got %d. newValue to be 4321, got %d", oldValue, newValue)
		}
	})
}
