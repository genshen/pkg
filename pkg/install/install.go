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
	buildCommand.FlagSet.BoolVar(&cmd.self, "self", false, "only build the package specified by `pkg` option(not build dependency packages)")
	buildCommand.FlagSet.StringVar(&cmd.cmakeConfigArg, "cmake-conf-arg", "", "arguments used in cmake configuration step.")
	buildCommand.FlagSet.StringVar(&cmd.cmakeBuildArg, "cmake-build-arg", "", "arguments used in cmake building step.")
	buildCommand.FlagSet.BoolVar(&cmd.verbose, "verbose", false, "show building logs while installing package(s).")

	buildCommand.FlagSet.Usage = buildCommand.Usage // use default usage provided by cmds.Command.
	buildCommand.Runner = &cmd
	cmds.AllCommands = append(cmds.AllCommands, buildCommand)
}

type install struct {
	PkgHome        string
	PkgName        string
	sh             bool   // generate shell script for building packages(sh)
	self           bool   // not build build dependency packages.
	verbose        bool   // log the building log (verbose)
	cmakeConfigArg string // config argument while installation
	cmakeBuildArg  string // build argument while installation
	Metas          map[string]pkg.PackageMeta
}

func (b *install) PreRun() error {
	if b.PkgHome == "" {
		return errors.New("flag p is required")
	}
	// check sum file
	pkgSumPath := pkg.GetPkgSumPath(b.PkgHome)
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
		lists []string
		Metas map[string]pkg.PackageMeta
	}{nil, b.Metas}

	if b.PkgName != "" { // build a specific package, not all packages.
		if b.self { // only build one package
			options.lists = make([]string, 0, 1)
			options.lists = append(options.lists, b.PkgName)
		} else { // also build its dependencies.
			if pkgLists, err := pkg.LoadListFromGraph(pkg.GetDepGraphPath(b.PkgHome), b.PkgName); err != nil {
				return err
			} else {
				options.lists = pkgLists
				options.lists = append(options.lists, b.PkgName)
			}
		}
	} else {
		// set default building packages if PkgName is not specified.
		b.PkgName = pkg.RootPKG
		if pkgLists, err := pkg.LoadListFromGraph(pkg.GetDepGraphPath(b.PkgHome), b.PkgName); err != nil {
			return err
		} else {
			options.lists = pkgLists
		}
	}

	if b.sh {
		if shellFile, err := os.OpenFile(pkg.GetPkgBuildPath(b.PkgHome), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755); err != nil {
			return err
		} else {
			buffWriter := bufio.NewWriter(shellFile)
			defer buffWriter.Flush()
			shWriter, err := NewInsShellWriter(b.PkgHome, buffWriter, b.cmakeConfigArg, b.cmakeBuildArg)
			if err != nil {
				return err
			}
			if err := buildPkg(shWriter, options.lists, options.Metas); err != nil {
				return err
			}

			log.Info("pkg building shell script generated at ", pkg.GetPkgBuildPath(b.PkgHome))
		}
	} else {
		var insExe = NewInsExecutor(b.PkgHome, b.verbose, b.cmakeConfigArg, b.cmakeBuildArg)
		if err := buildPkg(insExe, options.lists, options.Metas); err != nil {
			return err
		}
		log.Info("all packages installed successfully.")
	}
	return nil
}
