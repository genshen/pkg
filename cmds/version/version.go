package version

import (
	"flag"
	"fmt"
	"github.com/genshen/cmds"
)

const VERSION = "0.3.0-alpha"

var versionCommand = &cmds.Command{
	Name:        "version",
	Summary:     "print pkg version",
	Description: "print current pkg version.",
	CustomFlags: false,
	HasOptions:  false,
}

//var (
//	h string
//)

func init() {
	versionCommand.Runner = &version{}
	fs := flag.NewFlagSet("version", flag.ContinueOnError)
	versionCommand.FlagSet = fs
	// versionCommand.FlagSet.StringVar(&h, "h", "default Haa","Hhhha")
	versionCommand.FlagSet.Usage = versionCommand.Usage // use default usage provided by cmds.Command.
	cmds.AllCommands = append(cmds.AllCommands, versionCommand)
}

type version struct{}

func (v *version) PreRun() error {
	return nil
}

func (v *version) Run() error {
	fmt.Printf("version\t %s.\n", VERSION)
	fmt.Println("Author\t genshenchu@gmail.com")
	fmt.Println("Url\t https://github.com/genshen/pkg")
	return nil
}
