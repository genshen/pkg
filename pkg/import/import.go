package _import

import (
	"errors"
	"flag"
	"fmt"
	"github.com/genshen/cmds"
	"github.com/genshen/pkg"
	"github.com/mholt/archiver"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
)

const ImportCacheDir = ".import.cache"

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
	fs := flag.NewFlagSet("import", flag.ExitOnError)
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
	// file path check of pkg.yaml
	pkgFilePath := filepath.Join(i.home, pkg.PkgFileName)
	// check pkg.yaml file existence.
	if fileInfo, err := os.Stat(pkgFilePath); err != nil {
		return err
	} else if fileInfo.IsDir() {
		return fmt.Errorf("%s is not a file", pkg.PkgFileName)
	}

	// check src dir in user home, import src file here.
	// and check vendor dir
	if homeSrcDir, err := pkg.GetHomeSrcPath(); err != nil {
		return err
	} else {
		vendorDir := pkg.GetVendorPath(i.home)
		if err := pkg.CheckDirLists(vendorDir,homeSrcDir); err != nil {
			return err
		}
	}
	return nil
}

func (i *_import) Run() error {
	importCache := filepath.Join(pkg.GetVendorPath(i.home), ImportCacheDir)
	err := os.MkdirAll(importCache, 0744)
	if err != nil {
		return err
	}
	defer func() {
		if err := os.RemoveAll(importCache); err != nil {
			log.Warning(err)
		}
	}()

	tar := archiver.Tar{
		OverwriteExisting: true,
		//ImplicitTopLevelFolder: true,
		MkdirAll: true,
	}
	if err := tar.Unarchive(i.input, importCache); err != nil {
		return err
	}
	// mv sum file
	if err := os.Rename(filepath.Join(importCache, pkg.PurgePkgSumFileName),
		pkg.GetPkgSumPath(i.home)); err != nil {
		return err
	}
	// mv packages
	if fileInfo, err := ioutil.ReadDir(importCache); err != nil {
		return err
	} else {
		// todo move dirs force option
		srcRootPath, err := pkg.GetHomeSrcPath()
		if err != nil {
			return err
		}
		// move one by one
		for _, file := range fileInfo {
			targetSrcPath := filepath.Join(srcRootPath, file.Name())
			if fileInfo, err := os.Stat(targetSrcPath); err != nil {
				if os.IsNotExist(err) { // directory not exists, can import.
					if err := os.Rename(filepath.Join(importCache, file.Name()), targetSrcPath); err != nil {
						return err
					}
					log.WithField("package", file.Name()).Info("import package success.")
				} else {
					return err
				}
			} else if !fileInfo.IsDir() { // if exists,but is not dir.
				return fmt.Errorf("%s is not a directory", targetSrcPath)
			} else {
				log.WithField("package", file.Name()).Warning("skip importing package, because the package already exists.")
			}
		}
	}
	log.Info(fmt.Sprintf("import succeeded."))
	return nil
}
