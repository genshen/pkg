package list

import (
	"errors"
	"flag"
	"fmt"
	"github.com/genshen/cmds"
	"github.com/genshen/pkg"
	"os"
	"sort"
)

var listCommand = &cmds.Command{
	Name:        "list",
	Summary:     "list all packages",
	Description: "list all dependency packages",
	CustomFlags: false,
	HasOptions:  true,
}

func init() {
	var pwd string
	var err error
	if pwd, err = os.Getwd(); err != nil {
		pwd = "./"
	}

	var packList list
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	listCommand.FlagSet = fs
	listCommand.FlagSet.StringVar(&packList.home, "home", pwd, "path of home directory")
	listCommand.FlagSet.BoolVar(&packList.listAll, "a", false, "list all packages, including indirect packages")
	listCommand.FlagSet.BoolVar(&packList.listLong, "l", false, "list long packages, including package source path")
	listCommand.FlagSet.Usage = listCommand.Usage // use default usage provided by cmds.Command.
	listCommand.Runner = &packList
	cmds.AllCommands = append(cmds.AllCommands, listCommand)
}

type list struct {
	home     string
	listAll  bool // flag of listing all packages, including indirect packages. Not used now.
	listLong bool // flag of listing long packages, including package source path
	Metas    map[string]pkg.PackageMeta
}

func (l *list) PreRun() error {
	if l.home == "" {
		return errors.New("flag home is required")
	}

	// check sum file
	pkgSumPath := pkg.GetPkgSumPath(l.home)
	if fileInfo, err := os.Stat(pkgSumPath); err != nil {
		return fmt.Errorf(`stat file %s failed, make sure you have run "pkg install"; error: %s`, pkgSumPath, err)
	} else if fileInfo.IsDir() {
		return fmt.Errorf("%s is not a file", pkg.PkgFileName)
	}
	// check vendor files
	includeDir := pkg.GetIncludePath(l.home)
	if err := pkg.CheckDir(includeDir); err != nil { // check include dir exist.
		return err
	}

	// resolve sum file.
	if err := pkg.DepTreeRecover(&l.Metas, pkgSumPath); err != nil {
		return err
	}
	return nil
}

func (l *list) Run() error {
	pkgList := []string{}
	for _, meta := range l.Metas {
		if meta.PackageName == "root" {
			continue
		}

		if l.listLong { // long info
			pkgList = append(pkgList, fmt.Sprintf("%s@%s src=%s", meta.PackageName, meta.Version, meta.VendorSrcPath("")))
		} else { // short info
			pkgList = append(pkgList, fmt.Sprintf("%s@%s", meta.PackageName, meta.Version))
		}
	}

	sort.Strings(pkgList)

	for _, p := range pkgList {
		fmt.Println(p)
	}
	return nil
}
