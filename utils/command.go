package utils

import (
	"strings"
	"path/filepath"
	"os"
	"io"
	"os/exec"
	"log"
)

// run instruction.
func RunIns(pkgHome, packageName, srcPath, ins string) error {
	ins = strings.Trim(ins, " ")
	// todo rewrite
	insTriple := strings.SplitN(ins, " ", 3)

	if len(insTriple) == 3 {
		switch insTriple[0] {
		case "CP":
			var des string
			if insTriple[2] == "{INCLUDE}" {
				includeDir := GetIncludePath(pkgHome)
				if err := CheckDir(includeDir); err != nil { // check include dir exist.
					return err
				}
				des = filepath.Join(GetIncludePath(pkgHome), insTriple[1]) // copy with the same name.
			} else {
				des = filepath.Join(srcPath, insTriple[2]) //todo make sure parent dir exists.
			}
			// run copy.
			if err := runInsCopy(filepath.Join(srcPath, insTriple[1]), des); err != nil {
				return err
			}

		case "RUN":
			cacheDir := insTriple[1] // fixme path not contains space.
			if insTriple[1] == "{CACHE}" {
				cacheDir = GetCachePath(pkgHome, packageName)
			}
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
			script = strings.Replace(script, "{SRC_DIR}", srcPath, -1)
			script = strings.Replace(script, "{PKG_DIR}", GetPkgPath(pkgHome, packageName), -1)
			log.Println("running [", script, "] in directory", cacheDir)

			cmd := exec.Command("sh", "-c", script) // todo only for linux OS.
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
