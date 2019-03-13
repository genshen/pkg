package install

import (
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// run instruction.
func RunIns(pkgHome, packageName, srcPath, ins string) error {
	ins = strings.Trim(ins, " ")
	// todo rewrite
	insTriple := strings.SplitN(ins, " ", 3)

	if len(insTriple) == 3 {
		switch insTriple[0] {
		case "CP":
			// run copy.
			if err := runInsCopy(filepath.Join(srcPath, insTriple[1]), insTriple[2]); err != nil {
				return err
			}

		case "RUN":
			cacheDir := insTriple[1] // fixme path not contains space.
			// remove old files.
			if _, err := os.Stat(cacheDir); err != nil {
				if !os.IsNotExist(err) {
					return err
				}
			} else {
				os.RemoveAll(cacheDir)
			}

			// make dirs
			if err := os.MkdirAll(cacheDir, 0744); err != nil {
				return err
			}
			// create command
			script := insTriple[2]
			log.Println("running [", script, "] in directory", cacheDir)

			cmd := exec.Command("sh", "-c", script) // todo only for linux OS or OSX.
			cmd.Dir = cacheDir
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return err
			}
		}
	}
	return nil
}

func runInsCopy(target, des string) error {

	from, err := os.Open(target)
	if err != nil {
		return err
	}
	defer from.Close()

	to, err := os.OpenFile(des, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer to.Close()

	_, err = io.Copy(to, from)
	if err != nil {
		return err
	}
	return nil
}
