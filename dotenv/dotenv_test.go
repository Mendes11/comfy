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

	t.Run("dotenv files priority respected", func(t *testing.T) {
		//  Process variables wins over all
		//  Local always wins over the other files
		// | Priority |  Source |
		// | +---+    | +---+ |
		// | 1        |  processEnv
		// | 2        |  .env.<app_env>.local
		// | 3        |  .env.local
		// | 4        |  .env.<app_env>
		// | 5        |  .env
		tempDir := t.TempDir()
		env := `
TEST_ENV=envValue
TEST_ENV2=envValue2
TEST_ENV3=envValue3
TEST_ENV4=envValue4
TEST_ENV5=envValue5
`
		envDev := `
TEST_ENV=devValue
TEST_ENV2=devValue2
TEST_ENV3=devValue3
TEST_ENV4=devValue4
`
		envLocal := `
TEST_ENV=envLocalValue
TEST_ENV2=envLocalValue2
TEST_ENV3=envLocalValue3
`
		envDevLocal := `
TEST_ENV=devValueLocal
TEST_ENV2=devValueLocal2
		`

		t.Setenv("TEST_ENV", "processEnvValue")
		writeEnv(t, tempDir, ".env", env)
		writeEnv(t, tempDir, ".env.local", envLocal)
		writeEnv(t, tempDir, ".env.development", envDev)
		writeEnv(t, tempDir, ".env.development.local", envDevLocal)

		if err := dotenv.SetEnvsDir(tempDir); err != nil {
			t.Fatal(err)
		}
		if err := dotenv.LoadDotEnvs(func() string { return "development" }); err != nil {
			t.Fatal(err)
		}

		assertEnv(t, "TEST_ENV", "processEnvValue")
		assertEnv(t, "TEST_ENV2", "devValueLocal2")
		assertEnv(t, "TEST_ENV3", "envLocalValue3")
		assertEnv(t, "TEST_ENV4", "devValue4")
		assertEnv(t, "TEST_ENV5", "envValue5")
	})
}

func assertEnv(t *testing.T, name, expected string) {
	t.Helper()
	if val, exists := os.LookupEnv(name); !exists {
		t.Fatalf("expected %s to be set", name)
	} else if val != expected {
		t.Fatalf("expected %s to be '%s', got '%s'", name, expected, val)
	}
}

func writeEnv(t *testing.T, dir string, name string, content string) {
	t.Helper()
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
