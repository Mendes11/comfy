package comfy_test

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/Mendes11/comfy"
)

type MyType struct {
	strVal string
	intVal int
}

func (m *MyType) ComfyParse(v string) error {
	values := strings.Split(v, "@")
	if len(values) != 2 {
		return fmt.Errorf("expected a string split with an @")
	}
	m.strVal = values[0]
	iVal, err := strconv.Atoi(values[1])
	if err != nil {
		return fmt.Errorf("MyType#ComfyParse failed to convert %s to int", values[1])
	}
	m.intVal = iVal
	return nil
}

// MyStringWithParser will truncate text to up to 10 characters
type MyStringWithParser string

func (m *MyStringWithParser) ComfyParse(s string) error {
	*m = MyStringWithParser(s[:10])
	return nil
}

func setupFileVariable(t *testing.T, value string) *os.File {
	testDir := t.TempDir()
	f, err := os.CreateTemp(testDir, "envfile")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = f.Write([]byte(value))
	_ = f.Close()
	return f
}

func TestBaseStructLoad(t *testing.T) {
	t.Run("fails when not a pointer", func(t *testing.T) {
		type Config struct {
			Var1 string
		}
		if err := comfy.Load(Config{}); err == nil {
			t.Fatalf("Expected to fail")
		} else if err != comfy.ErrExpectedPtr {
			t.Fatalf("Expected comfy.NotPointerErr, got %s", err)
		}
	})
	t.Run("Load loads with all variables set", func(t *testing.T) {
		t.Setenv("VAR1", "test_var_1")
		t.Setenv("VAR2", "test_var_2")

		type Config struct {
			Var1 string
			Var2 string
		}
		conf := Config{}
		err := comfy.Load(&conf)
		if err != nil {
			t.Fatal(err)
		}

		if conf.Var1 != "test_var_1" {
			t.Fatalf("expected Var1='test_var_1', got='%s'", conf.Var1)
		}
		if conf.Var2 != "test_var_2" {
			t.Fatalf("expected Var2='test_var_2', got='%s'", conf.Var2)
		}
	})

	t.Run("private attributes aren't loaded", func(t *testing.T) {
		t.Setenv("VAR1", "test_var_1")
		t.Setenv("VAR2", "test_var_2")

		type Config struct {
			var1 string
			Var2 string
		}
		conf := Config{}
		err := comfy.Load(&conf)
		if err != nil {
			t.Fatal(err)
		}

		if conf.var1 != "" {
			t.Fatalf("expected Var1='', got='%s'", conf.var1)
		}
		if conf.Var2 != "test_var_2" {
			t.Fatalf("expected Var2='test_var_2', got='%s'", conf.Var2)
		}
	})

	t.Run("nested structs are loaded", func(t *testing.T) {
		t.Setenv("VAR1", "test_var_1")
		t.Setenv("INNER_VAR1", "inner_var_value")
		type InnerStruct struct {
			Var1 string
		}
		type Config struct {
			Var1  string
			Inner InnerStruct
		}

		conf := Config{}
		err := comfy.Load(&conf)
		if err != nil {
			t.Fatal(err)
		}
		if conf.Var1 != "test_var_1" {
			t.Fatalf("expected Var1='test_var_1', got='%s'", conf.Var1)
		}
		if conf.Inner.Var1 != "inner_var_value" {
			t.Fatalf("expected Inner.Var1='inner_var_value', got='%s'", conf.Inner.Var1)
		}
	})

	t.Run("optional fields support", func(t *testing.T) {
		t.Setenv("OPT_VAR_FILLED", "test_filled")
		t.Setenv("OPT_STRUCT_FILLED_OPT_VAR", "test_inner_filled")
		type InnerStruct struct {
			Var1   string
			OptVar *string
		}
		type Config struct {
			OptVarFilled    *string
			OptVar          *string
			OptStruct       *InnerStruct // Should be left nil if no var is ever set.
			OptStructFilled *InnerStruct
		}
		conf := Config{}
		err := comfy.Load(&conf)
		if err != nil {
			t.Fatal(err)
		}
		if conf.OptVar != nil {
			t.Fatalf("expected OptVar to be nil, got '%s'", *conf.OptVar)
		}
		if conf.OptVarFilled == nil {
			t.Fatalf("expected OptVarFilled to be 'test_filled', got nil")
		} else if *conf.OptVarFilled != "test_filled" {
			t.Fatalf("expected OptVarFilled to be 'test_filled', got '%s'", *conf.OptVarFilled)
		}
		if conf.OptStruct != nil {
			t.Fatalf("expected OptStruct to be nil, got %+x", *conf.OptStruct)
		}
		if conf.OptStructFilled == nil {
			t.Fatalf("expected OptStructFilled to not be nil")
		}
	})

	t.Run("nested has no-prefix configuration support", func(t *testing.T) {
		t.Setenv("VAR2", "test_var_2")
		type InnerStruct struct {
			Var2 string
		}
		type Config struct {
			Inner InnerStruct `comfy:",no-prefix"`
		}
		conf := Config{}
		err := comfy.Load(&conf)
		if err != nil {
			t.Fatal(err)
		}
		if conf.Inner.Var2 != "test_var_2" {
			t.Fatalf("expected Inner.Var2 to be 'test_var_2', got '%s'", conf.Inner.Var2)
		}
	})

	t.Run("env name override supported", func(t *testing.T) {
		t.Setenv("NEW_NAME", "test_var_1")
		t.Setenv("NEW_NAME_INNER_ANOTHER_NAME", "test_var_inner")

		type InnerStruct struct {
			Var1 string `comfy:"ANOTHER_NAME"`
		}
		type Config struct {
			Var1  string      `comfy:"NEW_NAME"`
			Inner InnerStruct `comfy:"NEW_NAME_INNER"`
		}

		conf := Config{}
		err := comfy.Load(&conf)
		if err != nil {
			t.Fatal(err)
		}
		if conf.Var1 != "test_var_1" {
			t.Fatalf("expected conf.Var1 to be 'test_var_1', got '%s'", conf.Var1)
		}
		if conf.Inner.Var1 != "test_var_inner" {
			t.Fatalf("expected conf.Inner.Var1 to be 'test_var_inner', got '%s'", conf.Inner.Var1)
		}
	})

	t.Run("default values kept if not set", func(t *testing.T) {
		t.Setenv("VAR1", "test_var_1")
		type Inner struct {
			Var1 string
			Var2 string
		}
		type Config struct {
			Var1  string
			Var2  string
			Inner *Inner
		}
		conf := Config{
			Var1: "default_value_1",
			Var2: "default_value_2",
			Inner: &Inner{
				Var1: "default_inner_1",
				Var2: "default_inner_2",
			},
		}
		err := comfy.Load(&conf)
		if err != nil {
			t.Fatal(err)
		}
		if conf.Var1 != "test_var_1" {
			t.Fatalf("expected conf.Var1 to be 'test_var_1', got '%s'", conf.Var1)
		}
		if conf.Var2 != "default_value_2" {
			t.Fatalf("expected conf.Var2 to be 'default_value_2', got '%s'", conf.Var2)
		}
		if conf.Inner == nil {
			t.Fatalf("expected conf.Inner to not be nil")
		}
		if conf.Inner.Var1 != "default_inner_1" {
			t.Fatalf("expected conf.Inner.Var1 to be 'default_inner_1', got '%s'", conf.Inner.Var1)
		}
		if conf.Inner.Var2 != "default_inner_2" {
			t.Fatalf("expected conf.Inner.Var2 to be 'default_inner_2', got '%s'", conf.Inner.Var2)
		}
	})
}

