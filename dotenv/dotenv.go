package dotenv

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"
)

type AppEnvLoader func() string

var (
	envsDir     = ""
	envsMatcher = regexp.MustCompile(`^\s*([A-Za-z_][A-Za-z0-9_]*)\s*=\s*(.*)$`)
)

func SetEnvsDir(dir string) error {
	if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
		return err
	}
	envsDir = dir
	return nil
}

func GetEnvsDir() string {
	if envsDir == "" {
		d, err := os.Getwd()
		if err != nil {
			return ""
		}
		envsDir = d
	}
	return envsDir
}

// LoadDotEnvs reads the .env files into the app environment
// If appEnvLoader is present, it also loads dotenv files
// for the returned environment name. It is only called
// after `.env`, and `.env.local` have been loaded.
//
// It reads files in the following order:
// 1. .env
// 2. .env.local
// 3. .env.<environment> # In case the app has an environment set
// 4. .env.<environment>.local # In case the app has an environment set
func LoadDotEnvs(appEnvLoader AppEnvLoader) error {
	for _, filename := range []string{".env", ".env.local"} {
		filepath := path.Join(GetEnvsDir(), filename)
		if !fileExists(filepath) {
			continue
		}
		if err := loadEnvFromFile(filepath); err != nil {
			return err
		}
	}
	if appEnvLoader != nil {
		env := appEnvLoader()
		if env != "" {
			for _, filename := range []string{".env.{env}", ".env.{env}.local"} {
				resolvedName := strings.Replace(filename, "{env}", env, 1)
				filepath := path.Join(GetEnvsDir(), resolvedName)
				if !fileExists(filepath) {
					continue
				}
				if err := loadEnvFromFile(filepath); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func loadEnvFromFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to load env from file: %w", err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		match := envsMatcher.FindStringSubmatch(line)
		if len(match) == 0 {
			continue
		}
		fmt.Println(match)
		if err = os.Setenv(match[1], match[2]); err != nil {
			return fmt.Errorf("failed to set %s from file %s: %w", match[1], path, err)
		}
	}
	if scanner.Err() != nil {
		return fmt.Errorf("failed to read file %s: %w", path, scanner.Err())
	}
	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}
