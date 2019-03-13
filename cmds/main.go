package main

import (
	"github.com/genshen/cmds"
	_ "github.com/genshen/pkg/cmds/fetch"
	_ "github.com/genshen/pkg/cmds/init"
	_ "github.com/genshen/pkg/cmds/install"
	_ "github.com/genshen/pkg/cmds/load"
	_ "github.com/genshen/pkg/cmds/version"
	log "github.com/sirupsen/logrus"
)

func main() {
	cmds.SetProgramName("pkg")
	err := cmds.Parse()
	if err != nil {
		log.Error(err)
	}
}
