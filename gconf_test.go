package gconf

import (
	"fmt"
	"testing"
)

type Option struct {
	Prova  string `toml:"prova"`
	Prova1 string `description:"merda"`
}

func TestConfAsMain(*testing.T) {
	g := NewConf("prova", "test.conf")
	g.AddSearchPath("~/bin")
	o := Option{}
	g.AddOptions(&o)
	g.LoadOptions()
	fmt.Println(o.Prova)
}
