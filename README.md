<p align="center">
    <picture>
      <source media="(prefers-color-scheme: light)" srcset="docs/content/assets/img/comfy.logo.png">
      <img alt="Comfy" title="Comfy" height="400" width="400" src="docs/content/assets/img/comfy.logo.png">
    </picture>
</p>

<p align="center">
  Comfortable, opinionated configuration management for Go applications.
</p>

<p align="center">
  <a href="https://pkg.go.dev/github.com/Mendes11/comfy"><img src="https://pkg.go.dev/badge/github.com/Mendes11/comfy.svg" alt="Go Reference"></a>
</p>

# About

Comfy takes care of the annoying parts of application configuration. You declare a struct, and Comfy fills it.

- **No setup.** Field names become environment variable names: `APIKey` reads `API_KEY`.
- **Many sources.** Values come from environment variables, `.env` files, and secret files.
- **Secrets from files.** Every field can also load from a file with `<NAME>_FILE`. This works with Docker and Kubernetes secrets.
- **Nested structs.** `DB.Host` reads `DB_HOST`.
- **Optional fields.** Pointer fields stay `nil` when no value is set.
- **Per-environment defaults.** Set different defaults for `development`, `production`, and so on.
- **Live values.** `FileField` reads a file each time you ask, and can tell you when the file changes.

# Contents

