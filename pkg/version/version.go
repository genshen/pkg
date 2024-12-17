package version

import (
	"flag"
	"fmt"
	"github.com/genshen/cmds"
	"github.com/genshen/pkg"
)

var versionCommand = &cmds.Command{
	Name:        "version",
	Summary:     "print pkg version",
	Description: "print current pkg version.",
	CustomFlags: false,
	HasOptions:  false,
}

var (
	GitCommitID string = "unknown"
	BuildTime 	string = ""
)

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
	fmt.Printf("version\t %s\n", pkg.VERSION)
	fmt.Printf("Commit \t %s\n", GitCommitID)
	fmt.Println("Author\t genshenchu@gmail.com")
	fmt.Println("Url\t https://github.com/genshen/pkg")
	fmt.Printf("Build time \t %s\n", BuildTime)
	return nil
}
