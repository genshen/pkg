package install

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"github.com/genshen/cmds"
	"github.com/genshen/pkg"
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
	var pwd string
	var err error
	if pwd, err = os.Getwd(); err != nil {
		pwd = "./"
	}

	var cmd install
	fs := flag.NewFlagSet("install", flag.ContinueOnError)
	buildCommand.FlagSet = fs
	buildCommand.FlagSet.StringVar(&cmd.PkgHome, "p", pwd, "absolute or relative path for pkg home.")
	buildCommand.FlagSet.StringVar(&cmd.PkgName, "pkg", "", "install a specific package, default is all packages.")
	buildCommand.FlagSet.BoolVar(&cmd.Skipdep, "skipdep", false,
		"skip to build & install dependency packages, only used in installing a specific package. ")

	buildCommand.FlagSet.Usage = buildCommand.Usage // use default usage provided by cmds.Command.
	buildCommand.Runner = &cmd
	cmds.AllCommands = append(cmds.AllCommands, buildCommand)
}

type install struct {
	PkgHome string
	PkgName string
	Skipdep bool
	DepTree pkg.DependencyTree
}

func (b *install) PreRun() error {
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
	if err := pkg.DepTreeRecover(&b.DepTree, pkgSumPath); err != nil {
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
	// besides, you can also just use source code in your project (e.g. use cmake package in cmake project).
	var options = struct {
		DepTree  *pkg.DependencyTree
		SkipDeps bool
		root     bool
	}{&b.DepTree, false, true}

	if b.PkgName != "" { // build a specific package, not all packages.
		// travel the tree to find the package.
		// todo check loop dependency.
		var depTree *pkg.DependencyTree = nil
		b.DepTree.Traversal(func(tree *pkg.DependencyTree) bool {
			if tree.Context.PackageName == b.PkgName {
				depTree = tree // save the matched tree node.
				return false
			}
			return true
		})
		if depTree == nil {
			return errors.New(fmt.Sprintf("package %s not found", b.PkgName))
		}
		options.DepTree = depTree
		options.SkipDeps = b.Skipdep
		options.root = false
	} else {
		b.DepTree.DlStatus = pkg.DlStatusEmpty // set DlStatusEmpty to skip root package.
	}
	pkgBuiltSet := make(map[string]bool)
	if err := buildPkg(options.DepTree, b.PkgHome, options.root, options.SkipDeps, &pkgBuiltSet); err != nil {
		return err
	}

	return nil
}

// pkgHome is always pkg root.
func createPkgDepCmake(pkgHome, srcHome string, depTree *pkg.DependencyTree) error {
	// create dep cmake file only for pkg based project.
	if !depTree.IsPkgPackage {
		return nil
	}

	// create cmake dep file for this package.
	if cmakeDepWriter, err := os.Create(filepath.Join(srcHome, pkg.CMakeDep)); err != nil {
		return err
	} else {
		pkgCMakeLibSet := make(map[string]bool)
		defer cmakeDepWriter.Close()
		bufWriter := bufio.NewWriter(cmakeDepWriter)

		// for all package, set @PkgHome/vendor as vendor home.
		bufWriter.WriteString(strings.Replace(PkgCMakeHeader, VendorPathReplace, pkg.GetVendorPath(pkgHome), -1))
		if err := cmakeLib(depTree, pkgHome, true, &pkgCMakeLibSet, bufWriter); err != nil {
			return err
		}
		bufWriter.Flush()
		log.Println("generated cmake for package", depTree.Context.PackageName)
	}
	// create cmake dep file for all its sub/child package.
	for _, v := range depTree.Dependencies {
		// for all non-root package, the srcHome is pkgHome/vendor/src/@packageName
		err := createPkgDepCmake(pkgHome, pkg.GetPackageSrcPath(pkgHome, v.Context.PackageName), v)
		if err != nil {
			return err // break loop.
		}
	}
	return nil
}
