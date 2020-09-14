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
	cleanCommand.Runner = &clean
	cmds.AllCommands = append(cmds.AllCommands, cleanCommand)
}

type clean struct {
	home string
}

func (c *clean) PreRun() error {
	if c.home == "" {
		return errors.New("flag home is required")
	}
	return nil
}

func (c *clean) Run() error {
	cachePath := filepath.Join(c.home, pkg.VendorName, pkg.VendorCache)
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		log.Warnln("no cache directory found, no need to clean")
		return nil
	} else if err != nil {
		return err
	} else {
		if err := os.RemoveAll(cachePath); err != nil {
			return err
		} else {
			log.Info("clean cache directory `", cachePath, "` successfully")
		}
	}
	return nil
}
