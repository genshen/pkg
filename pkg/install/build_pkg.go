package install

import (
	"bufio"
	"fmt"
	"github.com/genshen/pkg"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
)

// build pkg from dependency tree.
// pkgHome: the location of file pkg.yaml
// skipDep: skip its dependency packages.
func buildPkg(lists []string, metas map[string]pkg.PackageMeta, pkgHome string, verbose bool) error {
	var insExe = NewInsExecutor(pkgHome, verbose)

	for _, item := range lists {
		log.WithFields(log.Fields{
			"pkg": item,
		}).Info("installing package.")

		meta, ok := metas[item]
		if !ok {
			return fmt.Errorf("package %s not found", item)
		}

		pkg.AddVendorPathEnv(pkgHome)                     // use absolute path.
		pkg.AddPathEnv(item, meta.VendorSrcPath(pkgHome)) // add vars for this package, using relative path.
		// if outer build is specified, then inner build will be ignored.
		if len(meta.Builder) == 0 {
			// run inner build,(self build).
			for _, ins := range meta.SelfBuild {
				// if it is auto pkg and outer build mode
				if pkgEnvInc := os.Getenv("PKG_INNER_BUILD"); pkgEnvInc == "" && ins == pkg.InsAutoPkg {
					// use cmake instruction with features (features as cmake options)
					cmakeOpts := featuresToOptions(meta.Features)
					ins = fmt.Sprintf(`%s "%s" "%s"`, pkg.InsCmake, cmakeOpts, "")
				}
				// replace vars in instruction with real value and run the instruction.
				if err := RunIns(&insExe, &meta, pkg.ProcessEnv(ins)); err != nil {
					return err
				}
			}
		} else {
			// run outer build.
			for _, ins := range meta.Builder {
				if err := RunIns(&insExe, &meta, pkg.ProcessEnv(ins)); err != nil {
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
	shWriter, err := NewInsShellWriter(pkgHome, w)
	if err != nil {
		return err
	}
	if err := shWriter.Setup(); err != nil {
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
				// if it is auto pkg and outer build mode
				if pkgEnvInc := os.Getenv("PKG_INNER_BUILD"); pkgEnvInc == "" && ins == pkg.InsAutoPkg {
					// use cmake instruction with features (features as cmake options)
					cmakeOpts := featuresToOptions(meta.Features)
					ins = fmt.Sprintf(`%s "%s" "%s"`, pkg.InsCmake, cmakeOpts, "")
				}
				// replace vars in instruction with real value and run the instruction.
				if err := RunIns(shWriter, &meta, pkg.ProcessEnv(ins)); err != nil {
					return err
				}
			}
		} else {
			// run outer build.
			for _, ins := range meta.Builder {
				if err := RunIns(shWriter, &meta, pkg.ProcessEnv(ins)); err != nil {
					return nil
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
