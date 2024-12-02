package fetch

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/genshen/cmds"
	"github.com/genshen/pkg"
	"github.com/genshen/pkg/conf"
	"github.com/otiai10/copy"
	"github.com/rogpeppe/go-internal/semver"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
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
	fetchCommand.FlagSet.StringVar(&f.CMakeFindPackageOption, "cmake-find-package-arg", "NO_DEFAULT_PATH", "global options for find_package when generating file pkg.dep.cmake")
	fetchCommand.FlagSet.StringVar(&f.FeaturesOption, "features", DefaultFeatureName, "Comma separated list of features to activate. e.g. --features=foo,bar")
	// todo make pkgHome abs path anyway.
	fetchCommand.FlagSet.Usage = fetchCommand.Usage // use default usage provided by cmds.Command.
	fetchCommand.Runner = &f
	cmds.AllCommands = append(cmds.AllCommands, fetchCommand)
}

type fetch struct {
	PkgHome                string   // the absolute path of root 'pkg.yaml' form command path.
	CMakeFindPackageOption string   // global find_package option, default is "NO_DEFAULT_PATH".
	FeaturesOption         string   // cli `feature` string
	FeatureList            []string // feature list parsed from cli option.
	MirrorConfPath         string   // the file path of repo mirror file.
	DepTree                pkg.DependencyTree
	Auth                   map[string]conf.Auth
	GlobalReplace          map[string]string
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
		f.GlobalReplace = config.GitReplace
	}

	// parse feature list
	if f.FeaturesOption != "" {
		log.Info("Following features are enabled: ", f.FeaturesOption)
		f.FeatureList = strings.Split(f.FeaturesOption, ",")
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
	if err := f.fetchSubDependency(pkg.RootPKG, f.PkgHome, f.FeatureList, &pkgLock, &f.DepTree); err != nil {
		return err
	}

	// process package conflict
	packageConflict := func(packageName string, packs pkg.PackageMetas) (pkg.PackageMeta, error) {
		if len(packs) == 1 {
			return packs[0], nil // quick return
		}

		helpBuff := bytes.Buffer{}

		if tmpl, err := template.New("help").Parse(`Packages: NO. PackageName@Version#Targte
{{range $i, $p := . }} {{$i}}: {{$p.PackageName}}@{{$p.Version}}#{{$p.TargetName}}
    Features: [{{range $f := $p.Features }} {{$f}} {{end}}]
    Builder: [{{range $b := $p.Builder }} {{$b}}; {{end}}]
    SelfBuild: [{{range $s := $p.SelfBuild }} {{$s}}; {{end}}]
    CMakeLib: {{$p.CMakeLib}}
    SelfCMakeLib: {{$p.SelfCMakeLib}}
{{end}}`); err != nil {
			return pkg.PackageMeta{}, err
		} else {
			if err := tmpl.Execute(&helpBuff, packs); err != nil {
				return pkg.PackageMeta{}, err
			}
		}

		var qs = []*survey.Question{
			{
				Name: "packages",
				Prompt: &survey.Input{
					Message: fmt.Sprintf("Package `%s` conflict, select one:", packageName),
					Help:    helpBuff.String(),
				},
			},
		}
		answers := struct {
			Selection int `survey:"packages"`
		}{}

		// perform the questions
		err := survey.Ask(qs, &answers)
		if err != nil {
			return pkg.PackageMeta{}, err
		}
		if answers.Selection >= 0 && answers.Selection < len(packs) {
			return packs[answers.Selection], nil
		} else {
			return pkg.PackageMeta{}, errors.New("conflict selection out of range")
		}
	}

	// dump dependency tree to file system
	if err := f.DepTree.Dump(pkg.GetPkgSumPath(f.PkgHome), packageConflict); err != nil { //fixme
		return err
	} else {
		log.WithFields(log.Fields{
			"file": pkg.PkgSumFileName,
		}).Info("saved dependencies tree to file.")
	}

	// dump all packages' dependencies.
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
	if err := createPkgDepCmake(f.PkgHome, &f.DepTree, f.CMakeFindPackageOption); err != nil {
		return err
	}

	log.Info("fetch succeeded.")
	return nil
}

