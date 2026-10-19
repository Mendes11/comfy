<p align="center">
    <picture>
      <source media="(prefers-color-scheme: light)" srcset="docs/content/assets/img/comfy.logo.png">
      <img alt="Comfy" title="Comfy" height="400" width="400" src="docs/content/assets/img/comfy.logo.png">
    </picture>
</p>

Comfortable, opinionated configurations management for Golang applications.

# About

Comfy abstracts for you all the annoying parts of dealing with configurations in your application.
It'll automatically set naming conventions and support for file secrets out of the box. No extra configuration needed besides you declaring your config struct.

# Usage

```go
type Config struct {
  Host string // Resolves to HOST
  Port int // Resolves to PORT
  APIKey string // Resolves to API_KEY
  MyConfig MyOwnType // MY_CONFIG -- As long as MyOwnType implements [comfy.ComfyParser], comfy can load it.
  DB DBConfig // comfy supports nested structs, by default it adds a prefix to their names DB_<field>
}

type DBConfig struct {
  Host string // Resolves to DB_HOST
  Port int // Resolves to DB_PORT
  User string // Resolves to DB_USER
  Password string // Resolves to DB_PASSWORD
  Name string // Resolves to DB_NAME
}

conf, err := comfy.LoadConfiguration[Config]()
```

That's it. Comfy will automatically load from your **environment variables**, **dotfiles**, and **secret files**.

## Conventions

Comfy will convert the golang struct field names into an environment variable name for you following a few conventions.

1. All names converted into Upper case
2. Add a "_" between words when an upper cased character is followed by a lower case one. eg: MyKey => MY_KEY
3. Nested structures, by default, have their struct field name added as a prefix to its fields. You can configure that.
4. All loadable fields have **native support** for loading their values from a file if an `<ENV_NAME>_FILE` is set. Eg: `DB_PASSWORD_FILE=/var/secrets/DB_PASSWORD`

## Configurations

You can configure how comfy will load a field using the `comfy` tag. Each configuration is separated by a comma ",". The first field will always be the environment name. If you want to keep the default, just leave it empty. Eg: `comfy:",no-prefix"`

`comfy:"[envName],[config[=value]]"`

### Custom ENV name

If you want to use a different name than the one comfy uses, you can pass it as the first argument to the `comfy` tag.

```go
type Config struct {
  Host string `comfy:"APP_HOST"`
}
```

### No nested structure prefix: `no-prefix`

If you don't want to have the nested structure's name prefixed on the environment variable name, you can have add the `no-prefix` tag to it.

eg:

```go
type Config struct {
  DB DBConfig `comfy:",no-prefix"`
}

type DBConfig struct {
  Host string
}
```

Comfy will resolve `DB.Host` as `HOST` only.

## Advanced Uses

Comfy not only loads from environment variables, but it also loads from `dotenv` files. By default it'll look for dotenv files in the current directory and load them as environment variables before processing the configurations.

The loading order is:

```
| Order       | Filename                 | Notes                             | Versionable |
| ----------- | ------------------------ | --------------------------------- | ----------- |
| 1           | .env                     | common values for any dev to use. | Yes         |
| 2           | .env.local               | local overrides of .env           | No          |
| 3           | .env.<environment>       | Environment-specific variables.   | Yes         |
| 4           | .env.<environment>.local | Environment-specific override     | No          |

```
