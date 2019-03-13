package load

import (
	"flag"
	"github.com/genshen/cmds"
)

var fetchCommand = &cmds.Command{
	Name:        "load",
	Summary:     "load dependency packages from tarball file",
	Description: "import and extract dependency packages from tarball file (.tar) specified by a file path",
	CustomFlags: false,
	HasOptions:  true,
}

var (
	output string
	url    string
)

func init() {
	fetchCommand.Runner = &load{}
	fs := flag.NewFlagSet("load", flag.ContinueOnError)
	fetchCommand.FlagSet = fs
	fetchCommand.FlagSet.StringVar(&url, "f", "", "tarball file path")
	// fetchCommand.FlagSet.StringVar(&output, "o", "", "output directory")
	fetchCommand.FlagSet.Usage = fetchCommand.Usage // use default usage provided by cmds.Command.
	cmds.AllCommands = append(cmds.AllCommands, fetchCommand)
}

type load struct{}

func (v *load) PreRun() error {
	return nil
}

func (v *load) Run() error {
	return nil
}
