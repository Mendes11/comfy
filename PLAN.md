# Plan

## Objetivos

Biblioteca simples de gestão de configurações via Struct, povoando por meio de leitura de env.

## Main features

### Env Files

Ler .env files, com ordem de prioridade, (maior vence)

| Prioridade |          Arquivo           |
|     0      |   .env                     |
|     1      |   .env.local               |
|     2      |   .env.<environment>       |
|     3      |   .env.<environment>.local |

Dessa maneira, o app irá setar o environment e o comfy vai carregar os arquivos corretos, na ordem correta, para preencher as variáveis de ambiente, permitindo termos múltiplas configurações, por ambiente.

O environment será definido através da variável "ENV". Quando não especificado, ele irá somente considerar os dois primeiros arquivos da tabela.

### _FILE suffix

Quando declaro a variável de ambiente, o comfy irá ler tanto dela como de sua variante com o sufixo "_FILE". Isso permitirá lermos qualquer configuração vinda de secret files.

### Naming Conventions

Para facilitar, o comfy vai ter uma convenção de nomes de variáveis, assim não precisamos ter que declarar o nome para cada field. A convenção é puramente substituir cammelCase por Upper case com snake_case.

DbHost => D_B_HOST
DBHost => DBHOST
ServerURL => SERVER_URL

Talvez eu possa criar uma regra, onde caso haja uma sequência de caracteres em upper case, que não terminem no final da palavra, eu crio um underscore entre os dois últimos.

DBHost => DB_HOST
TestID => TEST_ID
ODBCString => ODBC_STRING
ServerURL => SERVER_URL

Estruturas "nested" irão ser delimitadas por um underscore. Eu posso configurar o comfy para não incluir o prefixo para uma estrutura

```go
type Config {
  Server ServerConfig `comfy="no-prefix"`
  DB DBConfig 
  
}

type ServerConfig struct {
  Host string
  Port int
}

type DBConfig struct {
  Name string
  Host string
  Port string
}
```

Com a convenção, as variáveis no ServerConfig não serão carregadas com o prefixo, e DB serão:

HOST
PORT
DB_NAME
DB_HOST
DB_PORT

### Default Structures

Eu gosto da ideia de termos defaults declarados como structs, não como tags, mas acho que podemos ter os dois. Com o default struct ganhando em prioridade.

devConfig = Config{
  Host: "localhost",
  Port: 3000,
  DBName: "proj_development",
}

testConfig = Config{
  Host: "localhost",
  Port: 3000,
  DBName: "proj_test"
}

comfy.SetDefault(devConfig)
comfy.SetDefaultForEnvironment(devConfig, "development")
comfy.SetDefaultForEnvironment(testConfig, "test")

### Type Conversions

Definir uma interface para conversão de tipos, onde eu crio uma conversão de string para aquele tipo:

Tipos customizados só precisam implementar a interface `Parser`

```go

type Parser[K any] interface {
  Parse(v string) (K, error)
}

func (s int) Parse(v string) (int, error) {
   return strconv.Atoi(v)
}
```
