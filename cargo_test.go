package cargo

//The options here for cargo is to load same command line parameters from a config file

import (
	"bytes"
	"fmt"
	"os/user"
	"path"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getUserSubDir(s string) string {
	var err error
	usr, err = user.Current()
	if err != nil {
		fmt.Println("error reading current user")
		return ""
	}
	return fmt.Sprintf("%s/%s", usr.HomeDir, s)
}

func TestConfigInitialize(t *testing.T) {
	g := NewConf("prova")
	assert.NotNil(t, g, "fail allocating conf set")
	g.AddSearchPath("~/bin")
	g.AddSearchPath("/test/prova/123") //absolute should be respected
	g.AddSearchPath("prova/123")       //relative should be resepected
	assert.NotEmpty(t, g.SearchPaths, "search path not addedd")
	assert.Len(t, g.SearchPaths, 3, "search path not addedd")
	assert.Equal(t, getUserSubDir("bin"), g.SearchPaths[0], "fail to resolve home directory")
	assert.Equal(t, "/test/prova/123", g.SearchPaths[1], "absolute path non respected")
	assert.Equal(t, "prova/123", g.SearchPaths[2], "relative path not respected")
	//initialize on multiple file name
	g1 := NewConf("prova2", "t1.conf", "t2.yaml", "t3.toml")
	assert.NotNil(t, g1, "fail allocating conf set")
	assert.Len(t, g1.FileNameFallback, 4, "some config file discarded")
	assert.EqualValues(t, []string{"prova2.conf", "t1.conf", "t2.yaml", "t3.toml"}, g1.FileNameFallback, "wrong values within fallback list")
}

type Option struct {
	TestSample       string `cargo:"partest,,"`
	TestNoName       string `cargo:",,this is the description"`
	SuppForTestSuite bool   `cargo:"test.v,, this come from test suite"`
	Escaping         string `cargo:"escape,prova\\,prova\\,aaa,this is a complex test"`
	TestInt          int    `cargo:"testint,10,this is an int"`
	unExportedVar    string `cargo:"BigHere,not present,"`
	Ignored          bool   `cargo:"-,true,ignored"`
}

func TestAddOptions(t *testing.T) {
	g := NewConf("again")
	assert.NotNil(t, g, "fail allocating conf set")
	o := &Option{}
	g.AddOptions(o)
	assert.NotEmpty(t, g.FlagSet.Lookup("partest"), "fail to get simple string option")
	assert.NotEmpty(t, g.FlagSet.Lookup("testnoname"), "fail to get simple string option")
	assert.NotEmpty(t, g.FlagSet.Lookup("test.v"), "fail to get simple string option")
	assert.NotEmpty(t, g.FlagSet.Lookup("escape"), "fail to get simple string option")
	assert.NotEmpty(t, g.FlagSet.Lookup("testint"), "fail to get simple string option")
	assert.Empty(t, g.FlagSet.Lookup("BigHere"), "fail to get simple string option")
	assert.Empty(t, g.FlagSet.Lookup("ignored"), "fail to get simple string option")
}

func TestLoadConfsFromByteArray(t *testing.T) {
	g := NewConf("again")
	assert.NotNil(t, g, "fail allocating conf set")
	o := &Option{}
	g.AddOptions(o)
	b := bytes.NewBufferString(`
partest="hi"
testnoname="this is standard"
byebeast="blablabla"
testint=11
[test]
  v=false
  `)
	g.LoadFromBuffer(b.Bytes())
	assert.Equal(t, "hi", o.TestSample, "failing setting value for simple string")
	assert.Equal(t, "this is standard", o.TestNoName, "fail setting value for no name")
	assert.Equal(t, false, o.SuppForTestSuite, "fail to set value in nested struct")
	assert.Equal(t, 11, o.TestInt, "fail to set value in nested struct")

	type MyOption struct {
		TestSample       string `cargo:"partest,,"`
		TestNoName       string `cargo:",,this is the description"`
		SuppForTestSuite bool   `cargo:"test.sample,, this come from test suite"`
		TestInt          int    `cargo:"testint,10,this is an int"`
		TestTrue         bool   `cargo:"testtrue,false,test again"`
		unExportedVar    string `cargo:"BigHere,not present,"`
		Ignored          bool   `cargo:"-,true,ignored"`
	}

	g1 := NewConf("again")
	assert.NotNil(t, g1, "fail allocating conf set")
	o1 := &MyOption{}
	g1.AddOptions(o1)
	b1 := bytes.NewBufferString(`
partest="hi"
testnoname="this is standard"
byebeast="blablabla"
testint=10
testtrue=true
[test]
  v=true
`)
	g1.LoadFromBuffer(b1.Bytes())
	assert.Equal(t, "hi", o1.TestSample, "failing setting value for simple string")
	assert.Equal(t, "this is standard", o1.TestNoName, "fail setting value for no name")
	assert.Equal(t, false, o1.SuppForTestSuite, "nested struct gets default value")
	assert.Equal(t, true, o1.TestTrue, "fail testing true  ")
	assert.Equal(t, 10, o1.TestInt, "fail to set value for integer")

}

func TestFromCommandLine(t *testing.T) {
	type MyOption struct {
		TestSample       string `cargo:"partest,,"`
		TestNoName       string `cargo:",,this is the description"`
		SuppForTestSuite bool   `cargo:"test.sample,, this come from test suite"`
		TestInt          int    `cargo:"testint,10,this is an int"`
		unExportedVar    string `cargo:"BigHere,not present,"`
		Ignored          bool   `cargo:"-,true,ignored"`
	}

	g1 := NewConf("again")
	assert.NotNil(t, g1, "fail allocating conf set")
	o1 := &MyOption{}
	g1.AddOptions(o1)
	p := []string{
		"--partest=blablabla",
		"--testnoname=noname",
		"--test.sample",
	}
	g1.Parse(p)
	assert.Equal(t, "blablabla", o1.TestSample, "failing setting value for simple string")
	assert.Equal(t, "noname", o1.TestNoName, "fail setting value for no name")
	assert.Equal(t, true, o1.SuppForTestSuite, "nested struct gets default value")

}

func TestConfFromFiles(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)

	type Option struct {
		TestSample       string `cargo:"partest,,"`
		TestNoName       string `cargo:",,this is the description"`
		SuppForTestSuite bool   `cargo:"test.v,, this come from test suite"`
		Escaping         string `cargo:"escape,prova\\,prova\\,aaa,this is a complex test"`
		TestInt          int    `cargo:"testint,10,this is an int"`
		unExportedVar    string `cargo:"BigHere,not present,"`
		Ignored          bool   `cargo:"-,true,ignored"`
	}

	g := NewConf("prova", "test.conf")
	assert.NotNil(t, g, "fail allocating conf set")
	g.AddSearchPath("~/bin")
	g.AddSearchPath("/test/prova/123") //absolute should be respected
	g.AddSearchPath("prova/123")       //relative should be resepected
	g.AddSearchPath(path.Dir(filename))
	confList := g.getConfFileList()
	assert.Len(t, confList, 8, "some path to possible config file not created")
	assert.EqualValues(t, []string{
		getUserSubDir("bin/prova.conf"),
		getUserSubDir("bin/test.conf"),
		"/test/prova/123/prova.conf",
		"/test/prova/123/test.conf",
		"prova/123/prova.conf",
		"prova/123/test.conf",
		path.Dir(filename) + "/prova.conf",
		path.Dir(filename) + "/test.conf"}, confList, "something wrong with conf file list")
	o := &Option{}
	g.AddOptions(o)
	g.Load()
	assert.Equal(t, "fromconfigfile", o.TestSample, "not setted from config file")
	assert.Equal(t, "this is from config file", o.TestNoName, "no name variable missing from config file")
	assert.Equal(t, 54, o.TestInt, "not working int from config file")
}