- [Install](#install)
- [Quick start](#quick-start)
- [The basics](#the-basics)
  - [Naming conventions](#naming-conventions)
  - [Where values come from](#where-values-come-from)
  - [Supported types](#supported-types)
  - [Default values](#default-values)
  - [Optional fields](#optional-fields)
  - [Nested structs](#nested-structs)
  - [Secrets from files](#secrets-from-files)
- [Customization](#customization)
  - [The `comfy` tag](#the-comfy-tag)
  - [Custom types](#custom-types)
  - [Values that change at runtime: `FileField`](#values-that-change-at-runtime-filefield)
  - [Dotenv files](#dotenv-files)
  - [Application environments](#application-environments)
  - [Environment defaults](#environment-defaults)
- [Errors](#errors)
- [API overview](#api-overview)

# Install

```sh
go get github.com/Mendes11/comfy
```

Comfy needs Go 1.26 or newer.

# Quick start

**1. Declare your configuration struct.**

```go
package main

import (
	"fmt"
	"log"

	"github.com/Mendes11/comfy"
)

type Config struct {
	Host   string   // HOST
	Port   int      // PORT
	APIKey string   // API_KEY
	DB     DBConfig // Nested struct: its fields get the DB_ prefix
}

type DBConfig struct {
	Host     string // DB_HOST
	Port     int    // DB_PORT
	User     string // DB_USER
	Password string // DB_PASSWORD
	Name     string // DB_NAME
}

func main() {
	conf, err := comfy.LoadConfig[Config]()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("listening on %s:%d\n", conf.Host, conf.Port)
}
```

**2. Give it values.** Use environment variables, a `.env` file, or both:

```sh
# .env
HOST=0.0.0.0
PORT=8080
API_KEY=dev-key
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_NAME=myapp
```

**3. Run it.** Here, the database password comes from a secret file:

```sh
DB_PASSWORD_FILE=/run/secrets/db_password go run .
```

That's it. You don't need tags or setup code. The sections below explain the rules, and how to change them when you need to.

# The basics

## Naming conventions

Comfy turns each struct field name into an environment variable name:

1. Words are split where a lower case letter is followed by an upper case letter: `MyKey` → `MY_KEY`.
2. A group of upper case letters stays together. The last letter starts a new word if a lower case letter follows it: `DBName` → `DB_NAME`.
3. All letters become upper case, and words are joined with `_`.
4. Fields of a nested struct get the struct field name as a prefix: `DB.Host` → `DB_HOST`.

| Field          | Environment variable |
| :------------- | :------------------- |
| `Host`         | `HOST`               |
| `APIKey`       | `API_KEY`            |
| `UserID`       | `USER_ID`            |
| `HTTPPort`     | `HTTP_PORT`          |
| `MaxOpenConns` | `MAX_OPEN_CONNS`     |
| `DB.Name`      | `DB_NAME`            |

Digits do not split words: `S3Bucket` → `S3BUCKET`. If you want a different name, use the [`comfy` tag](#custom-environment-variable-name).

You can call `comfy.ConvertToEnvName("MyField")` to see the name Comfy will use.

Comfy ignores unexported (lower case) fields.

## Where values come from

When you load a configuration, Comfy does these steps:

1. It reads the [dotenv files](#dotenv-files) and adds their values to the process environment. Variables that are already set are not replaced.
2. It starts with a base value:
   - `LoadConfig[T]()` starts from the [environment default](#environment-defaults), if you registered one. If not, it starts from the zero value of `T`.
   - `Load(&conf)` starts from the values already in `conf`.
3. For each field, it looks for a value in this order:
   1. The environment variable, for example `DB_PASSWORD`.
   2. The file at the path in `<NAME>_FILE`, for example `DB_PASSWORD_FILE`.
   3. If neither is set, the field keeps its base value.

## Supported types

| Type                                             | Example value |
| :----------------------------------------------- | :------------ |
| `string`                                         | `hello`       |
| `int`, `int8`, `int16`, `int32`, `int64`         | `-42`         |
| `uint`, `uint8`, `uint16`, `uint32`, `uint64`    | `42`          |
| A pointer to a supported type, for example `*int` | See [Optional fields](#optional-fields) |
| A struct, or a pointer to a struct               | See [Nested structs](#nested-structs) |
| `comfy.FileField[T]`                             | See [`FileField`](#values-that-change-at-runtime-filefield) |
| Any type that implements `comfy.ComfyParser`     | See [Custom types](#custom-types) |

If a field has another type, and a value is set for it, loading fails with a [`MissingParserError`](#errors).

## Default values

`comfy.Load` fills a struct that you already have. Fields that have no value in the environment keep the value you gave them. Use this to set defaults in code:

```go
conf := Config{
	Host: "localhost",
	Port: 8080,
}
if err := comfy.Load(&conf); err != nil {
	log.Fatal(err)
}
// If PORT is not set, conf.Port is still 8080.
```

`comfy.Load` needs a pointer to a struct. If you give it something else, it returns `comfy.ErrExpectedPtr` or `comfy.ErrExpectedStructType`.

To set different defaults for each environment, see [Environment defaults](#environment-defaults).

## Optional fields

Use a pointer when you need to know if a value was set. The pointer stays `nil` when there is no value for it.

```go
type Config struct {
	SentryDSN *string // nil when SENTRY_DSN is not set
}
```

This also works for nested structs. A `*struct` field stays `nil` when none of its fields has a value.

## Nested structs

Nested structs are loaded field by field. Each level adds its own prefix:

```go
type Config struct {
	Cache CacheConfig
}

type CacheConfig struct {
	TTL   int         // CACHE_TTL
	Redis RedisConfig
}

type RedisConfig struct {
	URL string // CACHE_REDIS_URL
}
```

To change or remove a prefix, see [The `comfy` tag](#the-comfy-tag).

## Secrets from files

Every field can load its value from a file. Set `<NAME>_FILE` to the path of the file, and Comfy reads the file content as the value:

```sh
DB_PASSWORD_FILE=/run/secrets/db_password
```

You do not need to change your struct for this. If both `DB_PASSWORD` and `DB_PASSWORD_FILE` are set, `DB_PASSWORD` wins.

Comfy reads the file one time, when the configuration loads. If the file can change while your application runs, use [`FileField`](#values-that-change-at-runtime-filefield).

# Customization

## The `comfy` tag

Use the `comfy` tag to change how Comfy loads a field. The tag has this format:

```
comfy:"[name][,option][,option]..."
```

The first item is always the environment variable name. To keep the default name, leave it empty. Options come after it, separated by commas.

| Item        | Use it on          | What it does                                                |
| :---------- | :----------------- | :---------------------------------------------------------- |
| `name`      | Any field          | Replaces the name that Comfy makes from the field name.     |
| `no-prefix` | Nested struct field | The fields inside the struct do not get the struct's prefix. |

An unknown option makes loading fail.

### Custom environment variable name

```go
type Config struct {
	Host string `comfy:"APP_HOST"` // APP_HOST instead of HOST
}
```

The parent prefix is still added. Here, `Host` reads `DB_ADDRESS`:

```go
type Config struct {
	DB DBConfig
}

type DBConfig struct {
	Host string `comfy:"ADDRESS"` // DB_ADDRESS
}
```

On a nested struct field, the name becomes the prefix for its fields:

```go
type Config struct {
	DB DBConfig `comfy:"DATABASE"`
}

type DBConfig struct {
	Host string // DATABASE_HOST
}
```

### Remove a nested struct prefix: `no-prefix`

```go
type Config struct {
	DB DBConfig `comfy:",no-prefix"`
}

type DBConfig struct {
	Host string // HOST, not DB_HOST
}
```

The prefixes from the levels above are kept. If `Config` is itself nested under a field `App`, `Host` reads `APP_HOST`.

## Custom types

To load a type that Comfy does not support, implement the `comfy.ComfyParser` interface:

```go
type ComfyParser interface {
	ComfyParse(s string) error
}
```

Comfy gives `ComfyParse` the raw string value. The method must set the value in place, so use a pointer receiver:

```go
// Hosts reads a comma-separated list, for example "a.example.com,b.example.com".
type Hosts []string

func (h *Hosts) ComfyParse(s string) error {
	*h = strings.Split(s, ",")
	return nil
}

type Config struct {
	AllowedHosts Hosts // ALLOWED_HOSTS
}
```

If `ComfyParse` returns an error, loading fails with a [`TypeMismatchError`](#errors).

## Values that change at runtime: `FileField`

Some values change while your application runs, for example a token that another process renews. A `comfy.FileField[T]` does not keep a copy of the value. It reads the file each time you ask for the value.

```go
type Config struct {
	APIToken comfy.FileField[string] // needs API_TOKEN_FILE
}

conf, err := comfy.LoadConfig[Config]()
if err != nil {
	log.Fatal(err)
}

token, err := conf.APIToken.Value() // reads and parses the file now
```

To run code when the value changes, subscribe to it:

```go
err = conf.APIToken.SubscribeToChange("refresh-client", func(id string, oldValue, newValue string) {
	client.SetToken(newValue)
})
```

Things to know:

- A `FileField` reads only from a file. `<NAME>_FILE` must be set, or loading fails.
- `T` can be any [supported type](#supported-types) that is also `comparable`.
- Comfy checks the file one time each second to find changes.
- The `id` names the subscription. If you subscribe again with the same `id`, the new function replaces the old one.

## Dotenv files

Before Comfy reads any value, it loads dotenv files into the process environment. It looks for these files in the current working directory, in this order:

| Order | File                       | Use it for                                    | Commit to git? |
| :---- | :------------------------- | :-------------------------------------------- | :------------- |
| 1     | `.env`                     | Shared values for all developers.             | Yes            |
| 2     | `.env.local`               | Your local changes to `.env`.                 | No             |
| 3     | `.env.<environment>`       | Values for one [environment](#application-environments). | Yes |
| 4     | `.env.<environment>.local` | Your local changes for one environment.       | No             |

- Comfy skips files that do not exist.
- Files 3 and 4 load only when an [application environment](#application-environments) is set.
- A file that loads later replaces the values of the files before it.
- Dotenv files do not replace environment variables that are already set. A real `PORT=9000` wins over `PORT=8080` in `.env`.

### File format

Write one `NAME=value` per line:

```sh
# Lines that are not NAME=value are skipped, like this comment.
HOST=localhost
PORT=8080
```

### Load dotenv files from another directory

Call `dotenv.SetEnvsDir` before you load the configuration:

```go
import "github.com/Mendes11/comfy/dotenv"

if err := dotenv.SetEnvsDir("./config"); err != nil {
	log.Fatal(err) // the directory does not exist
}
conf, err := comfy.LoadConfig[Config]()
```

## Application environments

An application environment is a name like `development`, `test`, or `production`. Comfy uses it to:

- load `.env.<environment>` and `.env.<environment>.local` (see [Dotenv files](#dotenv-files)), and
- pick the [environment defaults](#environment-defaults).

By default, Comfy reads the environment from `APP_ENV`:

```sh
APP_ENV=production ./myapp
```

You can change this:

```go
// Read the environment from another variable.
comfy.SetAppEnvName("ENVIRONMENT")

// Or set it in code. This wins over the environment variable.
comfy.SetAppEnvironment("test")

// Read the current environment. It is "" when none is set.
env := comfy.AppEnvironment()
```

## Environment defaults

This feature is inspired by Ruby on Rails environments. You can register a set of defaults for each application environment:

```go
type Config struct {
	StorageBackend string
	DBName         string
}

var DevelopmentConfig = Config{
	StorageBackend: "local",
	DBName:         "myapp_development",
}

var ProductionConfig = Config{
	StorageBackend: "s3",
	DBName:         "myapp",
}

func main() {
	comfy.SetEnvironmentDefault("development", DevelopmentConfig)
	comfy.SetEnvironmentDefault("production", ProductionConfig)

	conf, err := comfy.LoadConfig[Config]()
	if err != nil {
		log.Fatal(err)
	}
}
```

With `APP_ENV=development`, `conf` starts as a copy of `DevelopmentConfig`. Values from the environment then replace the defaults, as usual.

Things to know:

- Only `LoadConfig` uses environment defaults. `Load` does not.
- Register the defaults before you call `LoadConfig`.
- The defaults must have the same type as the `T` in `LoadConfig[T]`.

# Errors

Loading stops at the first error. These are the errors Comfy returns:

| Error                         | When                                                                  |
| :---------------------------- | :-------------------------------------------------------------------- |
| `comfy.ErrExpectedPtr`        | You gave `Load` a value, not a pointer.                               |
| `comfy.ErrExpectedStructType` | You gave `Load` a pointer to something that is not a struct.          |
| `comfy.TypeMismatchError`     | A value cannot be converted to the field type, for example `PORT=abc`. |
| `comfy.MissingParserError`    | A field has a type that Comfy cannot load. See [Custom types](#custom-types). |

Comfy wraps errors, so use `errors.Is` and `errors.As` to check them:

```go
conf, err := comfy.LoadConfig[Config]()

var mismatch comfy.TypeMismatchError
if errors.As(err, &mismatch) {
	log.Fatalf("field %s: cannot use %q as %s", mismatch.FieldName, mismatch.FieldValue, mismatch.TargetType)
}
```

Comfy can also return errors when it cannot read a dotenv file, a `<NAME>_FILE` file, or when a `comfy` tag has an unknown option.

# API overview

| Function / type                          | What it does                                                         |
| :--------------------------------------- | :------------------------------------------------------------------- |
| `comfy.LoadConfig[T]() (T, error)`       | Makes a new `T` and loads it. Uses environment defaults.             |
| `comfy.Load(conf any) error`             | Loads into a struct you already have. Keeps the values that are not set. |
| `comfy.SetAppEnvName(name string)`       | Changes the variable that holds the environment name (default `APP_ENV`). |
| `comfy.SetAppEnvironment(env string)`    | Sets the application environment in code.                            |
| `comfy.AppEnvironment() string`          | Returns the current application environment.                         |
| `comfy.SetEnvironmentDefault(env string, conf any)` | Registers the defaults for one environment.               |
| `comfy.ConvertToEnvName(name string) string` | Shows the environment variable name for a field name.            |
| `comfy.FileField[T]`                     | A field that reads its value from a file each time you ask.          |
| `comfy.ComfyParser`                      | The interface for your own types.                                    |
| `dotenv.SetEnvsDir(dir string) error`    | Changes the directory where Comfy looks for dotenv files.            |

For the full reference, see [pkg.go.dev](https://pkg.go.dev/github.com/Mendes11/comfy).
