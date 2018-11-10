package main

import (
	"github.com/genshen/cmds"
	_ "github.com/genshen/pkg/fetch"
	_ "github.com/genshen/pkg/install"
	_ "github.com/genshen/pkg/version"
	log "github.com/sirupsen/logrus"
)

func main() {
	cmds.SetProgramName("pkg")
	err := cmds.Parse()
	if err != nil {
		log.Error(err)
	}
}
