package install

import (
	"bufio"
	"fmt"
	"github.com/genshen/pkg"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// run instruction.
func RunIns(pkgHome, packageName, srcPath, ins string, verbose bool) error {
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
			workDir := insTriple[1] // fixme path not contains space.
			// remove old work dir files.
			if _, err := os.Stat(workDir); err != nil {
				if !os.IsNotExist(err) {
					return err
				}
			} else {
				if err := os.RemoveAll(workDir); err != nil {
					return err
				}
			}

			// make dirs
			if err := os.MkdirAll(workDir, 0744); err != nil {
				return err
			}
			// create command
			script := insTriple[2]
			if verbose {
				log.Println("running [", script, "] in directory ", workDir)
			}

			cmd := exec.Command("sh", "-c", script) // todo only for linux OS or OSX.
			cmd.Dir = workDir
			if verbose {
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
			}
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

func WriteIns(w *bufio.Writer, pkgHome, packageName, ins string) error {
	ins = strings.Trim(ins, " ")
	// todo rewrite
	insTriple := strings.SplitN(ins, " ", 3)

	if len(insTriple) == 3 {
		switch insTriple[0] {
		case "CP":
			// run copy.
			if _, err := w.WriteString(fmt.Sprintf("mkdir -p \"%s\"\n", pkg.GetIncludePath(pkgHome))); err != nil {
				return err
			}
			if _, err := w.WriteString(fmt.Sprintf("cp -r \"%s\" \"%s\"\n",
				filepath.Join(pkg.GetPackageSrcPath(pkgHome, packageName), insTriple[1]), insTriple[2])); err != nil {
				return err
			}
		case "RUN":
			if _, err := w.WriteString(fmt.Sprintf("mkdir -p \"%s\"\ncd \"%s\"\n%s\n",
				insTriple[1], insTriple[1], insTriple[2])); err != nil {
				return err
			}
		}
	}
	return nil
}
