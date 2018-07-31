package install

import (
	"log"
	"github.com/genshen/pkg/utils"
)

// build pkg from dependency tree.
// pkgHome: the location of file pkg.json
func buildPkg(dep *DependencyTree, pkgHome string) error {
	// build children
	for _, v := range dep.Dependency {
		if err := buildPkg(v, pkgHome); err != nil {
			return err // break loop.
		}
	}

	// build self.
	if dep.DlStatus != DlStatusOk { // ignore skip package and empty package.
		return nil
	}

	if err := postInstall(&dep.DepPkgContext, pkgHome, dep.Builder); err != nil {
		return err
	}
	return nil
}

// the source code is in vendor/src/{packageName} directory.
// pkgHomePath is the location of 'pkg.json'.
// context: dependency context
// in this function, it start to execute 'build' command (e.g. copy header into include directory.).
func postInstall(context *DepPkgContext, pkgHomePath string, build []string) error {
	log.Println("installing package", context.PackageName)
	for _, ins := range build {
		if err := utils.RunIns(pkgHomePath, context.SrcPath, ins); err != nil {
			return err
		}
	}
	return nil
}
