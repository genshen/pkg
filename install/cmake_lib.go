package install

import (
	"io"
	"os"
	"strings"
	"github.com/genshen/pkg/utils"
	"log"
)

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
	if !dep.Context.CMakeLibOverride { // self cmake
		if err := genCMake(dep.SelfCMakeLib, dep.Context.PackageName,
			utils.GetPkgPath(pkgHome, dep.Context.PackageName), writer); err != nil {
			return err
		}
	}
	// outer cmake.
	if err := genCMake(dep.CMakeLib, dep.Context.PackageName,
		utils.GetPkgPath(pkgHome, dep.Context.PackageName), writer); err != nil {
		return err
	}
	log.Println("generated cmake for package", dep.Context.PackageName)
	return nil
}

func genCMake(cmake, packageName, pkgPath string, writer io.Writer) error {
	if cmake == "" {
		return nil
	}
	// replace {PKG_DIR} variable with relative path.
	if pwd, err := os.Getwd(); err != nil {
		return err
	} else {
		relPkg := strings.TrimPrefix(pkgPath, pwd) // relative pkg path
		cmake = strings.Replace(cmake, "{PKG_DIR}", relPkg, -1)
	}
	cmake = "#lib " + packageName + "\n" + cmake + "\n" // add a new line, add lib comment. // todo interface.
	_, err := writer.Write([]byte(cmake))
	return err
}
