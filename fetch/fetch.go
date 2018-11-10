package fetch

import (
	"flag"
	"github.com/genshen/cmds"
)

var fetchCommand = &cmds.Command{
	Name:        "fetch",
	Summary:     "fetch a package from remote",
	Description: "fetch a cpp package(zip,cmake,makefile,.etc format) from remote.",
	CustomFlags: false,
	HasOptions:  true,
}

var (
	output string
	url    string
)

func init() {
	fetchCommand.Runner = &get{}
	fs := flag.NewFlagSet("fetch", flag.ContinueOnError)
	fetchCommand.FlagSet = fs
	fetchCommand.FlagSet.StringVar(&url, "url", "", "addr")
	fetchCommand.FlagSet.StringVar(&output, "o", "", "output directory")
	fetchCommand.FlagSet.Usage = fetchCommand.Usage // use default usage provided by cmds.Command.
	cmds.AllCommands = append(cmds.AllCommands, fetchCommand)
}

type get struct{}

func (v *get) PreRun() error {
	return nil
}

func (v *get) Run() error {
	return nil
}
