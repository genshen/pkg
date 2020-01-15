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
	if err := inst.Setup(); err != nil {
		return nil
	}
	for _, item := range lists {
		meta, ok := metas[item]
		if !ok {
			return fmt.Errorf("package `%s` not found", item)
		}

		packageEnv, err := inst.PkgPreInstall(&meta)
		if err != nil {
			return err
		}

		// if outer build is specified, then inner build will be ignored.
		if len(meta.Builder) == 0 {
			// run inner build,(self build).
			for _, ins := range meta.SelfBuild {
				if err := RunIns(inst, &meta, packageEnv, ins); err != nil {
					return err
				}
			}
		} else {
			// run outer build.
			for _, ins := range meta.Builder {
				if err := RunIns(inst, &meta, packageEnv, ins); err != nil {
					return err
				}
			}
		}
		if err := inst.PkgPostInstall(&meta); err != nil {
			return err
		}
	}
	return nil
}

// run instruction.
func RunIns(inst InsInterface, meta *pkg.PackageMeta, envs *pkg.PackageEnvs, ins string) error {
	if expandedIns, err := pkg.ExpandEnv(ins, envs); err != nil {
		return err
	} else {
		// parse instruction
		triple, err := pkg.ParseIns(strings.Trim(expandedIns, " "))
		if err != nil {
			return err
		}

		switch triple.First {
		case "CP":
			if err := inst.InsCp(triple, meta); err != nil {
				return err
			}
		case "RUN":
			if err := inst.InsRun(triple, meta); err != nil {
				return err
			}
		case pkg.InsCmake: // run cmake commands, format: CMAKE {config args} {build args}
			if err := inst.InsCMake(triple, meta); err != nil {
				return err
			}
		case pkg.InsAutoPkg:
			if err := inst.InsAutoPkg(triple, meta); err != nil {
				return err
			}
		}
		return nil
	}
}

func featuresToOptions(features []string) string {
	var strBuilder strings.Builder
	for _, feature := range features {
		strBuilder.WriteString(fmt.Sprintf("-D%s ", feature))
	}
	return strBuilder.String()
}
