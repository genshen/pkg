package main

import (
	"github.com/genshen/cmds"
	_ "github.com/genshen/pkg/fetch"
	_ "github.com/genshen/pkg/install"
	_ "github.com/genshen/pkg/version"
)

func main() {
	cmds.SetProgramName("pkg")
	cmds.Parse()
}