func TestConfPanicByOpt(t *testing.T) {
	g := NewConf("xxx")
	o := Option{}
	g.AddOptions(&o)
	assert.Panics(t, func() {
		g.Load()
	}, "Should panic")
}

func TestConfPanicWithFile(t *testing.T) {
	filename := "test_panic.conf"
	g := NewConf("test_panic")
	g.AddSearchPath(path.Dir(filename))
	o := Option{}
	g.AddOptions(&o)
	assert.Panics(t, func() {
		g.Load()
	}, "Should panic")
}

func TestConfPanicByBuffer(t *testing.T) {
	g := NewConf("again")
	assert.NotNil(t, g, "fail allocating conf set")
	o := Option{}
	g.AddOptions(&o)
	b := bytes.NewBufferString(`
partest="hi"
testnoname=""
byebeast="blablabla"
testint=10
testtrue=true
[test]
  v=true
`)
	assert.Panics(t, func() {
		g.LoadFromBuffer(b.Bytes())
	}, "Should panic")
}

func TestConfNotPanic(t *testing.T) {
	filename := "test.conf"
	g := NewConf("test")
	g.AddSearchPath("~/aaa")
	g.AddSearchPath("~/bbb")
	g.AddSearchPath(path.Dir(filename))
	o := Option{}
	g.AddOptions(&o)
	assert.NotPanics(t, func() {
		g.Load()
	}, "Should not panic")
}

// func TestConfAsMain(t *testing.T) {
// 	g := NewConf("prova", "test.conf")
// 	g.AddSearchPath("~/bin")
// 	o := Option{}
// 	g.AddOptions(&o)
// 	g.Load()
// 	if o.TestSample != "hi" {
// 		fmt.Println("merda", o.TestSample)
// 		t.Error("failed to get value for test from config file")
// 	}
// 	if o.TestNoName != "this is standard" {
// 		fmt.Println(o.TestNoName)
// 		t.Error("failed to get value for testnoname from config file")
// 	}
// 	if o.SuppForTestSuite != true {
// 		t.Error("failed to override conf file value with command line ones for test.v")
// 	}
// }
