package install

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/genshen/pkg"
	"os"
	"path/filepath"
	"strings"
)

// writer instructions as shell format to file
type InsShellWriter struct {
	BaseInsExecutor
	pkgHome string        // home directory of running pkg command
	writer  *bufio.Writer // shell script file writer
}

func NewInsShellWriter(pkgHome string, w *bufio.Writer, cmakeConfigArg, cmakeBuildArg string) (*InsShellWriter, error) {
	return &InsShellWriter{
		BaseInsExecutor: BaseInsExecutor{
			cmakeConfigArg: cmakeConfigArg,
			cmakeBuildArg:  cmakeBuildArg,
		},
		pkgHome: pkgHome,
		writer:  w,
	}, nil
}

func (sh *InsShellWriter) Setup() error {
	const shellHead = `#!/bin/sh
set -e

export PKG_VENDOR_PATH=%s
PROJECT_HOME=%s
PKG_SRC_PATH=%s
`

	pkgSrcPath := pkg.GetPkgSrcPath(sh.pkgHome)
	if _, err := sh.writer.WriteString(fmt.Sprintf(shellHead, pkg.GetVendorPath(sh.pkgHome), sh.pkgHome, pkgSrcPath)); err != nil {
		return err
	}

	return nil
}

func (sh *InsShellWriter) PkgPreInstall(meta *pkg.PackageMeta) (*pkg.PackageEnvs, error) {
	// using short path with env '$PKG_SRC_PATH'.
	packageSrcPath := strings.Replace(meta.VendorSrcPath(sh.pkgHome), pkg.GetPkgSrcPath(sh.pkgHome), "$PKG_SRC_PATH", 1)
	// package env
	packageEnv := pkg.NewPackageEnvs("$PROJECT_HOME", meta.PackageName, packageSrcPath)
	if _, err := sh.writer.WriteString(fmt.Sprintf("\n## pacakge %s\n", meta.PackageName)); err != nil {
		return nil, err
	}
	return packageEnv, nil
}

func (sh *InsShellWriter) PkgPostInstall(meta *pkg.PackageMeta) error {
	return nil
}

func (sh *InsShellWriter) InsCp(triple pkg.InsTriple, meta *pkg.PackageMeta) error {
	if triple.Second == "" || triple.Third == "" {
		return errors.New("CP instruction must have src and des")
	}

	pathBase := "${PROJECT_HOME}"
	srcPath := meta.VendorSrcPath(pathBase)
	if _, err := sh.writer.WriteString(fmt.Sprintf("mkdir -p \"%s\"\n", pkg.GetIncludePath(pathBase))); err != nil {
		return err
	}
	if _, err := sh.writer.WriteString(fmt.Sprintf("cp -r \"%s\" \"%s\"\n",
		filepath.Join(srcPath, triple.Second), triple.Third)); err != nil {
		return err
	}
	return nil
}

func (sh *InsShellWriter) InsRun(triple pkg.InsTriple, meta *pkg.PackageMeta) error {
	if triple.Second == "" || triple.Third == "" {
		return errors.New("RUN instruction must be a triple")
	}

	if _, err := sh.writer.WriteString(fmt.Sprintf("mkdir -p \"%s\"\ncd \"%s\"\n%s\n",
		triple.Second, triple.Second, triple.Third)); err != nil {
		return err
	}
	return nil
}

func (sh *InsShellWriter) InsCMake(triple pkg.InsTriple, meta *pkg.PackageMeta) error {
	pathBase := "${PROJECT_HOME}"
	srcPath := meta.VendorSrcPath(pathBase)

	if sh.cmakeConfigArg != "" {
		triple.Second = triple.Second + " " + sh.cmakeConfigArg
	}
	if sh.cmakeBuildArg != "" {
		triple.Third = triple.Third + " " + sh.cmakeBuildArg
	}
	var configCmd = fmt.Sprintf("cmake -S \"%s\" -B \"%s\" -DCMAKE_BUILD_TYPE=Release -DCMAKE_INSTALL_PREFIX=\"%s\" %s",
		srcPath, pkg.GetCachePath(pathBase, meta.PackageName),
		pkg.GetPackagePkgPath(pathBase, meta.PackageName), triple.Second)
	var buildCmd = fmt.Sprintf("cmake --build \"%s\" --target install %s",
		pkg.GetCachePath(pathBase, meta.PackageName), triple.Third)
	if _, err := sh.writer.WriteString(fmt.Sprintf("cd \"%s\"\n%s\n%s\n", pathBase, configCmd, buildCmd)); err != nil {
		return err
	}
	return nil
}

func (sh *InsShellWriter) InsAutoPkg(triple pkg.InsTriple, meta *pkg.PackageMeta) error {
	// if it is auto pkg and outer build mode
	// todo only for inner builders
	if pkgEnvInc := os.Getenv("PKG_INNER_BUILD"); pkgEnvInc == "" {
		// use cmake instruction with features (features as cmake options)
		triple.First = pkg.InsCmake
		triple.Second = featuresToOptions(meta.Features)
		triple.Third = ""
		return sh.InsCMake(triple, meta)
	}
	return nil
}
