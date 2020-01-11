package install

import (
	"errors"
	"flag"
	"fmt"
	"github.com/genshen/cmds"
	"github.com/genshen/pkg"
	"github.com/genshen/pkg/conf"
	"github.com/otiai10/copy"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path/filepath"
)

var fetchCommand = &cmds.Command{
	Name:        "fetch",
	Summary:     "fetch packages from remote based an existed file " + pkg.PkgFileName,
	Description: "fetch packages(zip,cmake,makefile,.etc format) existed file " + pkg.PkgFileName + ".",
	CustomFlags: false,
	HasOptions:  true,
}

func init() {
	var pwd string
	//var absRoot bool
	var err error
	if pwd, err = os.Getwd(); err != nil {
		pwd = "./"
	}

	var f fetch
	fs := flag.NewFlagSet("fetch", flag.ExitOnError)
	fetchCommand.FlagSet = fs
	//fetchCommand.FlagSet.BoolVar(&absRoot, "abspath", false, "use absolute path, not relative path")
	fetchCommand.FlagSet.StringVar(&f.PkgHome, "p", pwd, "absolute or relative path for file "+pkg.PkgFileName)
	// todo make pkgHome abs path anyway.
	fetchCommand.FlagSet.Usage = fetchCommand.Usage // use default usage provided by cmds.Command.
	fetchCommand.Runner = &f
	cmds.AllCommands = append(cmds.AllCommands, fetchCommand)
}

type fetch struct {
	PkgHome string // the absolute path of root 'pkg.yaml' form command path.
	DepTree pkg.DependencyTree
	Auth    map[string]conf.Auth
}

func (f *fetch) PreRun() error {
	if f.PkgHome == "" {
		return errors.New("flag p is required")
	}

	pkgFilePath := filepath.Join(f.PkgHome, pkg.PkgFileName)
	// check pkg.yaml file existence.
	if fileInfo, err := os.Stat(pkgFilePath); err != nil {
		return err
	} else if fileInfo.IsDir() {
		return fmt.Errorf("%s is not a file", pkg.PkgFileName)
	}

	// check vendor dir
	vendorDir := pkg.GetVendorPath(f.PkgHome)
	if err := pkg.CheckDir(vendorDir); err != nil {
		return err
	}

	//parse git clone auth file.
	if config, err := conf.ParseConfig(f.PkgHome); err != nil {
		return err
	} else {
		f.Auth = config.Auth
	}

	return nil
	// check .vendor and some related directory, if not exists, create it.
	// return pkg.CheckVendorPath(pkgFilePath)
}

func (f *fetch) Run() error {
	// build pkg.yaml and download source code (yaml file must exists).
	pkgSrcDir, err := pkg.GetHomeSrcPath()
	if err != nil {
		return err
	}
	// fetch packages to user home directory.
	log.Info("packages will be downloaded to directory ", pkgSrcDir)
	pkgLock := make(map[string]string)
	if err := f.fetchSubDependency(pkg.RootPKG, f.PkgHome, &pkgLock, &f.DepTree); err != nil {
		return err
	}

	// dump dependency tree to file system
	if err := f.DepTree.Dump(pkg.GetPkgSumPath(f.PkgHome)); err != nil { //fixme
		return err
	} else {
		log.WithFields(log.Fields{
			"file": pkg.PkgSumFileName,
		}).Info("saved dependencies tree to file.")
	}

	// dump all packages's dependencies.
	if file, err := os.Create(pkg.GetDepGraphPath(f.PkgHome)); err != nil {
		return err
	} else {
		defer file.Close()
		if err := f.DepTree.MarshalGraph(file); err != nil {
			return err
		}
	}

	// generating cmake script to include dependency libs.
	// the generated cmake file is stored at where pkg command runs.
	// for project package, its srcHome equals to PkgHome.
	if err := createPkgDepCmake(f.PkgHome, &f.DepTree); err != nil {
		return err
	}

	log.Info("fetch succeeded.")
	return nil
}

