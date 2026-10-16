package comfy_test

import (
	"os"
	"testing"

	"github.com/Mendes11/comfy"
)

func TestBaseStructLoad(t *testing.T) {
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
		if conf.OptVarFilled == nil || *conf.OptVarFilled != "test_filled" {
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
		t.Log(f.Name())
		if conf.Var1 != "fromEnv" {
			t.Fatalf("expected Var1 to be 'fromEnv' got '%s'", conf.Var1)
		}
		if conf.Var2 != "valueFromFile" {
			t.Fatalf("expected Var2 to be 'valueFromFile', got '%s'", conf.Var2)
		}
	})
}