func TestAutomaticFileLoading(t *testing.T) {
	t.Run("variables load from a file path by default", func(t *testing.T) {
		type Config struct {
			Var1 string
			Var2 string
		}

		testDir := t.TempDir()
		f, err := os.CreateTemp(testDir, "var1")
		if err != nil {
			t.Fatal(err)
		}
		f.Write([]byte("valueFromFile"))
		f.Close()

		t.Setenv("VAR1", "fromEnv")
		t.Setenv("VAR2_FILE", f.Name())
		conf := &Config{}
		err = comfy.Load(conf)
		if err != nil {
			t.Fatal(err)
		}
		if conf.Var1 != "fromEnv" {
			t.Fatalf("expected Var1 to be 'fromEnv' got '%s'", conf.Var1)
		}
		if conf.Var2 != "valueFromFile" {
			t.Fatalf("expected Var2 to be 'valueFromFile', got '%s'", conf.Var2)
		}
	})
	t.Run("FileField type is correctly setup", func(t *testing.T) {
		type Config struct {
			Var1 comfy.FileField[int]
		}
		f := setupFileVariable(t, "444")
		t.Setenv("VAR1_FILE", f.Name())

		conf := &Config{}
		err := comfy.Load(conf)
		if err != nil {
			t.Fatal(err)
		}
		v, err := conf.Var1.Value()
		if err != nil {
			t.Fatal(err)
		}

		if v != 444 {
			t.Fatalf("expected Var1 to be 444, got %d", v)
		}
	})
}

