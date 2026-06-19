package clean

import (
	"errors"
	"flag"
	"os"
	"path/filepath"

	"github.com/genshen/cmds"
	"github.com/genshen/pkg"
	log "github.com/sirupsen/logrus"
)

var cleanCommand = &cmds.Command{
	Name:    "clean",
	Summary: "clean building cache files of dependency packages",
	Description: "clean building cache files of dependency packages, which is usually located in " +
		pkg.VendorName + "/" + pkg.VendorCache,
	CustomFlags: false,
	HasOptions:  true,
}

func init() {
	var pwd string
	var err error
	if pwd, err = os.Getwd(); err != nil {
		pwd = "./"
	}

	var clean clean
	fs := flag.NewFlagSet("clear", flag.ExitOnError)
	cleanCommand.FlagSet = fs
	cleanCommand.FlagSet.StringVar(&clean.home, "home", pwd, "path of home directory")
	cleanCommand.FlagSet.Usage = cleanCommand.Usage // use default usage provided by cmds.Command.
	cleanCommand.FlagSet.BoolVar(&clean.cleanInstallDir, "install-dir", false, "clean the installation directory under `vendor/pkg` (default: false)")
	cleanCommand.FlagSet.BoolVar(&clean.cleanCacheDir, "cache-dir", true, "clean the cache directory under `vendor/cache`.")
	cleanCommand.FlagSet.BoolVar(&clean.cleanAll, "all", false, "clean the both installation directory and cache directory. (default: false)")
	cleanCommand.Runner = &clean
	cmds.AllCommands = append(cmds.AllCommands, cleanCommand)
}

type clean struct {
	home            string
	cleanAll        bool
	cleanCacheDir   bool
	cleanInstallDir bool
}

func (c *clean) PreRun() error {
	if c.home == "" {
		return errors.New("flag home is required")
	}
	return nil
}

func (c *clean) Run() error {
	if c.cleanAll || c.cleanCacheDir {
		cachePath := filepath.Join(c.home, pkg.VendorName, pkg.VendorCache)
		err := c.cleanDir("cache", cachePath)
		if err != nil {
			return err
		}
	}

	if c.cleanAll || c.cleanInstallDir {
		installPath := filepath.Join(c.home, pkg.VendorName, pkg.VendorPkg)
		err := c.cleanDir("install", installPath)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *clean) cleanDir(logDirName string, dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		log.Warnln("no cache directory found, no need to clean")
		return nil
	} else if err != nil {
		return err
	} else {
		if err := os.RemoveAll(dirPath); err != nil {
			return err
		} else {
			log.Info("clean ", logDirName, " directory `", dirPath, "` successfully")
			return nil
		}
	}
}
