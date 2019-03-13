package export

import (
	"errors"
	"flag"
	"fmt"
	"github.com/genshen/cmds"
	"github.com/genshen/pkg"
	"github.com/mholt/archiver"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
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
	fs := flag.NewFlagSet("export", flag.ContinueOnError)
	exportCommand.FlagSet = fs
	exportCommand.FlagSet.StringVar(&ex.output, "out", "", "path to save tarball file")
	exportCommand.FlagSet.StringVar(&ex.home, "home", pwd, "path of home directory (where is "+pkg.PkgFileName+" file located)")
	exportCommand.FlagSet.Usage = exportCommand.Usage // use default usage provided by cmds.Command.
	exportCommand.Runner = &ex
	cmds.AllCommands = append(cmds.AllCommands, exportCommand)
}

// import
type export struct {
	output string // todo auto filename by time.
	home   string
}

func (e *export) PreRun() error {
	if e.output == "" {
		return errors.New("flag out is required")
	}
	if e.home == "" {
		return errors.New("flag home is required")
	}
	// file path check
	pkgFilePath := filepath.Join(e.home, pkg.PkgFileName)
	// check pkg.yaml file existence.
	if fileInfo, err := os.Stat(pkgFilePath); err != nil {
		return err
	} else if fileInfo.IsDir() {
		return fmt.Errorf("%s is not a file", pkg.PkgFileName)
	}
	return nil
}

func (e *export) Run() error {
	tar := archiver.Tar{}
	if err := tar.Archive([]string{
		pkg.GetPkgSumPath(e.home),
		pkg.GetSrcPath(e.home),
	}, e.output); err != nil {
		return err
	}
	log.Info(fmt.Sprintf("export succeeded, file is saved at %s.", e.output))
	return nil
}
