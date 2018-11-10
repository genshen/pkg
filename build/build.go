package install

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/genshen/cmds"
	"github.com/genshen/pkg/utils"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var buildCommand = &cmds.Command{
	Name:        "install",
	Summary:     "compile and install dependency packages",
	Description: "compile dependency packages source code and install it.",
	CustomFlags: false,
	HasOptions:  true,
}

func init() {
	var pkgHome, pwd string
	var err error
	if pwd, err = os.Getwd(); err != nil {
		pwd = "./"
	}

	buildCommand.Runner = &install{}
	fs := flag.NewFlagSet("install", flag.ContinueOnError)
	buildCommand.FlagSet = fs
	buildCommand.FlagSet.StringVar(&pkgHome, "p", pwd, "absolute path for pkg home.")
	buildCommand.FlagSet.Usage = buildCommand.Usage // use default usage provided by cmds.Command.
	buildCommand.Runner = &install{PkgHome: pkgHome}
	cmds.AllCommands = append(cmds.AllCommands, buildCommand)
}

type install struct {
	PkgHome string
	DepTree utils.DependencyTree
}

func (b *install) PreRun() error {
	// check sum file
	pkgSumPath := filepath.Join(b.PkgHome, utils.PkgSumFileName)
	if fileInfo, err := os.Stat(pkgSumPath); err != nil {
		return fmt.Errorf(`stat file %s failed, make sure you have run "pkg install"; error: %s`, pkgSumPath, err)
	} else if fileInfo.IsDir() {
		return fmt.Errorf("%s is not a file", utils.PkgFileName)
	}
	// resolve sum file.
	if err := utils.DepTreeRecover(&b.DepTree, pkgSumPath); err != nil {
		return err
	}
	return nil
}

func (b *install) Run() error {
	// generating cmake script to include dependency libs.
	// the generated cmake file is stored at where pkg command runs.
	// for root package, its srcHome equals to PkgHome.
	if err := createPkgDepCmake(b.PkgHome, b.PkgHome, &b.DepTree); err != nil {
		return err
	}

	// compile and install the source code.
	// besides, you can just use source code in your project (e.g. use cmake package in cmake project).
	b.DepTree.DlStatus = utils.DlStatusEmpty
	pkgBuiltSet := make(map[string]bool)
	if err := buildPkg(&b.DepTree, b.PkgHome, true, &pkgBuiltSet); err != nil {
		return err
	}

	return nil
}

func createPkgDepCmake(pkgHome, srcHome string, depTree *utils.DependencyTree) error {
	// build dep cmake file only for pkg based project.
	if !depTree.IsPkgPackage {
		return nil
	}

	// create cmake dep file for this package.
	if cmakeDepWriter, err := os.Create(filepath.Join(srcHome, utils.CMakeDep)); err != nil {
		return err
	} else {
		pkgCMakeLibSet := make(map[string]bool)
		defer cmakeDepWriter.Close()
		bufWriter := bufio.NewWriter(cmakeDepWriter)

		// for all package, set @PkgHome/vendor as vendor home.
		bufWriter.WriteString(strings.Replace(PkgCMakeHeader, VendorPathReplace, utils.GetVendorPath(pkgHome), -1))
		if err := cmakeLib(depTree, pkgHome, true, &pkgCMakeLibSet, bufWriter); err != nil {
			return err
		}
		bufWriter.Flush()
		log.Println("generated cmake for package", depTree.Context.PackageName)
	}
	// create cmake dep file for all its sub/child package.
	for _, v := range depTree.Dependencies {
		// for all non-root package, the srcHome is pkgHome/vendor/src/@packageName
		err := createPkgDepCmake(pkgHome, utils.GetPackageSrcPath(pkgHome, v.Context.PackageName), v)
		if err != nil {
			return err // break loop.
		}
	}
	return nil
}
