package install

import (
	"bufio"
	"fmt"
	"github.com/genshen/pkg"
	log "github.com/sirupsen/logrus"
)

// build pkg from dependency tree.
// pkgHome: the location of file pkg.json
// skipDep: skip its dependency packages.
func buildPkg(lists []string, metas map[string]pkg.PackageMeta, pkgHome string, verbose bool) error {
	for _, item := range lists {
		log.WithFields(log.Fields{
			"pkg": item,
		}).Info("installing package.")

		meta, ok := metas[item]
		if !ok {
			return fmt.Errorf("package %s not found", item)
		}

		pkg.AddVendorPathEnv(pkgHome)      // use absolute path.
		pkg.AddPathEnv(item, meta.SrcPath) // add vars for this package, using relative path.
		// if outer build is specified, then inner build will be ignored.
		if len(meta.Builder) == 0 {
			// run inner build,(self build).
			for _, ins := range meta.SelfBuild {
				// replace vars in instruction with real value and run the instruction.
				if err := RunIns(pkgHome, item, meta.SrcPath, pkg.ProcessEnv(ins), verbose); err != nil {
					return err
				}
			}
		} else {
			// run outer build.
			for _, ins := range meta.Builder {
				if err := RunIns(pkgHome, item, meta.SrcPath, pkg.ProcessEnv(ins), verbose); err != nil {
					return err
				}
			}
		}

		log.WithFields(log.Fields{
			"pkg": item,
		}).Info("package built and installed.")
	}
	return nil
}

func generateShell(w *bufio.Writer, lists []string, metas map[string]pkg.PackageMeta, pkgHome string) error {
	if _, err := w.WriteString("#!/bin/sh\n"); err != nil {
		return err
	}

	for _, item := range lists {
		meta, ok := metas[item]
		if !ok {
			return fmt.Errorf("package `%s` not found", item)
		}

		if _, err := w.WriteString(fmt.Sprintf("\n## pacakge %s\n", item)); err != nil {
			return err
		}

		pkg.AddVendorPathEnv(pkgHome)                              // use absolute path.
		if err := pkg.AddPathEnv(item, meta.SrcPath); err != nil { // add vars for this package, using relative path.
			return err
		}
		// if outer build is specified, then inner build will be ignored.
		if len(meta.Builder) == 0 {
			// run inner build,(self build).
			for _, ins := range meta.SelfBuild {
				// replace vars in instruction with real value and run the instruction.
				if err := WriteIns(w, pkgHome, item, pkg.ProcessEnv(ins)); err != nil {
					return err
				}
			}
		} else {
			// run outer build.
			for _, ins := range meta.Builder {
				if err := WriteIns(w, pkgHome, item, pkg.ProcessEnv(ins)); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
