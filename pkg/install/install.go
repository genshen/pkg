package install

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"github.com/genshen/cmds"
	"github.com/genshen/pkg"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

var buildCommand = &cmds.Command{
	Name:        "install",
	Summary:     "compile and install dependency packages",
	Description: "compile dependency packages source code and install it.",
	CustomFlags: false,
	HasOptions:  true,
}

func init() {
	var pwd string
	var err error
	if pwd, err = os.Getwd(); err != nil {
		pwd = "./"
	}

	var cmd install
	fs := flag.NewFlagSet("install", flag.ExitOnError)
	buildCommand.FlagSet = fs
	buildCommand.FlagSet.StringVar(&cmd.PkgHome, "p", pwd, "absolute or relative path for pkg home.")
	buildCommand.FlagSet.StringVar(&cmd.PkgName, "pkg", "", "install a specific package, default is all packages.")
	buildCommand.FlagSet.BoolVar(&cmd.sh, "sh", false, "skip building, but generate shell script for building packages.")
	buildCommand.FlagSet.BoolVar(&cmd.verbose, "verbose", false, "show building logs while installing package(s).")

	buildCommand.FlagSet.Usage = buildCommand.Usage // use default usage provided by cmds.Command.
	buildCommand.Runner = &cmd
	cmds.AllCommands = append(cmds.AllCommands, buildCommand)
}

type install struct {
	PkgHome string
	PkgName string
	sh      bool // generate shell script for building packages(sh)
	verbose bool // log the building log (verbose)
	Metas   []pkg.PackageMeta
}

func (b *install) PreRun() error {
	if b.PkgHome == "" {
		return errors.New("flag p is required")
	}
	// check sum file
	pkgSumPath := filepath.Join(b.PkgHome, pkg.PkgSumFileName)
	if fileInfo, err := os.Stat(pkgSumPath); err != nil {
		return fmt.Errorf(`stat file %s failed, make sure you have run "pkg install"; error: %s`, pkgSumPath, err)
	} else if fileInfo.IsDir() {
		return fmt.Errorf("%s is not a file", pkg.PkgFileName)
	}
	// check vendor files
	includeDir := pkg.GetIncludePath(b.PkgHome)
	if err := pkg.CheckDir(includeDir); err != nil { // check include dir exist.
		return err
	}

	// resolve sum file.
	if err := pkg.DepTreeRecover(&b.Metas, pkgSumPath); err != nil {
		return err
	}
	return nil
}

func (b *install) Run() error {
	// compile and install the source code.
	// besides, you can also just use source code in your project (e.g. use cmake package in cmake project).
	var options = struct {
		Metas []pkg.PackageMeta
		root  bool
	}{nil, true}

	if b.PkgName != "" { // build a specific package, not all packages.
		// travel the tree to find the package.
		var found = false
		for _, v := range b.Metas {
			if v.PackageName == b.PkgName {
				// save the matched tree node.
				found = true
				options.Metas = make([]pkg.PackageMeta, 1)
				options.Metas = append(options.Metas, v)
				break
			}
		}
		if !found {
			return errors.New(fmt.Sprintf("package %s not found", b.PkgName))
		}
		options.root = false
	} else {
		options.Metas = b.Metas
		// remove root package building.
		for i, v := range options.Metas {
			if v.PackageName == "" {
				options.Metas = append(options.Metas[:i], options.Metas[i+1:]...)
				break
			}
		}
	}

	if b.sh {
		if shellFile, err := os.OpenFile(pkg.GetPkgBuildPath(b.PkgHome), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755); err != nil {
			return err
		} else {
			buffWriter := bufio.NewWriter(shellFile)
			defer buffWriter.Flush()
			if err := generateShell(buffWriter, options.Metas, b.PkgHome); err != nil {
				return err
			}

			log.Info("pkg building shell script generated at ", pkg.GetPkgBuildPath(b.PkgHome))
		}
	} else {
		if err := buildPkg(options.Metas, b.PkgHome, b.verbose); err != nil {
			return err
		}
		log.Info("all packages installed successfully.")
	}
	return nil
}
