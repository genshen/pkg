package main

import (
	"github.com/genshen/cmds"
	_ "github.com/genshen/pkg/pkg/clean"
	_ "github.com/genshen/pkg/pkg/export"
	_ "github.com/genshen/pkg/pkg/fetch"
	_ "github.com/genshen/pkg/pkg/import"
	_ "github.com/genshen/pkg/pkg/init"
	_ "github.com/genshen/pkg/pkg/install"
	_ "github.com/genshen/pkg/pkg/version"
	log "github.com/sirupsen/logrus"
)

func main() {
	cmds.SetProgramName("pkg")
	err := cmds.Parse()
	if err != nil {
		log.Fatal(err)
	}
}
