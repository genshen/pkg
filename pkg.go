package main

import (
	_ "github.com/genshen/pkg/fetch"
	_ "github.com/genshen/pkg/install"
	_ "github.com/genshen/pkg/version"
	"github.com/genshen/cmds"
)

func main() {
	cmds.SetProgramName("pkg")
	cmds.Parse()
}
