package export

import (
	"errors"
	"flag"
	"fmt"
	"github.com/genshen/cmds"
	"github.com/genshen/pkg"
	"github.com/mholt/archiver/v3"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"time"
)

var exportCommand = &cmds.Command{
	Name:        "export",
	Summary:     "export dependency packages export a tarball file",
	Description: "export dependency packages to a tarball file (.tar) specified by a file path",
	CustomFlags: false,
	HasOptions:  true,
}

func init() {
	var pwd string
	var err error
	if pwd, err = os.Getwd(); err != nil {
		pwd = "./"
	}

	var ex export
	fs := flag.NewFlagSet("export", flag.ExitOnError)
	exportCommand.FlagSet = fs
	exportCommand.FlagSet.StringVar(&ex.output, "out", "", "path to save tarball file")
	exportCommand.FlagSet.StringVar(&ex.home, "home", pwd, "path of home directory (where is "+pkg.PkgFileName+" file located)")
	exportCommand.FlagSet.Usage = exportCommand.Usage // use default usage provided by cmds.Command.
	exportCommand.Runner = &ex
	cmds.AllCommands = append(cmds.AllCommands, exportCommand)
}

// import
type export struct {
	output string
	home   string
	metas  map[string]pkg.PackageMeta
}

func (e *export) PreRun() error {
	if e.home == "" {
		return errors.New("flag home is required")
	}
	if e.output == "" {
		e.output = pkg.VendorName + "-" + time.Now().Format("20060102-150405.999999") + ".tar"
	}
	// file path check
	pkgFilePath := filepath.Join(e.home, pkg.PkgFileName)
	// check pkg.yaml file existence.
	if fileInfo, err := os.Stat(pkgFilePath); err != nil {
		return err
	} else if fileInfo.IsDir() {
		return fmt.Errorf("%s is not a file", pkg.PkgFileName)
	}

	// check sum file
	pkgSumPath := pkg.GetPkgSumPath(e.home)
	if fileInfo, err := os.Stat(pkgSumPath); err != nil {
		return fmt.Errorf(`stat file %s failed, make sure you have run "pkg install"; error: %s`, pkgSumPath, err)
	} else if fileInfo.IsDir() {
		return fmt.Errorf("%s is not a file", pkg.PkgFileName)
	}
	// resolve sum file.
	if err := pkg.DepTreeRecover(&e.metas, pkgSumPath); err != nil {
		return err
	}
	return nil
}

func (e *export) Run() error {
	tar := archiver.Tar{}

	tarFiles := [3]string{pkg.GetPkgSumPath(e.home), pkg.GetDepGraphPath(e.home), pkg.GetPkgSrcPath(e.home)}

	if err := tar.Archive(tarFiles[:], e.output); err != nil {
		return err
	}
	log.Info(fmt.Sprintf("export succeeded, file is saved at %s.", e.output))
	return nil
}
