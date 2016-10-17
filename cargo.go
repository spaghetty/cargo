package cargo

import (
	"bufio"
	"bytes"
	"flag"
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
		fmt.Println("error reading current user")
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

func getSlice(s string) (string, int) {
	currIndex := strings.Index(s, ",")
	if currIndex < 0 {
		return s, len(s)
	}
	if currIndex > 0 && s[currIndex-1] == '\\' {
		tmp, i := getSlice(s[currIndex+1:])
		return fmt.Sprintf("%s,%s", s[:currIndex-1], tmp), len(s[:currIndex]) + i + 1
	}
	return s[:currIndex], len(s[:currIndex])
}

//AddOptions add handled option
func (g *Conf) AddOptions(options interface{}) {
	x := reflect.ValueOf(options)
	s := x.Elem()
	typeOfT := s.Type()
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		tag := typeOfT.Field(i).Tag
		data := tag.Get("cargo")
		var name, defaultVal, description string
		var nl, dl int
		name, nl = getSlice(data)
		defaultVal, dl = getSlice(data[nl+1:])
		description, _ = getSlice(data[nl+dl+2:])
		if name == "" || name == "-" {
			name = strings.ToLower(typeOfT.Field(i).Name)
		}
		//fmt.Println(typeOfT.Field(i).Name, f.Type())
		switch fvar := f.Addr().Interface().(type) {
		case *string:
			g.FlagSet.StringVar(fvar,
				name,
				defaultVal,
				description)
		case *bool:
			v, _ := strconv.ParseBool(defaultVal)
			g.FlagSet.BoolVar(fvar,
				name,
				v,
				description)
		case *int:
			v, _ := strconv.Atoi(defaultVal)
			g.FlagSet.IntVar(fvar, name, v, description)
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
	if err != nil {
		fmt.Println("Parse File Error:", err.Error())
	}
	g.FlagSet.Parse(os.Args[1:])
	visitor := func(a *flag.Flag) {
		if a.Value.String() == "" {
			fmt.Printf("Flag %s has an empty value\n", a.Name)
			panic("flag empty value")
		}
	}
	g.FlagSet.VisitAll(visitor)
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
