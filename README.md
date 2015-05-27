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

//the tag for cargo is build of 3 parts custom name, default value and description
// if custom name is not specified the name will be the lowered version of the field name
// - means that the filed will be ignored as parameter  
type Option struct {
  OptionOne     string    `cargo:"option_one,,this is"`
  OptionTwo     string    `cargo:",,this one will be serialized lowered"`
  Serialize     bool      `cargo:"-,false,new cool option"`
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