// install dependencies to a directory, installPath is the root path of sub-dependency(always be the project root).
// pkgPath: the given package name/path (e.g github.com/google/googletest) from top level package.
//pkgVendorSrcPath: path of source file directory in vendor.
// todo circle detect
func (f *fetch) fetchSubDependency(pkgPath string, pkgVendorSrcPath string, pkgLock *map[string]string, depTree *pkg.DependencyTree) error {
	// check pkg.yaml file in vendor directory
	if pkgYamlFile, err := os.Open(filepath.Join(pkgVendorSrcPath, pkg.PkgFileName)); err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	} else { // pkg.yaml exists.
		defer pkgYamlFile.Close()
		if bytes, err := ioutil.ReadAll(pkgYamlFile); err != nil { // read file contents
			return err
		} else {
			pkgYaml := pkg.YamlPkg{}
			if err := yaml.Unmarshal(bytes, &pkgYaml); err != nil { // unmarshal yaml to struct
				return err
			}

			if pkgPath == pkg.RootPKG {
				depTree.Context.PackageName = pkg.RootPKG
			} else { // check the package name in its pkg.yaml, then give a warning if it does not match
				if depTree.Context.PackageName != pkgYaml.PkgName {
					log.Warningf("package name does not match in pkg.yaml file(top level package name: %s, package name in pkg.yaml: %s).",
						depTree.Context.PackageName,
						pkgYaml.PkgName,
					)
				}
			}

			// add to build this package.
			// only all its dependency packages are downloaded, can this package be built.
			builder := pkgYaml.FindBuilder()
			// overwrite the default builder command and cmake lib,
			// if they are specified here.
			if !(builder == nil || len(builder) == 0) || pkgYaml.CMakeLib != "" {
				depTree.Context.SelfBuild = builder
				depTree.Context.SelfCMakeLib = pkgYaml.CMakeLib // add cmake include script for this lib
			}

			depTree.IsPkgPackage = true
			if depTree.Dependencies == nil {
				depTree.Dependencies = make([]*pkg.DependencyTree, 0)
			}

			// migrate package based on pkg.yaml v1 to v2
			if err := pkgYaml.Packages.MigrateToV2(&pkgYaml.Deps); err != nil {
				return err
			}

			// download git based packages source of direct dependencies.
			if deps, err := f.dlPackagesDepSrc(pkgLock, gitPkgsToInterface(pkgYaml.Deps.GitPackages)); err != nil {
				return err
			} else {
				// copy downloaded packages to vendor directory
				if err := f.copyPkgToVendor(deps); err != nil {
					return err
				}
				// add and install sub dependencies for this package.
				depTree.Dependencies = append(depTree.Dependencies, deps...)
				for _, dep := range deps {
					if err := f.fetchSubDependency(dep.Context.PackageName, dep.Context.VendorSrcPath(f.PkgHome), pkgLock, dep); err != nil {
						return err
					}
				}
			}
			// download file based packages
			if deps, err := f.dlPackagesDepSrc(pkgLock, filesPkgsToInterface(pkgYaml.Deps.FilesPackages)); err != nil {
				return err
			} else {
				// copy downloaded packages to vendor directory
				if err := f.copyPkgToVendor(deps); err != nil {
					return err
				}
				depTree.Dependencies = append(depTree.Dependencies, deps...)
			}
		}
	}
	return nil
}

// download a package source to destination refer to installPath, including source code and installed files.
// usually src files are located at 'vendor/src/PackageName/', installed files are located at 'vendor/pkg/PackageName/'.
// pkgHome: project root direction.
func (f *fetch) dlPackagesDepSrc(pkgLock *map[string]string, packages map[string]PackageFetcher) ([]*pkg.DependencyTree, error) {
	var deps []*pkg.DependencyTree
	// todo packages have dependencies.
	// todo check install.

	// download archive src package.
	//for key, archPkg := range packages.ArchivePackages {
	//	if err := archiveSrc(pkgHome, key, archPkg.Path); err != nil {
	//		// todo rollback, clean src.
	//		return nil, err
	//	} else {
	//		// if source code downloading succeed, then compile and install it;
	//		// besides, you can also use source code in your project (e.g. use cmake package in cmake project).
	//	}
	//}

	if packages == nil {
		return deps, nil
	}
	// download git src, and add it to build tree.
	for key, p := range packages {
		var context pkg.PackageMeta
		// before fetching package, set version and package name/path
		if err := p.setPackageMeta(key, &context); err != nil {
			return nil, err
		}
		// set save directory path
		status := pkg.DlStatusEmpty

		// version conflict and deciding
		if ver, ok := (*pkgLock)[context.PackageName]; ok {
			// use the matched version package
			context.Version = ver
			log.WithFields(log.Fields{
				"pkg":     key,
				"version": ver,
			}).Trace("package matches another version.")
		} else {
			// save this version
			(*pkgLock)[context.PackageName] = context.Version
		}

		// src path in (global) user home
		srcDes := context.HomeCacheSrcPath()
		vendorSrcDes := context.VendorSrcPath(f.PkgHome)

		// check target directory to save src files.
		_, errHomeSrc := os.Stat(srcDes)
		_, errVendorSrc := os.Stat(vendorSrcDes)
		// check home cache dir and vendor dir
		if os.IsNotExist(errHomeSrc) && os.IsNotExist(errVendorSrc) {
			log.WithFields(log.Fields{
				"pkg":     context.PackageName,
				"storage": srcDes,
			}).Info("downloading dependencies.")
			if err := p.fetch(f.Auth, srcDes, context); err != nil {
				return nil, err
			}
			status = pkg.DlStatusOk
		} else {
			if errHomeSrc != nil && !os.IsNotExist(errHomeSrc) {
				return nil, errHomeSrc
			} else if errVendorSrc != nil && !os.IsNotExist(errVendorSrc) {
				return nil, errVendorSrc
			}
			status = pkg.DlStatusSkip
			log.WithFields(log.Fields{
				"pkg":      key,
				"src_path": srcDes,
			}).Info("skipped fetching package, because it already exists.")
		}

		// add to dependency tree.
		dep := pkg.DependencyTree{
			DlStatus: status,
			Context:  context,
		}
		deps = append(deps, &dep)
	}
	return deps, nil
}

// cope packages specified by array deps from user home directory to vendor/src directory
func (f *fetch) copyPkgToVendor(deps []*pkg.DependencyTree) error {
	//log.Info("copy dependencies from cache to vendor.")
	for _, dep := range deps {
		pkgCachePath := dep.Context.HomeCacheSrcPath()
		pkgVendorPath := dep.Context.VendorSrcPath(f.PkgHome)

		// copy only when global cache exist
		if _, err := os.Stat(pkgCachePath); os.IsNotExist(err) { // cache not exist
			if _, err := os.Stat(pkgVendorPath); err != nil {
				if os.IsNotExist(err) { // vendor not exist
					return fmt.Errorf("cache and vendor path of package `%s` does not exists", dep.Context.PackageName)
				}
				return err
			} // else: vendor exist
		} else if err != nil {
			return err
		} else {
			if err := copy.Copy(pkgCachePath, pkgVendorPath); err != nil {
				return err
			}
		}
	}
	return nil
}
