package install

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/genshen/pkg"
	"path/filepath"
)

// writer instructions as shell format to file
type InsShellWriter struct {
	pkgHome string        // home directory of running pkg command
	writer  *bufio.Writer // shell script file writer
}

func NewInsShellWriter(pkgHome string, w *bufio.Writer, ) (*InsShellWriter, error) {
	return &InsShellWriter{
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

func (sh *InsShellWriter) InsCp(triple pkg.InsTriple, meta *pkg.PackageMeta) error {
	if triple.Second == "" || triple.Third == "" {
		return errors.New("CP instruction must have src and des")
	}

	srcPath := meta.VendorSrcPath(sh.pkgHome) // todo: use shell variable as prefix
	if _, err := sh.writer.WriteString(fmt.Sprintf("mkdir -p \"%s\"\n", pkg.GetIncludePath(sh.pkgHome))); err != nil {
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
	srcPath := meta.VendorSrcPath(sh.pkgHome) // todo: use shell variable as prefix

	var configCmd = fmt.Sprintf("cmake -S \"%s\" -B \"%s\" -DCMAKE_BUILD_TYPE=Release -DCMAKE_INSTALL_PREFIX=\"%s\" %s",
		srcPath, pkg.GetCachePath(sh.pkgHome, meta.PackageName),
		pkg.GetPackagePkgPath(sh.pkgHome, meta.PackageName), triple.Second)
	var buildCmd = fmt.Sprintf("cmake --build \"%s\" --target install %s",
		pkg.GetCachePath(sh.pkgHome, meta.PackageName), triple.Third)
	if _, err := sh.writer.WriteString(fmt.Sprintf("cd \"%s\"\n%s\n%s\n", sh.pkgHome, configCmd, buildCmd)); err != nil {
		return err
	}
	return nil
}

func (sh *InsShellWriter) InsAutoPkg(triple pkg.InsTriple, meta *pkg.PackageMeta) error {
	return nil
}
