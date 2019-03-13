package install

import (
	"github.com/genshen/pkg"
	log "github.com/sirupsen/logrus"
)

// build pkg from dependency tree.
// pkgHome: the location of file pkg.json
// skipDep: skip its dependency packages.
func buildPkg(dep *pkg.DependencyTree, pkgHome string, root bool, skipDep bool, builtSet *map[string]bool) error {
	// if this package has been built, skip it and its dependency.
	if _, ok := (*builtSet)[dep.Context.PackageName]; ok {
		return nil
	}

	log.WithFields(log.Fields{
		"pkg": dep.Context.PackageName,
	}).Info("installing package.")

	// load children
	if !skipDep {
		// keep to build its dependency packages.
		for _, v := range dep.Dependencies {
			if err := buildPkg(v, pkgHome, false, skipDep, builtSet); err != nil {
				return err // break loop.
			}
		}
	}

	if dep.DlStatus == pkg.DlStatusEmpty || root { // ignore empty and root package.
		return nil
	}

	addVendorPathEnv(pkgHome)           // use absolute path.
	addPathEnv(dep.Context.PackageName) // add vars for this package, using relative path.
	// if outer build is specified, then inner build will be ignored.
	if len(dep.Builder) == 0 {
		// run inner build,(self build).
		for _, ins := range dep.SelfBuild {
			// replace vars in instruction with real value and run the instruction.
			if err := RunIns(pkgHome, dep.Context.PackageName, dep.Context.SrcPath, processEnv(ins)); err != nil {
				return err
			}
		}
	} else {
		// run outer build.
		for _, ins := range dep.Builder {
			if err := RunIns(pkgHome, dep.Context.PackageName, dep.Context.SrcPath, processEnv(ins)); err != nil {
				return err
			}
		}
	}

	(*builtSet)[dep.Context.PackageName] = true
	log.WithFields(log.Fields{
		"pkg": dep.Context.PackageName,
	}).Info("package installed.")
	return nil
}
