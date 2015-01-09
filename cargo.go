package cargo

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/user"
	"path"
	"reflect"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/stvp/go-toml-config"
)

// Conf is structure containing configuration
type Conf struct {
	*config.ConfigSet
	FileName    string
	SearchPaths []string
}

var (
	usr     *user.User
	homedir string
)

func init() {
	var err error
	usr, err = user.Current()
	if err != nil {
		fmt.Println("merda non riesco a leggere l'untete")
		return
	}
	homedir = usr.HomeDir

}

func panicIf(e error) {
	if e != nil {
		panic(e)
	}
}

func pathExistsAsFolder(path string) bool {
	t, err := os.Stat(path)
	if err != nil {
		return false
	}
	return t.IsDir()
}

func pathExistsAsFile(path string) bool {
	t, err := os.Stat(path)
	if err != nil {
		return false
	}
	return t.Mode().IsRegular()
}

// NewConf return new config object
func NewConf(progName string, cFileName string) *Conf {
	return &Conf{
		config.NewConfigSet(progName, config.ExitOnError),
		cFileName,
		make([]string, 0),
	}
}

//AddOptions add handled option
func (g *Conf) AddOptions(options interface{}) {
	x := reflect.ValueOf(options)
	s := x.Elem()
	typeOfT := s.Type()
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		tag := typeOfT.Field(i).Tag
		name := tag.Get("toml")
		if name == "" || name == "-" {
			name = strings.ToLower(typeOfT.Field(i).Name)
		}
		fmt.Println(typeOfT.Field(i).Name, f.Type())
		switch fvar := f.Addr().Interface().(type) {
		case *string:
			g.FlagSet.StringVar(fvar,
				name,
				tag.Get("default"),
				tag.Get("description"))
		case *bool:
			v, _ := strconv.ParseBool(tag.Get("default"))
			g.FlagSet.BoolVar(fvar,
				name,
				v,
				tag.Get("description"))
		}

	}
}

// AddSearchPath add new search path to the set
func (g *Conf) AddSearchPath(newPath string) bool {
	tmp := newPath
	if strings.HasPrefix(tmp, "~/") {
		tmp = strings.Replace(tmp, "~", homedir, 1)
	}
	if pathExistsAsFolder(tmp) {
		g.SearchPaths = append(g.SearchPaths, tmp)
		return true
	}
	return false
}

// Load loads options from command line and conf files
func (g *Conf) Load() {
	i := 0
	err := g.Parse(g.FileName)
	for ; i < len(g.SearchPaths) && err != nil; err, i = g.Parse(path.Join(g.SearchPaths[i], g.FileName)), i+1 {
	}
	//panicIf(err)
	g.FlagSet.Parse(os.Args[1:])
}

// Serialize write current running configuration to a user related config file
func (g *Conf) Serialize(options interface{}) {
	buf := new(bytes.Buffer)
	if err := toml.NewEncoder(buf).Encode(options); err != nil {
		fmt.Println("Fail to serialize: ", err)
	}
	// we can define here a default file place
	f, err := os.Create(g.FileName)
	defer f.Close()
	panicIf(err)
	w := bufio.NewWriter(f)
	w.Write(buf.Bytes())
	w.Flush()
}
