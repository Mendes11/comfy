package comfy

import (
	"os"
	"strings"
	"unicode"
)

// convertToEnvName uses Comfy's convention to transform field names into env var names.
// It essentially converts cammelCase into upper case'd SNAKE_CASE, with extra rules:
//
// 1. When an upper case follows a lower case, split that into two -> MyVariable -> MY_VARIABLE
// 2. When sequence of upper case letters, it splits by the last item in the sequence. eg: DBName -> DB_NAME
func ConvertToEnvName(name string) string {
	nameRunes := []rune(name)
	start := 0
	parts := make([]string, 0)
	for i, c := range name {
		// Splitting boundary 1: change from lower to Upper case.
		// eg: testName -> [test, Name]
		if i > 0 && unicode.IsUpper(c) && unicode.IsLower(nameRunes[i-1]) {
			parts = append(parts, name[start:i])
			start = i
		}
		// Splitting boundary 2: Sequence of upper cases, followed by an Upper case + lower case
		if (i-start) >= 2 && unicode.IsLower(c) && unicode.IsUpper(nameRunes[i-1]) && unicode.IsUpper(nameRunes[i-2]) {
			parts = append(parts, name[start:i-1])
			start = i - 1
		}
	}
	parts = append(parts, name[start:])
	return strings.ToUpper(strings.Join(parts, "_"))
}

// loadEnvVar will try to load the env var by name or by its _FILE suffix.
func loadEnvVar(name string, fileSuffix string) (val string, loaded bool, err error) {
	val, loaded = os.LookupEnv(name)

	if !loaded && fileSuffix != "" {
		var filePath string
		filePath, loaded = os.LookupEnv(name + fileSuffix)
		if loaded {
			valueBytes, err := os.ReadFile(filePath)
			if err != nil {
				return "", false, err
			}
			val = string(valueBytes)
		}
	}

	return val, loaded, nil
}
