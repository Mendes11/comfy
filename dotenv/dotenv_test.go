package dotenv_test

import (
	"os"
	"path"
	"testing"

	"github.com/Mendes11/comfy/dotenv"
)

func TestDotEnvLoader(t *testing.T) {
	t.Run("no file present won't fail", func(t *testing.T) {
		err := dotenv.LoadDotEnvs(nil)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
	})
	t.Run("loads `.env` file", func(t *testing.T) {
		tempDir := t.TempDir()
		env := `
TEST_ENV=value
# TEST_ENV=value_commented
TEST_ENV2=value2
// TEST_ENV2=value2_commented
		`
		writeEnv(t, tempDir, ".env", env)
		if err := dotenv.SetEnvsDir(tempDir); err != nil {
			t.Fatal(err)
		}
		if err := dotenv.LoadDotEnvs(nil); err != nil {
			t.Fatal(err)
		}

		assertEnv(t, "TEST_ENV", "value")
		assertEnv(t, "TEST_ENV2", "value2")
	})

	t.Run("loads .env and .env.local overrides", func(t *testing.T) {
		tempDir := t.TempDir()
		env := `
TEST_ENV=value
TEST_ENV2=value2
		`
		envLocal := `
TEST_ENV2=localValue
TEST_ENV3=value3
		`
		writeEnv(t, tempDir, ".env", env)
		writeEnv(t, tempDir, ".env.local", envLocal)
		if err := dotenv.SetEnvsDir(tempDir); err != nil {
			t.Fatal(err)
		}
		if err := dotenv.LoadDotEnvs(nil); err != nil {
			t.Fatal(err)
		}

		assertEnv(t, "TEST_ENV", "value")
		assertEnv(t, "TEST_ENV2", "localValue")
		assertEnv(t, "TEST_ENV3", "value3")
	})

	t.Run("loads environment-specific files", func(t *testing.T) {
		tempDir := t.TempDir()
		env := `
TEST_ENV=value
TEST_ENV2=value2
		`
		envDev := `
TEST_ENV2=devValue2
TEST_ENV3=devValue3
`
		envDevLocal := `
TEST_ENV3=devValueLocal3
TEST_ENV4=devValueLocal4
		`

		writeEnv(t, tempDir, ".env", env)
		writeEnv(t, tempDir, ".env.development", envDev)
		writeEnv(t, tempDir, ".env.development.local", envDevLocal)
		if err := dotenv.SetEnvsDir(tempDir); err != nil {
			t.Fatal(err)
		}
		if err := dotenv.LoadDotEnvs(func() string { return "development" }); err != nil {
			t.Fatal(err)
		}

		assertEnv(t, "TEST_ENV", "value")
		assertEnv(t, "TEST_ENV2", "devValue2")
		assertEnv(t, "TEST_ENV3", "devValueLocal3")
		assertEnv(t, "TEST_ENV4", "devValueLocal4")
	})

	t.Run("loads environment from dotenv file", func(t *testing.T) {
		tempDir := t.TempDir()
		env := `
APP_ENV=development
		`
		envDev := `
TEST_ENV=devValue
`
		writeEnv(t, tempDir, ".env", env)
		writeEnv(t, tempDir, ".env.development", envDev)

		if err := dotenv.SetEnvsDir(tempDir); err != nil {
			t.Fatal(err)
		}
		if err := dotenv.LoadDotEnvs(func() string { return os.Getenv("APP_ENV") }); err != nil {
			t.Fatal(err)
		}

		assertEnv(t, "APP_ENV", "development")
		assertEnv(t, "TEST_ENV", "devValue")
	})
}

func assertEnv(t *testing.T, name, expected string) {
	if val, exists := os.LookupEnv(name); !exists {
		t.Fatalf("expected %s to be set", name)
	} else if val != expected {
		t.Fatalf("expected %s to be '%s', got '%s'", name, expected, val)
	}
}

func writeEnv(t *testing.T, dir string, name string, content string) {
	f, err := os.OpenFile(path.Join(dir, name), os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0o666)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	_, err = f.Write([]byte(content))
	if err != nil {
		t.Fatal(err)
	}
}
