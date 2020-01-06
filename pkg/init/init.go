package init

import (
	"errors"
	"flag"
	"github.com/genshen/cmds"
	"github.com/genshen/pkg"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

var initCommand = &cmds.Command{
	Name:        "init",
	Summary:     "init package dependency file: " + pkg.PkgFileName,
	Description: "import and extract dependency packages from tarball file (.tar) specified by a file path",
	CustomFlags: false,
	HasOptions:  true,
}

func init() {
	var pwd string
	var err error
	if pwd, err = os.Getwd(); err != nil {
		pwd = "./"
	}

	var initial initialization
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	initCommand.FlagSet = fs
	initCommand.FlagSet.StringVar(&initial.home, "home", pwd, "path of home directory")
	// initCommand.FlagSet.StringVar(&output, "o", "", "output directory")
	initCommand.FlagSet.Usage = initCommand.Usage // use default usage provided by cmds.Command.
	initCommand.Runner = &initial
	cmds.AllCommands = append(cmds.AllCommands, initCommand)
}

type initialization struct {
	home string
}

func (i *initialization) PreRun() error {
	if i.home == "" {
		return errors.New("flag home is required")
	}
	return nil
}

func (i *initialization) Run() error {
	pkgFilePath := filepath.Join(i.home, pkg.PkgFileName)
	if pkgFile, err := os.Create(pkgFilePath); err != nil {
		return err
	} else {
		var pkgExample pkg.YamlPkg
		pkgExample.Version = 2
		pkgExample.PkgName = "github.com/foo/bar"
		pkgExample.Args = make(map[string]string)
		if pkgExampleBytes, err := yaml.Marshal(&pkgExample); err != nil {
			return err
		} else {
			if _, err := pkgFile.Write(pkgExampleBytes); err != nil {
				return err
			}
		}
	}
	return nil
}
