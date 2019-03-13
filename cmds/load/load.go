package load

import (
	"flag"
	"github.com/genshen/cmds"
)

var loadCommand = &cmds.Command{
	Name:        "load",
	Summary:     "load dependency packages from tarball file",
	Description: "import and extract dependency packages from tarball file (.tar) specified by a file path",
	CustomFlags: false,
	HasOptions:  true,
}

func init() {
	var l load
	fs := flag.NewFlagSet("load", flag.ContinueOnError)
	loadCommand.FlagSet = fs
	loadCommand.FlagSet.StringVar(&l.url, "f", "", "tarball file path")
	// loadCommand.FlagSet.StringVar(&output, "o", "", "output directory")
	loadCommand.FlagSet.Usage = loadCommand.Usage // use default usage provided by cmds.Command.
	loadCommand.Runner = &l
	cmds.AllCommands = append(cmds.AllCommands, loadCommand)
}

type load struct {
	output string
	url    string
}

func (v *load) PreRun() error {
	return nil
}

func (v *load) Run() error {
	return nil
}
