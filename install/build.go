package install

import (
	"log"
	"github.com/genshen/pkg/utils"
)

// build pkg from dependency tree.
// pkgHome: the location of file pkg.json
func buildPkg(dep *DependencyTree, pkgHome string, builtSet *map[string]bool) error {
	// if this package has been built, skip it and its dependency.
	if _, ok := (*builtSet)[dep.Context.PackageName]; ok {
		return nil
	}

	log.Println("installing package", dep.Context.PackageName)

	// build children
	for _, v := range dep.Dependency {
		if err := buildPkg(v, pkgHome, builtSet); err != nil {
			return err // break loop.
		}
	}

	log.Println(dep.DlStatus)
	if dep.DlStatus == DlStatusEmpty { // ignore empty package, if it is skipped package, just reinstall it.
		return nil
	}
	// run self build.
	if !dep.Context.Override {
		for _, ins := range dep.SelfBuild {
			if err := utils.RunIns(pkgHome, dep.Context.PackageName, dep.Context.SrcPath, ins); err != nil {
				return err
			}
		}
	}
	// run outer build.
	for _, ins := range dep.Builder {
		if err := utils.RunIns(pkgHome, dep.Context.PackageName, dep.Context.SrcPath, ins); err != nil {
			return err
		}
	}
	(*builtSet)[dep.Context.PackageName] = true
	log.Println("package", dep.Context.PackageName, "installed")
	return nil
}
