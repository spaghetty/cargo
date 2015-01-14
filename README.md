Cargo
-----

cargo is a library for simple management of command line parameters and config
file, using a single structure of options.

### Installation

```
go get github.com/spaghetty/cargo
```

### Usage

```go
import (
  "fmt"

  "github.com/spaghetty/cargo"
  )

type Option struct {
  OptionOne     string    `toml:"option_one", default:"", description:"this is"`
  OptionTwo     string    `description:"this one will be serialized lowered"`
  Serialize     bool      `toml:"-", default:"false",description:"boh"`
}

var (
  Configs *cargo.Conf
  Options Option
)

func init() {
  Configs = cargo.NewConf("prova", "test.conf")
  Configs.AddSearchPath("~/bin")
  Options = Option{}
  Configs.AddOptions(&Options)

}

func main() {
  Configs.Load()
  if Options.Serialize {
    Configs.Serialize(Options)
  }
  fmt.Println(Options.OptionOne)
}
```
