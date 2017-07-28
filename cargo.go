package cargo

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/user"
	"path"
	"reflect"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

// Conf is structure containing configuration
type Conf struct {
	*flag.FlagSet
	SearchPaths      []string
	FileNameFallback []string
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
func NewConf(progName string, files ...string) *Conf {
	tmp := &Conf{
		flag.NewFlagSet(progName, flag.ExitOnError),
		//config.NewConfigSet(progName, config.ContinueOnError),
		make([]string, 0),
		make([]string, len(files)+1),
	}
	tmp.FileNameFallback[0] = progName + ".conf"
	for i, fn := range files {
		tmp.FileNameFallback[i+1] = fn
	}
	return tmp
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
		if !f.CanSet() { //if the filed is private just skip it
			continue
		}
		tag := typeOfT.Field(i).Tag
		data := tag.Get("cargo")
		var name, defaultVal, description string
		var nl, dl int
		name, nl = getSlice(data)
		defaultVal, dl = getSlice(data[nl+1:])

		if len(defaultVal) > 0 && []rune(defaultVal)[0] == '$' {
			defaultVal = os.Getenv(defaultVal[1:])
		}
		description, _ = getSlice(data[nl+dl+2:])
		if name == "" {
			name = strings.ToLower(typeOfT.Field(i).Name)
		}
		if name == "-" {
			continue
		}
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
func (g *Conf) AddSearchPath(newPath string) error {
	tmp := newPath
	if strings.HasPrefix(tmp, "~/") {
		tmp = strings.Replace(tmp, "~", homedir, 1)
	}
	g.SearchPaths = append(g.SearchPaths, tmp)
	return nil
}

func toString(val interface{}) string {
	switch val.(type) {
	case string:
		return val.(string)
	case int, int32, int64:
		return strconv.FormatInt(val.(int64), 10)
	case bool:
		if val.(bool) {
			return "true"
		}
		return "false"
	}
	return ""
}

func (g *Conf) loadFlagsFromMap(myMap map[string]interface{}) error {
	var err error
	g.FlagSet.VisitAll(func(f *flag.Flag) {
		if strings.Contains(f.Name, ".") {
			steps := strings.Split(f.Name, ".")
			tmp := (interface{})(myMap)
			for _, level := range steps {
				var ok bool
				if tmp, ok = tmp.(map[string]interface{})[level]; !ok {
					err = errors.New("missing key")
					return
				}
			}
			f.Value.Set(toString(tmp))
			return
		}
		if val, ok := myMap[f.Name]; ok {
			f.Value.Set(toString(val))
			return
		}
	})
	return err
}

//LoadFromBuffer load configurations from byte array
func (g *Conf) LoadFromBuffer(data []byte) error {
	myMap := make(map[string]interface{})
	toml.Unmarshal(data, &myMap)
	err := g.loadFlagsFromMap(myMap)
	g.checkFlagEmptyValue()
	return err
}

func (g *Conf) getConfFileList() []string {
	filePaths := make([]string, 0, len(g.FileNameFallback)*len(g.SearchPaths))
	for _, p := range g.SearchPaths {
		for _, f := range g.FileNameFallback {
			filePaths = append(filePaths, path.Join(p, f))
		}
	}
	return filePaths
}

// Load loads options from command line and conf files
func (g *Conf) Load() {
	//i := 0
	filePaths := g.getConfFileList()
	var i int
	for ; i < len(filePaths) && !pathExistsAsFile(filePaths[i]); i++ {
	}
	if i < len(filePaths) {
		myMap := make(map[string]interface{})
		toml.DecodeFile(filePaths[i], &myMap)
		g.loadFlagsFromMap(myMap)
	} else {
		fmt.Printf("Warning parse file: file not found \n")
	}
	g.FlagSet.Parse(os.Args[1:])
	g.checkFlagEmptyValue()
}

// Check empty flag value
func (g *Conf) checkFlagEmptyValue() {
	visitor := func(a *flag.Flag) {
		if a.Value.String() == "" {
			fmt.Printf("Flag %s has an empty value\n", a.Name)
			panic("flag empty value")
		}
	}
	g.FlagSet.VisitAll(visitor)
}

// // Serialize write current running configuration to a user related config file
// func (g *Conf) Serialize(options interface{}) {
// 	buf := new(bytes.Buffer)
// 	if err := toml.NewEncoder(buf).Encode(options); err != nil {
// 		fmt.Println("Fail to serialize: ", err)
// 	}
// 	// we can define here a default file place
// 	f, err := os.Create(g.FileName)
// 	defer f.Close()
// 	panicIf(err)
// 	w := bufio.NewWriter(f)
// 	w.Write(buf.Bytes())
// 	w.Flush()
// }
