package install

import (
	"fmt"
	"github.com/genshen/pkg"
	"strings"
)

// build pkg from dependency tree.
// pkgHome: the location of file pkg.yaml
// skipDep: skip its dependency packages.
func buildPkg(inst InsInterface, lists []string, metas map[string]pkg.PackageMeta, pkgHome string) error {
	for _, item := range lists {
		meta, ok := metas[item]
		if !ok {
			return fmt.Errorf("package %s not found", item)
		}

		pkg.AddVendorPathEnv(pkgHome) // use absolute path.
		// add vars for this package, using relative path.
		if err := pkg.AddPathEnv(item, meta.VendorSrcPath(pkgHome)); err != nil {
			return err
		}
		// if outer build is specified, then inner build will be ignored.
		if len(meta.Builder) == 0 {
			// run inner build,(self build).
			for _, ins := range meta.SelfBuild {
				// replace vars in instruction with real value and run the instruction.
				if err := RunIns(inst, &meta, pkg.ProcessEnv(ins)); err != nil {
					return err
				}
			}
		} else {
			// run outer build.
			for _, ins := range meta.Builder {
				if err := RunIns(inst, &meta, pkg.ProcessEnv(ins)); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func generateShell(inst InsInterface, lists []string, metas map[string]pkg.PackageMeta, pkgHome string) error {
	if err := inst.Setup(); err != nil {
		return err
	}

	for _, item := range lists {
		meta, ok := metas[item]
		if !ok {
			return fmt.Errorf("package `%s` not found", item)
		}

		if err := inst.PkgPreInstall(&meta); err != nil {
			return err
		}

		// using short path with env '$PKG_SRC_PATH'.
		packageSrc := strings.Replace(meta.VendorSrcPath(pkgHome), pkg.GetPkgSrcPath(pkgHome), "$PKG_SRC_PATH", 1)
		pkg.AddVendorPathEnv("$PROJECT_HOME") // use absolute path.
		// add vars for this package
		if err := pkg.AddPathEnv(item, packageSrc); err != nil {
			return err
		}
		// if outer build is specified, then inner build will be ignored.
		if len(meta.Builder) == 0 {
			// run inner build,(self build).
			for _, ins := range meta.SelfBuild {
				// replace vars in instruction with real value and run the instruction.
				if err := RunIns(inst, &meta, pkg.ProcessEnv(ins)); err != nil {
					return err
				}
			}
		} else {
			// run outer build.
			for _, ins := range meta.Builder {
				if err := RunIns(inst, &meta, pkg.ProcessEnv(ins)); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func featuresToOptions(features []string) string {
	var strBuilder strings.Builder
	for _, feature := range features {
		strBuilder.WriteString(fmt.Sprintf("-D%s ", feature))
	}
	return strBuilder.String()
}
