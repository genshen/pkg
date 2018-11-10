/**
 * created by genshen on 2018/11/10
 */

package utils

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

const (
	DlStatusEmpty = iota
	DlStatusSkip
	DlStatusOk
)

type DependencyTree struct {
	Context      DepPkgContext
	Dependencies []*DependencyTree
	Builder      []string // outer builder (lib used by others)
	SelfBuild    []string // inner builder (shows how this package is built)
	CMakeLib     string   // outer cmake script to add this lib.
	SelfCMakeLib string   // inner cmake script to add this lib.
	DlStatus     int
	IsPkgPackage bool
}

type DepPkgContext struct {
	Override         bool
	CMakeLibOverride bool
	PackageName      string
	SrcPath          string
}

// marshal dependency tree content to a json file.
func (depTree *DependencyTree) Dump(filename string) error {
	if content, err := json.Marshal(depTree); err != nil { // unmarshal json to struct
		return err
	} else {
		if dumpFile, err := os.Create(filename); err != nil {
			return err
		} else {
			dumpFile.Write(content)
		}
	}
	return nil
}

// recover the dependency tree from a json file.
// the result is saved in variable deps.
func DepTreeRecover(deps *DependencyTree, filename string) (error) {
	if depFile, err := os.Open(filename); err != nil { // file open error or not exists.
		return err
	} else {
		defer depFile.Close()
		if bytes, err := ioutil.ReadAll(depFile); err != nil { // read file contents
			return err
		} else {
			if err := json.Unmarshal(bytes, &deps); err != nil { // unmarshal json to struct
				return err
			}
			return nil
		}
	}
}


