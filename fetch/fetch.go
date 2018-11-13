package install

import (
	"flag"
	"fmt"
	"github.com/genshen/cmds"
	"github.com/genshen/pkg/utils"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
)

var getCommand = &cmds.Command{
	Name:        "fetch",
	Summary:     "fetch packages from remote based an existed file " + utils.PkgFileName,
	Description: "fetch packages(zip,cmake,makefile,.etc format) existed file " + utils.PkgFileName + ".",
	CustomFlags: false,
	HasOptions:  true,
}

func init() {
	var pkgHome, pwd string
	//var absRoot bool
	var err error
	if pwd, err = os.Getwd(); err != nil {
		pwd = "./"
	}

	fs := flag.NewFlagSet("fetch", flag.ContinueOnError)
	getCommand.FlagSet = fs
	//getCommand.FlagSet.BoolVar(&absRoot, "abspath", false, "use absolute path, not relative path")
	getCommand.FlagSet.StringVar(&pkgHome, "p", pwd, "absolute or relative path for file "+utils.PkgFileName)
	// todo make pkgHome abs path anyway.
	getCommand.FlagSet.Usage = getCommand.Usage // use default usage provided by cmds.Command.
	getCommand.Runner = &fetch{PkgHome: pkgHome}
	cmds.AllCommands = append(cmds.AllCommands, getCommand)
}

type fetch struct {
	PkgHome string // the absolute path of root 'pkg.json' form command path.
	DepTree utils.DependencyTree
}

func (get *fetch) PreRun() error {
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

func (get *fetch) Run() error {
	// build pkg.json and download source code (json file must exists).
	if err := get.installSubDependency(get.PkgHome, &get.DepTree); err != nil {
		return err
	}
	// dump dependency tree to file system
	log.WithFields(log.Fields{
		"file": utils.PkgSumFileName,
	}).Info("saving dependencies tree to file.")
	if err := get.DepTree.Dump(utils.PkgSumFileName); err == nil {
		log.WithFields(log.Fields{
			"file": utils.PkgSumFileName,
		}).Info("saved dependencies tree to file.")
	} else {
		return err
	}
	log.Info("fetch succeeded.")
	return nil
}

// install dependency in a dependency, installPath is the path of sub-dependency(pkg file location).
// todo circle detect
func (get *fetch) installSubDependency(installPath string, depTree *utils.DependencyTree) error {
	if pkgJsonPath, err := os.Open(filepath.Join(installPath, utils.PkgFileName)); err == nil { // pkg.json exists.
		defer pkgJsonPath.Close()
		if bytes, err := ioutil.ReadAll(pkgJsonPath); err != nil { // read file contents
			return err
		} else {
			pkgs := utils.Pkg{}
			if err := yaml.Unmarshal(bytes, &pkgs); err != nil { // unmarshal yaml to struct
				return err
			}
			d, err := yaml.Marshal(&pkgs)
			if err != nil {
				log.Fatalf("error: %v", err)
			}
			fmt.Printf("--- m dump:\n%s\n\n", string(d))

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
func (get *fetch) dlSrc(pkgHome string, packages *utils.Packages) ([]*utils.DependencyTree, error) {
	var deps []*utils.DependencyTree
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
		status := utils.DlStatusEmpty
		if _, err := os.Stat(srcDes); os.IsNotExist(err) {
			if err := filesSrc(srcDes, key, pkg.Path, pkg.Files); err != nil {
				// todo rollback, clean src.
				return nil, err
			}
			status = utils.DlStatusOk
		} else if err != nil {
			return nil, err
		} else {
			status = utils.DlStatusSkip
			log.WithFields(log.Fields{
				"pkg":      key,
				"src_path": srcDes,
			}).Info("skipped fetching package, because it already exists.")
		}
		// add to dependency tree.
		dep := utils.DependencyTree{
			Builder:  pkg.Package.Build[:],
			DlStatus: status,
			CMakeLib: pkg.CMakeLib,
			Context: utils.DepPkgContext{
				SrcPath:          utils.GetPackageSrcPath("", key), // make it relative path.
				PackageName:      key,
			},
		}
		deps = append(deps, &dep)
	}
	// download git src, and add it to build tree.
	for key, pkg := range packages.GitPackages {
		srcDes := utils.GetPackageSrcPath(pkgHome, key)
		status := utils.DlStatusEmpty
		// check directory, if not exists, then create it.
		if _, err := os.Stat(srcDes); os.IsNotExist(err) {
			if err := gitSrc(srcDes, key, pkg.Path, pkg.Hash, pkg.Branch, pkg.Tag); err != nil {
				// todo rollback, clean src.
				return nil, err
			}
			status = utils.DlStatusOk
		} else if err != nil {
			return nil, err
		} else {
			status = utils.DlStatusSkip
			log.WithFields(log.Fields{
				"pkg":      key,
				"src_path": srcDes,
			}).Info("skipped fetching package, because it already exists.")
		}
		// add to dependency tree.
		dep := utils.DependencyTree{
			Builder:  pkg.Package.Build[:],
			DlStatus: status,
			CMakeLib: pkg.CMakeLib,
			Context: utils.DepPkgContext{
				SrcPath:          utils.GetPackageSrcPath("", key), // make it relative path.
				PackageName:      key,
			},
		}
		deps = append(deps, &dep)
	}
	return deps, nil
}
