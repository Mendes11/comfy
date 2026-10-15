package comfy

import (
	"testing"
)

func TestNamingConvention(t *testing.T) {
	t.Run("cammelCase converts to snake_case", func(t *testing.T) {
		v := ConvertToEnvName("TestVar")
		if v != "TEST_VAR" {
			t.Fatalf("expected 'TestVar' to become 'TEST_VAR', got %s", v)
		}
	})

	t.Run("upper case sequence is not splitted", func(t *testing.T) {
		v := ConvertToEnvName("TestVAR")
		if v != "TEST_VAR" {
			t.Fatalf("expected 'TestVAR' to become 'TEST_VAR', got %s", v)
		}

		v = ConvertToEnvName("TESTVar")
		if v != "TEST_VAR" {
			t.Fatalf("expected 'TESTVar' to become 'TEST_VAR', got %s", v)
		}

		v = ConvertToEnvName("DB")
		if v != "DB" {
			t.Fatalf("expected 'DB' to become 'DB', got %s", v)
		}

		v = ConvertToEnvName("DBName")
		if v != "DB_NAME" {
			t.Fatalf("expected 'DBName' to become 'DB_NAME', got %s", v)
		}

		v = ConvertToEnvName("TestID")
		if v != "TEST_ID" {
			t.Fatalf("expected 'TestID' to become 'TEST_ID', got %s", v)
		}
		v = ConvertToEnvName("QTest")
		if v != "Q_TEST" {
			t.Fatalf("expected 'QTest' to become 'Q_TEST', got %s", v)
		}
	})
}
