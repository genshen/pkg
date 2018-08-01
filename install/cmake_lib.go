package install

import (
	"io"
	"github.com/genshen/pkg/utils"
	"log"
	"text/template"
	"strings"
	"os"
	"path/filepath"
)

type cmakeDepData struct {
	LibName           string
	PkgHome           string
	SrcDir            string
	PkgDir            string
	InnerBuildCommand []string
	OuterBuildCommand []string
	InnerCMake        string
	OuterCMake        string
}

const CmakeToFile = `
# lib {.LibName}
# src: {{.SrcDir}}
# pkg: {{.PkgDir}}
# build command:
#     inner build command: {{.InnerBuildCommand}}
#     outer build command: {{.OuterBuildCommand}}
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
		PkgHome:           pkgHome,
		SrcDir:            relativePath(utils.GetPackageSrcPath(pkgHome, dep.Context.PackageName)),
		PkgDir:            relativePath(utils.GetPkgPath(pkgHome, dep.Context.PackageName)),
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

// replace {CACHE} {PKG_DIR} {SRC_DIR} to template style
func preRender(target, pkgHome, packageName string) string {
	target = strings.Replace(target, "{CACHE}", relativePath(utils.GetCachePath(pkgHome, packageName)), -1)
	target = strings.Replace(target, "{PKG_DIR}", relativePath(utils.GetPkgPath(pkgHome, packageName)), -1)
	target = strings.Replace(target, "{SRC_DIR}", relativePath(utils.GetPackageSrcPath(pkgHome, packageName)), -1)
	return target
}

//// change path to relative path, replace PKG_DIR with relative path.
func relativePath(target string) string {
	//	// replace absolute patg with relative path.
	if pwd, err := os.Getwd(); err != nil {
		return ""
	} else {
		relPath := strings.TrimPrefix(target, pwd) // relative pkg path
		relPath = strings.TrimPrefix(relPath, string(filepath.Separator))
		return relPath
	}
}

func genCMake(cmake cmakeDepData, writer io.Writer) error {
	if cmake.InnerCMake == "" && cmake.OuterCMake == "" {
		return nil
	}
	cmake.InnerCMake = preRender(cmake.InnerCMake, cmake.PkgHome, cmake.LibName)
	cmake.OuterCMake = preRender(cmake.OuterCMake, cmake.PkgHome, cmake.LibName)
	for i, v := range cmake.InnerBuildCommand {
		cmake.InnerBuildCommand[i] = preRender(v, cmake.PkgHome, cmake.LibName)
	}
	for i, v := range cmake.OuterBuildCommand {
		cmake.OuterBuildCommand[i] = preRender(v, cmake.PkgHome, cmake.LibName)
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