// fetchSubDependency installs dependencies to a directory.
// installPath is the root path of sub-dependency(always be the project root).
// pkgPath: the given package name/path (e.g github.com/google/googletest) from top level package.
// activeFeatList: a list of features to be active.
// pkgVendorSrcPath: path of source file directory in vendor.
// todo circle detect
func (f *fetch) fetchSubDependency(pkgPath string, pkgVendorSrcPath string, activeFeatList []string, pkgLock *map[string]string, depTree *pkg.DependencyTree) error {
	// check pkg.yaml file in vendor directory
	if pkgYamlFile, err := os.Open(filepath.Join(pkgVendorSrcPath, pkg.PkgFileName)); err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	} else { // pkg.yaml exists.
		defer pkgYamlFile.Close()
		if bytes, err := io.ReadAll(pkgYamlFile); err != nil { // read file contents
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

			// initial empty local replace list
			if pkgYaml.GitReplace == nil {
				pkgYaml.GitReplace = make(map[string]string)
			}

			// compare min pkg version in yaml file
			if pkgYaml.MinPkgVersion != "" {
				if !semver.IsValid(pkgYaml.MinPkgVersion) {
					return fmt.Errorf("min package version %s is not valid", pkgYaml.MinPkgVersion)
				} else if semver.Compare(pkgYaml.MinPkgVersion, pkg.VERSION) > 0 {
					return fmt.Errorf("pkg version %s is too old, require min pkg version is %s", pkg.VERSION, pkgYaml.MinPkgVersion)
				}
			}

			if pkgYaml.FormatVersion != pkg.FORMAT_VERSION {
				return fmt.Errorf("package format version does not match, require format version is %d", pkg.FORMAT_VERSION)
			}

			// process features: filter active features and get the optional packages for the features.
			err, activateFeatPkgs := activeFeatureOptionalPackages(pkgYaml.Features, activeFeatList)
			if err != nil {
				return err
			}

			// download git based packages source of direct dependencies.
			if deps, err := f.dlPackagesDepSrc(pkgLock, activateFeatPkgs, pkgYaml.GitReplace, f.GlobalReplace, gitPkgsToInterface(pkgYaml.Deps.GitPackages)); err != nil {
				return err
			} else {
				// add and install sub dependencies for this package.
				depTree.Dependencies = append(depTree.Dependencies, deps...)
				for _, dep := range deps {
					// todo: currently, we disable features for sub dependencies.
					if err := f.fetchSubDependency(dep.Context.PackageName, dep.Context.VendorSrcPath(f.PkgHome), nil, pkgLock, dep); err != nil {
						return err
					}
				}
			}
			// download file based packages without recursion.
			if deps, err := f.dlPackagesDepSrc(pkgLock, activateFeatPkgs, pkgYaml.GitReplace, f.GlobalReplace, filesPkgsToInterface(pkgYaml.Deps.FilesPackages)); err != nil {
				return err
			} else {
				depTree.Dependencies = append(depTree.Dependencies, deps...)
			}
			// download archive based packages without recursion.
			if deps, err := f.dlPackagesDepSrc(pkgLock, activateFeatPkgs, pkgYaml.GitReplace, f.GlobalReplace, archivePkgsToInterface(pkgYaml.Deps.ArchivePackages)); err != nil {
				return err
			} else {
				depTree.Dependencies = append(depTree.Dependencies, deps...)
			}
		}
	}
	return nil
}

// download a package source to destination refer to installPath, including source code and installed files.
// usually src files are located at 'vendor/src/PackageName/', installed files are located at 'vendor/pkg/PackageName/'.
// pkgHome: project root direction.
func (f *fetch) dlPackagesDepSrc(pkgLock *map[string]string, featPkgList []string, localReplace, globalReplace map[string]string,
	packages map[string]PackageFetcher) ([]*pkg.DependencyTree, error) {
	var deps []*pkg.DependencyTree
	// todo check install.

	if packages == nil {
		return deps, nil
	}
	// download git, files and archive src, and add it to build tree.
	for key, p := range packages {
		var context pkg.PackageMeta
		// before fetching package, set version and package name/path
		if err := p.setPackageMeta(key, &context); err != nil {
			return nil, err
		}

		if context.Optional && !checkOptionalPackageFeatureMatches(context, featPkgList) { // skip optional packages. We do not add to dependency records.
			log.WithFields(log.Fields{"pkg": context.PackageName}).Info("optional package.")
			continue
		}

		// set save directory path
		status := pkg.DlStatusEmpty

		// src path in (global) user home
		srcDes := context.HomeCacheSrcPath()
		vendorSrcDes := context.VendorSrcPath(f.PkgHome)

		err, strategy := determinePackageCacheStrategy(context, f.PkgHome)
		if err != nil {
			return nil, err
		}

		switch strategy {
		case CacheStrategyDownloadFromRemote:
			log.WithFields(log.Fields{"pkg": context.PackageName, "storage": srcDes}).Info("downloading dependencies.")
			if err := p.fetch(f.Auth, localReplace, globalReplace, srcDes, context); err != nil {
				return nil, err
			}
			status = pkg.DlStatusOk
		case CacheStrategyCopyFromGlobalCache:
			log.WithFields(log.Fields{"pkg": key, "src_path": srcDes}).Info("skipped fetching package, because it already exists.")
			// copy downloaded packages from global cache to vendor directory
			if err := copy.Copy(srcDes, vendorSrcDes); err != nil {
				return nil, err
			}
			status = pkg.DlStatusSkip
		case CacheStrategyUserLocalVendor:
			log.WithFields(log.Fields{"pkg": key, "src_path": vendorSrcDes}).Info("skipped fetching package, because it already exists.")
			status = pkg.DlStatusSkip
		case CacheStrategySkip:
			// not handled: skip with error.
		default:
			return nil, fmt.Errorf("unknown package cache strategy: %d", strategy)
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
