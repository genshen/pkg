package main

import (
	"errors"
	"flag"

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
	if err := cmds.Parse(); err != nil {
		if err == flag.ErrHelp {
		    return
		}
		// skip error in sub command parsing, because the error has been printed.
		if !errors.Is(err, &cmds.SubCommandParseError{}) {
			log.Fatal(err)
		}
	}
}