func TestLoadConfig(t *testing.T) {
	t.Run("LoadConfig without default", func(t *testing.T) {
		type Config struct {
			Var1 string
		}
		t.Setenv("VAR1", "value1")
		conf, err := comfy.LoadConfig[Config]()
		if err != nil {
			t.Fatal(err)
		}
		if conf.Var1 != "value1" {
			t.Fatalf("expected VAR1 to be 'value1', got '%s'", conf.Var1)
		}
	})
	t.Run("LoadConfig with environment default", func(t *testing.T) {
		type Config struct {
			Var1 string
			Var2 string
		}
		t.Setenv("VAR2", "valueInEnv")
		comfy.SetEnvironmentDefault("development", Config{Var1: "defaultValue1"})
		comfy.SetAppEnvironment("development")
		conf, err := comfy.LoadConfig[Config]()
		if err != nil {
			t.Fatal(err)
		}
		if conf.Var1 != "defaultValue1" {
			t.Fatalf("expected Var1 to be 'defaultValue1', got '%s'", conf.Var1)
		}
		if conf.Var2 != "valueInEnv" {
			t.Fatalf("expected Var2 to be 'valueInEnv', got '%s'", conf.Var2)
		}
	})
}

func TestLoadWithCustomParser(t *testing.T) {
	t.Run("Load with custom parser correctly parsed", func(t *testing.T) {
		type Config struct {
			Var1 MyType
		}
		t.Setenv("VAR1", "Test@2")
		conf := &Config{}
		err := comfy.Load(conf)
		if err != nil {
			t.Fatal(err)
		}
		if conf.Var1.strVal != "Test" {
			t.Fatalf("expected strval to be 'Test', got '%s'", conf.Var1.strVal)
		}
		if conf.Var1.intVal != 2 {
			t.Fatalf("expected intVal to be 2, got %d", conf.Var1.intVal)
		}
	})

	t.Run("Load user-defined primitive type correctly parses", func(t *testing.T) {
		type MyString string
		type Config struct {
			Var1 MyString
		}
		t.Setenv("VAR1", "Testing")
		conf := &Config{}
		err := comfy.Load(conf)
		if err != nil {
			t.Fatal(err)
		}
		if conf.Var1 != "Testing" {
			t.Fatalf("expected Var1 to be 'Testing', got '%s'", conf.Var1)
		}
	})

	t.Run("Load user-defined primitive with custom parser", func(t *testing.T) {
		type Config struct {
			Var1 MyStringWithParser
		}
		t.Setenv("VAR1", "Some very long text to be truncated")
		conf := &Config{}
		err := comfy.Load(conf)
		if err != nil {
			t.Fatal(err)
		}
		if len(conf.Var1) != 10 {
			t.Fatalf("expected Var1 to be truncated with 10 characters, got %d", len(conf.Var1))
		}
		if conf.Var1 != "Some very " {
			t.Fatalf("expected Var1 to be 'Some very l', got '%s'", conf.Var1)
		}
	})
}
