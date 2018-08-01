package install

import (
	"io"
	"github.com/genshen/pkg/utils"
	"log"
	"os"
	"strings"
	"path/filepath"
	"text/template"
)

type cmakeDepData struct {
	LibName           string
	SrcPath           string
	InnerBuildCommand []string
	OuterBuildCommand []string
	InnerCMake        string
	OuterCMake        string
}

const CmakeToFile = `
# lib {.LibName}
# build command:
#     inner build command: {{.InnerBuildCommand}}
#     outer build command: {{.OuterBuildCommand}}
# src: {{.SrcPath}}
{{.InnerCMake}} # inner cmake
{{.OuterCMake}} # outer cmake
`

// todo combine this function anf function buildPkg.
func cmakeLib(dep *DependencyTree, pkgHome string, cmakeLibSet *map[string]bool, writer io.Writer) error {
	// if this package has been built, skip it and its dependency.
	if _, ok := (*cmakeLibSet)[dep.Context.PackageName]; ok {
		return nil
	}

	for _, v := range dep.Dependency {
		if err := cmakeLib(v, pkgHome, cmakeLibSet, writer); err != nil {
			return err // break loop.
		}
	}

	// do not generate cmake script for empty lib (e.g. root lib).
	if dep.DlStatus == DlStatusEmpty {
		return nil
	}

	// generating cmake script.
	toFile := cmakeDepData{
		LibName:           dep.Context.PackageName,
		InnerCMake:        dep.SelfCMakeLib,
		OuterCMake:        dep.CMakeLib,
		OuterBuildCommand: dep.Builder,
		InnerBuildCommand: dep.SelfBuild,
		SrcPath:           utils.GetPkgPath(pkgHome, dep.Context.PackageName),
	}
	if dep.Context.CMakeLibOverride { // self cmake
		toFile.InnerCMake = ""
	}
	if err := genCMake(toFile, writer); err != nil {
		return err
	}
	log.Println("generated cmake for package", dep.Context.PackageName)
	return nil
}

// change path to relative path, replace PKG_DIR with relative path.
func preRender(cmake, pkgPath string) (error, string) {
	// replace {PKG_DIR} variable with relative path.
	if pwd, err := os.Getwd(); err != nil {
		return err, ""
	} else {
		relPkg := strings.TrimPrefix(pkgPath, pwd) // relative pkg path
		cmake = strings.Replace(cmake, "{PKG_DIR}", relPkg, -1)
		cmake = strings.TrimPrefix(cmake, string(filepath.Separator))
		return nil, cmake
	}
}

func genCMake(cmake cmakeDepData, writer io.Writer) error {
	// convert relative path
	var err error
	if err, cmake.InnerCMake = preRender(cmake.InnerCMake, cmake.SrcPath); err != nil {
		return err
	}
	if err, cmake.OuterCMake = preRender(cmake.OuterCMake, cmake.SrcPath); err != nil {
		return err
	}
	// render template.
	if t, err := template.New("cmake").Parse(CmakeToFile); err != nil {
		return err
	} else {
		if err := t.Execute(writer, cmake); err != nil {
			return err
		}
	}
	return nil
}
