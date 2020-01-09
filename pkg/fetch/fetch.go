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
	"runtime"
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
	if err := f.installSubDependency(f.PkgHome, &pkgLock, &f.DepTree); err != nil {
		return err
	}

	// cope packages in user home directory to vendor/src directory
	if err := f.DepTree.TraversalDeep(func(tree *pkg.DependencyTree) error {
		if tree.Context.PackageName == pkg.RootPKG {
			return nil
		} // don't copy root package
		if err := copy.Copy(tree.Context.HomeCacheSrcPath(), tree.Context.VendorSrcPath(f.PkgHome)); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	// dump dependency tree to file system
	if err := f.DepTree.Dump(pkg.GetPkgSumPath(f.PkgHome)); err == nil { //fixme
		log.WithFields(log.Fields{
			"file": pkg.PkgSumFileName,
		}).Info("saved dependencies tree to file.")
	} else {
		return err
	}

	// generating cmake script to include dependency libs.
	// the generated cmake file is stored at where pkg command runs.
	// for project package, its srcHome equals to PkgHome.
	if err := createPkgDepCmake(f.PkgHome, f.PkgHome, true, &f.DepTree); err != nil {
		return err
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
	log.Info("fetch succeeded.")
	return nil
}

// install dependencies to a directory, installPath is the root path of sub-dependency(always be the project root).
// todo circle detect
func (f *fetch) installSubDependency(pkgSrcPath string, pkgLock *map[string]string, depTree *pkg.DependencyTree) error {
	if pkgYamlPath, err := os.Open(filepath.Join(pkgSrcPath, pkg.PkgFileName)); err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	} else { // pkg.yaml exists.
		defer pkgYamlPath.Close()
		if bytes, err := ioutil.ReadAll(pkgYamlPath); err != nil { // read file contents
			return err
		} else {
			pkgYaml := pkg.YamlPkg{}
			if err := yaml.Unmarshal(bytes, &pkgYaml); err != nil { // unmarshal yaml to struct
				return err
			}

			if f.PkgHome == pkgSrcPath {
				depTree.Context.PackageName = pkg.RootPKG
			} else { // check the package name in its pkg.yaml, then give a warning if it does not match
				if depTree.Context.PackageName != pkgYaml.PkgName {
					log.Warningf("package name does not match in pkg.yaml file(top level package name %s, package name in pkg.yaml).",
						depTree.Context.PackageName,
						pkgYaml.PkgName,
					)
				}
			}

			// add to build this package.
			// only all its dependency packages are downloaded, can this package be built.
			if build, ok := pkgYaml.Build[runtime.GOOS]; ok {
				depTree.Context.SelfBuild = build[:]
			}
			depTree.Context.SelfCMakeLib = pkgYaml.CMakeLib // add cmake include script for this lib
			depTree.IsPkgPackage = true
			if depTree.Dependencies == nil {
				depTree.Dependencies = make([]*pkg.DependencyTree, 0)
			}

			// migrate package based on pkg.yaml v1 to v2
			if err := pkgYaml.Packages.MigrateToV2(&pkgYaml.Deps); err != nil {
				return err
			}

			// download git based packages source of direct dependencies.
			if deps, err := f.dlPackagesDepSrc(pkgLock, pkgYaml.Deps.GitPackages); err != nil {
				return err
			} else {
				// add and install sub dependencies for this package.
				depTree.Dependencies = append(depTree.Dependencies, deps...)
				for _, dep := range deps {
					if err := f.installSubDependency(dep.Context.HomeCacheSrcPath(), pkgLock, dep); err != nil {
						return err
					}
				}
			}
			// download files packages
			if deps, err := f.dlFilesDepSrc(pkgLock, pkgYaml.Deps.FilesPackages); err != nil {
				return err
			} else {
				depTree.Dependencies = append(depTree.Dependencies, deps...)
			}
		}
		return nil
	}
}

// download file based packages
func (f *fetch) dlFilesDepSrc(pkgLock *map[string]string, packages map[string]pkg.YamlFilesPackage) ([]*pkg.DependencyTree, error) {
	var deps []*pkg.DependencyTree
	// todo packages have dependencies.
	// todo check install.
	if packages == nil {
		return deps, nil
	}
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
	// download files src, and add it to build tree.
	for key, filePkg := range packages {
		status := pkg.DlStatusEmpty
		version := "latest"
		srcDes, err := pkg.GetCachedPackageSrcPath(key, version)
		if err != nil {
			return nil, err
		}
		if ver, ok := (*pkgLock)[key]; ok {
			log.WithFields(log.Fields{
				"pkg":     key,
				"version": ver,
			}).Trace("package matches another version.")
		} else {
			// log version
			(*pkgLock)[key] = version
		}
		// check target directory to save src.
		if _, err := os.Stat(srcDes); os.IsNotExist(err) {
			if err := filesSrc(srcDes, key, filePkg.Path, filePkg.Files); err != nil {
				_ = os.RemoveAll(srcDes)
				return nil, err
			}
			status = pkg.DlStatusOk
		} else if err != nil {
			return nil, err
		} else {
			status = pkg.DlStatusSkip
			log.WithFields(log.Fields{
				"pkg": key,
			}).Info("skipped fetching package, because it already exists.")
		}

		// add to dependency tree.
		dep := pkg.DependencyTree{
			DlStatus: status,
			Context: pkg.PackageMeta{
				PackageName: key,
				TargetName:  key,
				Version:     version,
				Builder:     filePkg.Build[:],
				CMakeLib:    filePkg.CMakeLib,
			},
		}
		deps = append(deps, &dep)
	}
	return deps, nil
}

// download a package source to destination refer to installPath, including source code and installed files.
// usually src files are located at 'vendor/src/PackageName/', installed files are located at 'vendor/pkg/PackageName/'.
// pkgHome: project root direction.
func (f *fetch) dlPackagesDepSrc(pkgLock *map[string]string, packages map[string]pkg.YamlGitPackage) ([]*pkg.DependencyTree, error) {
	var deps []*pkg.DependencyTree
	// todo packages have dependencies.
	// todo check install.
	if packages == nil {
		return deps, nil
	}
	// download git src, and add it to build tree.
	for key, gitPkg := range packages {
		context := pkg.PackageMeta{
			Version:    gitPkg.Version,
			TargetName: gitPkg.Target,
			CMakeLib:   gitPkg.CMakeLib,
			Builder:    gitPkg.Build[:],
		}
		// parse package path(name), target and version from key and gitPkg
		if err := context.SetPackageName(key); err != nil {
			return nil, err
		}

		// set save directory path
		status := pkg.DlStatusEmpty
		srcDes, err := pkg.GetCachedPackageSrcPath(context.PackageName, context.Version)
		if err != nil {
			return nil, err
		}
		// version deciding
		if ver, ok := (*pkgLock)[context.PackageName]; ok {
			newVerDes, err := pkg.GetCachedPackageSrcPath(context.PackageName, ver)
			if err != nil {
				return nil, err
			}
			srcDes = newVerDes // use the matched version package
			context.Version = ver
			log.WithFields(log.Fields{
				"pkg":     key,
				"version": ver,
			}).Trace("package matches another version.")
		} else {
			// log version
			(*pkgLock)[context.PackageName] = context.Version
		}
		// check directory, if not exists, then create it.
		if _, err := os.Stat(srcDes); os.IsNotExist(err) {
			if gitPkg.Path == "" {
				gitPkg.Path = fmt.Sprintf("https://%s.git", context.PackageName)
			}
			if err := gitSrc(f.Auth, srcDes, srcDes, gitPkg.Path, context.Version); err != nil {
				_ = os.RemoveAll(srcDes)
				return nil, err
			}
			status = pkg.DlStatusOk
		} else if err != nil {
			return nil, err
		} else {
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
