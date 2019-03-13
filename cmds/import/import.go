package _import

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

var importCommand = &cmds.Command{
	Name:        "import",
	Summary:     "import dependency packages from tarball file",
	Description: "import and extract dependency packages from tarball file (.tar) specified by a file path",
	CustomFlags: false,
	HasOptions:  true,
}

func init() {
	var pwd string
	var err error
	if pwd, err = os.Getwd(); err != nil {
		pwd = "./"
	}

	var l _import
	fs := flag.NewFlagSet("import", flag.ContinueOnError)
	importCommand.FlagSet = fs
	importCommand.FlagSet.StringVar(&l.input, "input", "", "path of tarball file")
	importCommand.FlagSet.StringVar(&l.home, "home", pwd, "path of home directory (where is "+pkg.PkgFileName+" file located)")
	importCommand.FlagSet.Usage = importCommand.Usage // use default usage provided by cmds.Command.
	importCommand.Runner = &l
	cmds.AllCommands = append(cmds.AllCommands, importCommand)
}

// import
type _import struct {
	input string
	home  string
}

func (i *_import) PreRun() error {
	if i.input == "" {
		return errors.New("flag input is required")
	}
	if i.home == "" {
		return errors.New("flag home is required")
	}
	// file path check
	pkgFilePath := filepath.Join(i.home, pkg.PkgFileName)
	// check pkg.yaml file existence.
	if fileInfo, err := os.Stat(pkgFilePath); err != nil {
		return err
	} else if fileInfo.IsDir() {
		return fmt.Errorf("%s is not a file", pkg.PkgFileName)
	}
	return nil
}

func (i *_import) Run() error {
	tar := archiver.Tar{
		//OverwriteExisting: true,
		//ImplicitTopLevelFolder: true,
		MkdirAll: true,
	}
	if err := tar.Unarchive(i.input, pkg.GetVendorPath(i.home)); err != nil {
		return err
	}
	log.Info(fmt.Sprintf("import succeeded."))
	return nil
}
