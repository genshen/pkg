package _import

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/genshen/cmds"
	"github.com/genshen/pkg"
	"github.com/mholt/archiver/v4"
	cp "github.com/otiai10/copy"
	log "github.com/sirupsen/logrus"
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
		if err := pkg.CheckDirLists(vendorDir, homeSrcDir); err != nil {
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

	format := archiver.CompressedArchive{
		Compression: archiver.Gz{},
		Archival:    archiver.Tar{},
	}

	handle := func(ctx context.Context, f archiver.File) error {
		pkg.Unzip(f, importCache) // extract to importCache dir
		return nil
	}

	inp, err := os.Open(i.input)
	if err != nil {
		return err
	}
	defer inp.Close()

	if err := format.Extract(context.Background(), inp, nil, handle); err != nil {
		return err
	}
	// mv sum file
	if err := os.Rename(filepath.Join(importCache, pkg.PurgePkgSumFileName),
		pkg.GetPkgSumPath(i.home)); err != nil {
		return err
	}
	// move graph file
	if err := os.Rename(filepath.Join(importCache, pkg.DepGraph),
		pkg.GetDepGraphPath(i.home)); err != nil {
		return err
	}

	// check sum file
	pkgSumPath := pkg.GetPkgSumPath(i.home)
	if fileInfo, err := os.Stat(pkgSumPath); err != nil {
		return fmt.Errorf(`stat file %s failed, %s`, pkgSumPath, err)
	} else if fileInfo.IsDir() {
		return fmt.Errorf("%s is not a file", pkg.PkgFileName)
	}
	// resolve sum file.
	metas := make(map[string]pkg.PackageMeta)
	if err := pkg.DepTreeRecover(&metas, pkgSumPath); err != nil {
		return err
	}
	// mv packages
	for name, meta := range metas {
		if name == pkg.RootPKG {
			continue
		}
		// todo use pkg api to get from path
		packagePathFrom := filepath.Join(importCache, pkg.VendorSrc, meta.PackageName+"@"+meta.Version)
		targetSrcPath := meta.VendorSrcPath(i.home)
		// check source directory
		if _, err := os.Stat(packagePathFrom); err != nil {
			return err
		}
		// todo move dirs force option
		// move one by one
		if fileInfo, err := os.Stat(targetSrcPath); err != nil {
			if os.IsNotExist(err) { // directory not exists, can import.
				if err := cp.Copy(packagePathFrom, targetSrcPath); err != nil {
					return err
				}
				log.WithField("package", meta.PackageName).Info("import package success.")
			} else {
				return err
			}
		} else if !fileInfo.IsDir() { // if exists,but is not dir.
			return fmt.Errorf("%s is not a directory", targetSrcPath)
		} else {
			log.WithField("package", meta.PackageName).Warning("skip importing package, because the package already exists.")
		}
	}
	log.Info(fmt.Sprintf("import succeeded."))
	return nil
}
