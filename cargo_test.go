package cargo

import (
	"fmt"
	"testing"
)

type Option struct {
	TestSample       string `cargo:"partest,,"`
	TestNoName       string `cargo:",,this is the description"`
	SuppForTestSuite bool   `cargo:"test.v,, this come from test suite"`
}

func TestConfAsMain(t *testing.T) {
	g := NewConf("prova", "test.conf")
	g.AddSearchPath("~/bin")
	o := Option{}
	g.AddOptions(&o)
	g.Load()
	if o.TestSample != "hi" {
		fmt.Println(o.TestSample)
		t.Error("failed to get value for test from config file")
	}
	if o.TestNoName != "ths is standard" {
		t.Error("failed to get value for testnoname from config file")
	}
	if o.SuppForTestSuite != true {
		t.Error("failed to override conf file value with command line ones for test.v")
	}
}
