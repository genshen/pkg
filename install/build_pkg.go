package install

import (
	"github.com/genshen/pkg/utils"
	log "github.com/sirupsen/logrus"
)

// build pkg from dependency tree.
// pkgHome: the location of file pkg.json
func buildPkg(dep *DependencyTree, pkgHome string,root bool, builtSet *map[string]bool) error {
	// if this package has been built, skip it and its dependency.
	if _, ok := (*builtSet)[dep.Context.PackageName]; ok {
		return nil
	}

	log.WithFields(log.Fields{
		"pkg": dep.Context.PackageName,
	}).Info("installing package.")

	// build children
	for _, v := range dep.Dependencies {
		if err := buildPkg(v, pkgHome,false, builtSet); err != nil {
			return err // break loop.
		}
	}

	if dep.DlStatus == DlStatusEmpty || root { // ignore empty and root package.
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
	log.WithFields(log.Fields{
		"pkg": dep.Context.PackageName,
	}).Info("package installed.")
	return nil
}
