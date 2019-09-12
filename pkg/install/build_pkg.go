package install

import (
	"github.com/genshen/pkg"
	log "github.com/sirupsen/logrus"
)

// build pkg from dependency tree.
// pkgHome: the location of file pkg.json
// skipDep: skip its dependency packages.
func buildPkg(metas []pkg.PackageMeta, pkgHome string,  verbose bool) error {
	for _, meta := range metas {
		log.WithFields(log.Fields{
			"pkg": meta.PackageName,
		}).Info("installing package.")

		pkg.AddVendorPathEnv(pkgHome)    // use absolute path.
		pkg.AddPathEnv(meta.PackageName) // add vars for this package, using relative path.
		// if outer build is specified, then inner build will be ignored.
		if len(meta.Builder) == 0 {
			// run inner build,(self build).
			for _, ins := range meta.SelfBuild {
				// replace vars in instruction with real value and run the instruction.
				if err := RunIns(pkgHome, meta.PackageName, meta.SrcPath, pkg.ProcessEnv(ins), verbose); err != nil {
					return err
				}
			}
		} else {
			// run outer build.
			for _, ins := range meta.Builder {
				if err := RunIns(pkgHome, meta.PackageName, meta.SrcPath, pkg.ProcessEnv(ins), verbose); err != nil {
					return err
				}
			}
		}

		log.WithFields(log.Fields{
			"pkg": meta.PackageName,
		}).Info("package installed.")
	}
	return nil
}
