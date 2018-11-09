package install

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/genshen/cmds"
	"github.com/genshen/pkg/utils"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	DlStatusEmpty = iota
	DlStatusSkip
	DlStatusOk
)

var getCommand = &cmds.Command{
	Name:        "install",
	Summary:     "install packages from existed file pkg.json",
	Description: "install packages(zip,cmake,makefile,.etc format) existed file pkg.json.",
	CustomFlags: false,
	HasOptions:  true,
}

func init() {
	var pkgHome, pwd string
	var err error
	if pwd, err = os.Getwd(); err != nil {
		pwd = "./"
	}

	fs := flag.NewFlagSet("install", flag.ContinueOnError)
	getCommand.FlagSet = fs
	getCommand.FlagSet.StringVar(&pkgHome, "p", pwd, "absolute path for file pkg.json")
	getCommand.FlagSet.Usage = getCommand.Usage // use default usage provided by cmds.Command.
	getCommand.Runner = &get{PkgHome: pkgHome}
	cmds.AllCommands = append(cmds.AllCommands, getCommand)
}

type get struct {
	PkgHome string // the absolute path of root 'pkg.json'
	DepTree DependencyTree
}

type DependencyTree struct {
	Context      DepPkgContext
	Dependencies []*DependencyTree
	Builder      []string // outer builder (lib used by others)
	SelfBuild    []string // inner builder (shows how this package is built)
	CMakeLib     string   // outer cmake script to add this lib.
	SelfCMakeLib string   // inner cmake script to add this lib.
	DlStatus     int
	IsPkgPackage bool
}

type DepPkgContext struct {
	Override         bool
	CMakeLibOverride bool
	PackageName      string
	SrcPath          string
}

func (get *get) PreRun() error {
	jsonPath := filepath.Join(get.PkgHome, utils.PkgFileName)
	// check pkg.json file existence.
	if fileInfo, err := os.Stat(jsonPath); err != nil {
		return err
	} else if fileInfo.IsDir() {
		return fmt.Errorf("%s is not a file", utils.PkgFileName)
	}

	return nil
	// check .vendor and some related directory, if not exists, create it.
	// return utils.CheckVendorPath(pkgFilePath)
}

func (get *get) Run() error {
	// build pkg.json and download source code (json file must exists).
	if err := get.installSubDependency(get.PkgHome, &get.DepTree); err != nil {
		return err
	}
	// generating cmake script to include dependency libs.
	// the generated cmake file is stored at where pkg command runs.
	// for root package, its srcHome equals to PkgHome.
	if err := createPkgDepCmake(get.PkgHome, get.PkgHome, &get.DepTree); err != nil {
		return err
	}

	// compile and install the source code.
	// besides, you can just use source code in your project (e.g. use cmake package in cmake project).
	get.DepTree.DlStatus = DlStatusEmpty
	pkgBuiltSet := make(map[string]bool)
	if err := buildPkg(&get.DepTree, get.PkgHome, true, &pkgBuiltSet); err != nil {
		return err
	}
	return nil
}

// install dependency in a dependency, installPath is the path of sub-dependency(pkg file location).
// todo circle detect
func (get *get) installSubDependency(installPath string, depTree *DependencyTree) error {
	if pkgJsonPath, err := os.Open(filepath.Join(installPath, utils.PkgFileName)); err == nil { // pkg.json exists.
		defer pkgJsonPath.Close()
		if bytes, err := ioutil.ReadAll(pkgJsonPath); err != nil { // read file contents
			return err
		} else {
			pkgs := utils.Pkg{}
			if err := json.Unmarshal(bytes, &pkgs); err != nil { // unmarshal json to struct
				return err
			}
			// add to build this package.
			// only all its dependency packages are downloaded, can this package be built.
			if build, ok := pkgs.Build[runtime.GOOS]; ok {
				depTree.SelfBuild = build[:]
			}
			depTree.SelfCMakeLib = pkgs.CMakeLib // add cmake include script for this lib
			depTree.IsPkgPackage = true
			// download packages source of direct dependencies.
			if deps, err := get.dlSrc(get.PkgHome, &pkgs.Packages); err == nil {
				depTree.Dependencies = deps
				// install sub dependencies for this package.
				for _, dep := range deps {
					if err := get.installSubDependency(dep.Context.SrcPath, dep); err != nil {
						return err
					}
				}
				return nil
			} else {
				return err
			}
		}
	} else {
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	}
}

//
// download a package source to destination refer to installPath, including source code and installed files.
// usually src files are located at 'vendor/src/PackageName/', installed files are located at 'vendor/pkg/PackageName/'.
// pkgHome: pkgHome is where the file pkg.json is located.
func (get *get) dlSrc(pkgHome string, packages *utils.Packages) ([]*DependencyTree, error) {
	var deps []*DependencyTree
	// todo packages have dependencies.
	// todo check install.
	// download archive src package.
	for key, pkg := range packages.ArchivePackages {
		if err := archiveSrc(pkgHome, key, pkg.Path); err != nil {
			// todo rollback, clean src.
			return nil, err
		} else {
			// if source code downloading succeed, then compile and install it;
			// besides, you can also use source code in your project (e.g. use cmake package in cmake project).
		}
	}
	// download files src, and add it to build tree.
	for key, pkg := range packages.FilesPackages {
		srcDes := utils.GetPackageSrcPath(pkgHome, key)
		status := DlStatusEmpty
		if _, err := os.Stat(srcDes); os.IsNotExist(err) {
			if err := filesSrc(srcDes, key, pkg.Path, pkg.Files); err != nil {
				// todo rollback, clean src.
				return nil, err
			}
			status = DlStatusOk
		} else if err != nil {
			return nil, err
		} else {
			status = DlStatusSkip
			log.Printf("skiped downloading %s in %s, because it already exists.\n", key, srcDes)
		}
		// add to dependency tree.
		dep := DependencyTree{
			Builder:  pkg.Package.Build[:],
			DlStatus: status,
			CMakeLib: pkg.CMakeLib,
			Context: DepPkgContext{
				Override:         pkg.Override,
				CMakeLibOverride: pkg.CMakeLibOverride,
				SrcPath:          srcDes,
				PackageName:      key,
			},
		}
		deps = append(deps, &dep)
	}
	// download git src, and add it to build tree.
	for key, pkg := range packages.GitPackages {
		srcDes := utils.GetPackageSrcPath(pkgHome, key)
		status := DlStatusEmpty
		// check directory, if not exists, then create it.
		if _, err := os.Stat(srcDes); os.IsNotExist(err) {
			if err := gitSrc(srcDes, key, pkg.Path, pkg.Hash, pkg.Branch, pkg.Tag); err != nil {
				// todo rollback, clean src.
				return nil, err
			}
			status = DlStatusOk
		} else if err != nil {
			return nil, err
		} else {
			status = DlStatusSkip
			log.Printf("skiped downloading %s in %s, because it already exists.\n", key, srcDes)
		}
		// add to dependency tree.
		dep := DependencyTree{
			Builder:  pkg.Package.Build[:],
			DlStatus: status,
			CMakeLib: pkg.CMakeLib,
			Context: DepPkgContext{
				Override:         pkg.Override,
				CMakeLibOverride: pkg.CMakeLibOverride,
				SrcPath:          srcDes,
				PackageName:      key,
			},
		}
		deps = append(deps, &dep)
	}
	return deps, nil
}

func createPkgDepCmake(pkgHome, srcHome string, depTree *DependencyTree) error {
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
