package utils

import (
	"strings"
	"path/filepath"
	"os"
	"io"
)

// run instruction.
func RunIns(installPath, srcPath, ins string) error {
	ins = strings.Trim(ins, " ")
	// todo rewrite
	insTriple := strings.SplitN(ins, " ", 3)
	if len(insTriple) == 3 {
		if insTriple[0] == "CP" { //todo copy
			var des string
			if insTriple[2] == "{INCLUDE}" {
				includeDir := GetIncludePath(installPath)
				if err := CheckDir(includeDir); err != nil { // check include dir exist.
					return err
				}
				des = filepath.Join(GetIncludePath(installPath), insTriple[1]) // copy with the same name.
			} else {
				des = filepath.Join(srcPath, insTriple[2]) //todo make sure parent dir exists.
			}
			// run copy.
			if err := runInsCopy(filepath.Join(srcPath, insTriple[1]), des); err != nil {
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
