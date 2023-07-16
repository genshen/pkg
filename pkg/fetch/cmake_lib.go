package fetch

import (
	"bufio"
	"fmt"
	"github.com/genshen/pkg"
	"github.com/genshen/pkg/pkg/version"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type cmakeDepData struct {
	pkg.PackageMeta
	SrcDir            string
	PkgDir            string
	DepsDir           string // cmake binary dir if it is added by add_subdirectory
	InnerBuildCommand []string
	OuterBuildCommand []string
}

type cmakeDepHeaderData struct {
	IsProjectPkg      bool
	ProjectVendorPath string // vendor directory path of project
	ProjectHomePath   string // directory of project
}

const PkgCMakeHeader = `##### this file is generated by pkg tool, version ` + version.VERSION +
	`
##### For more details, please address https://github.com/genshen/pkg.

# vendor path
{{if .IsProjectPkg}}
# you should change VENDOR_PATH if you moved this directory to other place.
set(VENDOR_PATH {{.ProjectVendorPath}})
{{else}}
# VENDOR_PATH is set by environment variable 'PKG_VENDOR_PATH'.
set(VENDOR_PATH $ENV{PKG_VENDOR_PATH})
{{end}}
set(PROJECT_HOME_PATH {{.ProjectHomePath}})

include_directories(${VENDOR_PATH}/include)
`

const (
	CmakeToFileOuterBuild = `
# lib <<.PackageName>>
# src: <<.SrcDir>>
# pkg: <<.PkgDir>>
# build command:
#     inner build command: <<.InnerBuildCommand>>
#     outer build command: <<.OuterBuildCommand>>
<<if eq .SelfCMakeLib "AUTO_PKG">>
if(NOT ${<<.TargetName>>_FOUND} OR ${<<.TargetName>>_FOUND} STREQUAL "")
	find_package(<<.TargetName>> PATHS <<.PkgDir>> NO_DEFAULT_PATH)
endif()
<<else>>
	<<.SelfCMakeLib>> # inner cmake
<<end>>
<<.CMakeLib>> # outer cmake
`

	CmakeToFileInnerBuild = `
# lib <<.PackageName>>
# src: <<.SrcDir>>
# pkg: <<.PkgDir>>
# build command:
#     inner build command: <<.InnerBuildCommand>>
#     outer build command: <<.OuterBuildCommand>>
<<if eq .SelfCMakeLib "AUTO_PKG">>
if(NOT ${PKG_<<.TargetName>>_FOUND} OR ${PKG_<<.TargetName>>_FOUND} STREQUAL "")
    << range $feature := .Features >>
    	<<cmake_opt $feature>>
	<< end >>
	add_subdirectory(<<.SrcDir>> <<.DepsDir>>)
	set(PKG_<<.TargetName>>_FOUND ON CACHE BOOL "set package <<.TargetName>> found")
endif()
<<else>>
	<<.SelfCMakeLib>> # inner cmake
<<end>>
<<.CMakeLib>> # outer cmake
`
)

// pkgHome is always pkg root.
// write cmake script for all direct and indirect dependencies packages.
func createPkgDepCmake(pkgHome string, rootDep *pkg.DependencyTree) error {
	depsList, err := rootDep.ListDeps(false)
	if err != nil {
		return err
	}

	for _, depTree := range depsList {
		var packageSrcPath = depTree.Context.VendorSrcPath(pkgHome)
		if depTree.Context.PackageName == pkg.RootPKG {
			packageSrcPath = pkgHome
		}
		var createDepCmake = func(isProjectPkg bool) error {
			// create cmake dep file for this package.
			if cmakeDepWriter, err := os.Create(filepath.Join(packageSrcPath, pkg.CMakeDep)); err != nil {
				return err
			} else {
				defer cmakeDepWriter.Close()
				bufWriter := bufio.NewWriter(cmakeDepWriter)

				// render header template.
				// In header rendering, for project package, set @PkgHome/vendor as vendor path.
				// for non-project package, vendor path will not be set, which can be set in command line while building.
				if err := renderCMakeHeader(bufWriter, isProjectPkg, pkgHome, pkg.GetVendorPath(pkgHome)); err != nil {
					return err
				}

				// compute and render body template.
				// (write cmake include and find_package script of all dependency packages)
				if err := cmakeLib(depTree, pkgHome, bufWriter); err != nil {
					return err
				}
				if err := bufWriter.Flush(); err != nil {
					return err
				}
				log.Trace("generated cmake for package", depTree.Context.PackageName)
			}
			return nil
		}

		// create dep cmake file only for pkg based package.
		if depTree.IsPkgPackage {
			var isProjectPkg = false
			if depTree == rootDep {
				isProjectPkg = true
			}
			// create cmake dep file for this package
			if err := createDepCmake(isProjectPkg); err != nil {
				return err
			}
		}
	}
	return nil
}

// generate/render cmake script of a package specified by depTree.
//the result comes from its dependencies,
// pkgHome: absolute path for pkg home.
func cmakeLib(depTree *pkg.DependencyTree, pkgHome string, writer io.Writer) error {
	// skip master package by setting parameter skipRoota as true,
	// do not generate cmake include and find_package script for the master package itself lib.
	depsList, err := depTree.ListDeps(true) // list all dependencies
	if err != nil {
		return err
	}

	basePath := "${PROJECT_HOME_PATH}"
	for _, dep := range depsList {
		src := dep.Context.VendorSrcPath(basePath) // vendor/src/@pkg@version,using relative path.
		// add env variables for this package, using relative path.
		packageEnv := pkg.NewPackageEnvs(basePath, dep.Context.PackageName, src)
		// generating cmake script.
		toFile := cmakeDepData{
			PackageMeta: pkg.PackageMeta{
				PackageName:  dep.Context.PackageName,
				Version:      dep.Context.Version,
				TargetName:   dep.Context.TargetName,
				SelfCMakeLib: dep.Context.SelfCMakeLib,
				CMakeLib:     dep.Context.CMakeLib,
				Features:     dep.Context.Features,
			},
			SrcDir:  src,
			DepsDir: pkg.GetPackageDepsPath(basePath, dep.Context.PackageName),
			PkgDir:  pkg.GetPackagePkgPath(basePath, dep.Context.PackageName),
		}
		// copy slice, don't modify the original data.
		toFile.OuterBuildCommand = make([]string, len(dep.Context.Builder))
		toFile.InnerBuildCommand = make([]string, len(dep.Context.SelfBuild))
		copy(toFile.OuterBuildCommand, dep.Context.Builder)
		copy(toFile.InnerBuildCommand, dep.Context.SelfBuild)

		// ignore self cmake if the cmake in override by outer cmake lib.
		if dep.Context.CMakeLib != "" {
			toFile.SelfCMakeLib = ""
		}
		if err := renderCMakeBody(toFile, packageEnv, writer); err != nil {
			return err
		}
	}
	return nil
}

// render script for cmake header.
// isProjectPkg: the current project package.
// ProjectVendorPath: current project's vendor path
func renderCMakeHeader(writer io.Writer, isProjectPkg bool, projectRootPath, projectVendorPath string) error {
	data := cmakeDepHeaderData{IsProjectPkg: isProjectPkg, ProjectVendorPath: projectVendorPath, ProjectHomePath: projectRootPath}
	if t, err := template.New("header").Parse(PkgCMakeHeader); err != nil {
		return err
	} else {
		if err := t.Execute(writer, data); err != nil {
			return err
		}
	}
	return nil
}

func renderCMakeBody(cmake cmakeDepData, packageEnv *pkg.PackageEnvs, writer io.Writer) error {
	if cmake.SelfCMakeLib == "" && cmake.CMakeLib == "" {
		return nil
	}
	// expand self cmake lib
	if selfCmakeLib, err := pkg.ExpandEnv(cmake.SelfCMakeLib, packageEnv); err != nil {
		return err
	} else {
		cmake.SelfCMakeLib = selfCmakeLib
	}
	// expand cmake lib
	if cmakeLib, err := pkg.ExpandEnv(cmake.CMakeLib, packageEnv); err != nil {
		return err
	} else {
		cmake.CMakeLib = cmakeLib
	}
	// InnerBuildCommand and OuterBuildCommand is just used in comment.
	for i, v := range cmake.InnerBuildCommand {
		if innerBuilder, err := pkg.ExpandEnv(v, packageEnv); err != nil {
			return err
		} else {
			cmake.InnerBuildCommand[i] = innerBuilder
		}
	}
	for i, v := range cmake.OuterBuildCommand {
		if outerBuilder, err := pkg.ExpandEnv(v, packageEnv); err != nil {
			return err
		} else {
			cmake.OuterBuildCommand[i] = outerBuilder
		}
	}

	// render template.
	var cmakeRenderTpl = CmakeToFileOuterBuild
	if pkgEnvInc := os.Getenv("PKG_INNER_BUILD"); pkgEnvInc != "" {
		cmakeRenderTpl = CmakeToFileInnerBuild
	}
	if t, err := template.New("cmake").Delims("<<", ">>").Funcs(template.FuncMap{"cmake_opt": CmakeOpt}).Parse(cmakeRenderTpl)
		err != nil {
		return err
	} else {
		if err := t.Execute(writer, cmake); err != nil {
			return err
		}
	}
	return nil
}

// set cmake option in template
func CmakeOpt(feature string) (string, error) {
	pair := strings.SplitN(feature, "=", 2)
	if len(pair) != 2 {
		return "", fmt.Errorf("feature %s with incorrect format", feature)
	}
	return fmt.Sprintf("set(%s %s CACHE BOOL \"enable/disable %s\")", pair[0], pair[1], pair[0]), nil
}
