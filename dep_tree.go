/**
 * created by genshen on 2018/11/10
 */

package pkg

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

const (
	DlStatusEmpty = iota
	DlStatusSkip
	DlStatusOk
)

type DependencyTree struct {
	Dependencies []*DependencyTree
	Context      PackageMeta
	DlStatus     int
	IsPkgPackage bool
}

// package metadata used in sum file.
type PackageMeta struct {
	PackageName string   `yaml:"pkg"`    // package name (usually it is a path)
	TargetName  string   `yaml:"target"` // cmake package name
	Features    []string `yaml:"features"`
	//	SrcPath      string   `yaml:"-"`
	Version      string   `yaml:"version"`
	Builder      []string `yaml:"builder"`        // outer builder (lib used by others, specified by others pkg)
	SelfBuild    []string `yaml:"self_build"`     // inner builder (shows how this package is built, specified in package's pkg.yaml file)
	CMakeLib     string   `yaml:"cmake_lib"`      // outer cmake script to add this lib.
	SelfCMakeLib string   `yaml:"self_cmake_lib"` // inner cmake script to add this lib.
}

func (ctx *PackageMeta) SetPackageName(key string) error {
	keySplit := strings.SplitN(key, "@", 3)
	var version = ""
	var target = ""
	if len(keySplit) == 1 {
		ctx.PackageName = keySplit[0]
	} else if len(keySplit) == 2 {
		ctx.PackageName = keySplit[0]
		version = keySplit[1]
	} else if len(keySplit) == 3 {
		ctx.PackageName = keySplit[0]
		version = keySplit[1]
		target = keySplit[2]
	} else {
		return fmt.Errorf("bad package key: %s", key)
	}
	// set version and target(optional)
	if ctx.Version == "" {
		if version != "" {
			ctx.Version = version
		} else {
			return fmt.Errorf("bad package key: %s(version is not specified)", key)
		}
	}
	if ctx.TargetName == "" {
		ctx.TargetName = target
	}
	return nil
}

// return directory path of cached source in user home
func (ctx *PackageMeta) HomeCacheSrcPath() string {
	if path, err := GetCachedPackageSrcPath(ctx.PackageName, ctx.Version); err != nil {
		log.Fatal(err) // todo raise error
		return ""
	} else {
		return path
	}
}

// return directory path of source in vendor
func (ctx *PackageMeta) VendorSrcPath(base string) string {
	return getPackageVendorSrcPath(base, ctx.PackageName, ctx.Version)
}

// marshal dependency tree content to a yaml file.
func (depTree *DependencyTree) Dump(filename string) error {
	metas := make(map[string]PackageMeta) // string is package name.

	err := depTree.TraversalDeep(func(node *DependencyTree) error {
		if _, ok := metas[node.Context.PackageName]; ok {
			return nil // the package have already been added to map.
		}
		metas[node.Context.PackageName] = node.Context
		return nil
	})
	if err != nil {
		return err
	}

	// buffer.WriteString()
	if content, err := yaml.Marshal(metas); err != nil { // marshal map to sum file of yaml format
		return err
	} else {
		if dumpFile, err := os.Create(filename); err != nil {
			return err
		} else {
			if _, err := dumpFile.Write(content); err != nil {
				return err
			}
		}
	}
	return nil
}

// list all dependencies packages name of a package by TraversalDeep.
func (depTree *DependencyTree) ListDepsName() ([]string, error) {
	// dump all its dependencies
	pkgTraversalFlag := make(map[string]bool)
	lists := make([]string, 0)

	pkgTraversalFlag[depTree.Context.PackageName] = true // skip the root package.
	err := depTree.TraversalDeep(func(tree *DependencyTree) error {
		if _, ok := (pkgTraversalFlag)[tree.Context.PackageName]; ok {
			return nil // skip
		}
		lists = append(lists, tree.Context.PackageName)
		pkgTraversalFlag[tree.Context.PackageName] = true
		return nil
	})
	if err != nil {
		return nil, err
	}
	return lists, nil
}

// list all dependencies packages of a package by TraversalDeep.
func (depTree *DependencyTree) ListDeps(skipRoot bool) ([] *DependencyTree, error) {
	// dump all its dependencies
	pkgTraversalFlag := make(map[string]bool)
	lists := make([]*DependencyTree, 0)

	if skipRoot {
		pkgTraversalFlag[depTree.Context.PackageName] = true // skip the root package.
	}
	err := depTree.TraversalDeep(func(tree *DependencyTree) error {
		if _, ok := (pkgTraversalFlag)[tree.Context.PackageName]; ok {
			return nil // skip
		}
		lists = append(lists, tree)
		pkgTraversalFlag[tree.Context.PackageName] = true
		return nil
	})
	if err != nil {
		return nil, err
	}
	return lists, nil
}

// recover the dependency tree from a yaml file.
// the result is saved in variable metas.
func DepTreeRecover(metas *map[string]PackageMeta, filename string) error {
	if depFile, err := os.Open(filename); err != nil { // file open error or not exists.
		return err
	} else {
		defer depFile.Close()
		if bytes, err := ioutil.ReadAll(depFile); err != nil { // read file contents
			return err
		} else {
			if err := yaml.Unmarshal(bytes, metas); err != nil { // unmarshal yaml to struct
				return err
			}
			return nil
		}
	}
}

// traversal all tree node with pre-order.
// if the return value of callback function is false, it will skip its children nodes.
func (depTree *DependencyTree) Traversal(callback func(*DependencyTree) bool) {
	if r := callback(depTree); r == false {
		return
	}
	// if this node has children
	if depTree.Dependencies == nil || len(depTree.Dependencies) == 0 {
		return
	} else {
		for _, d := range depTree.Dependencies {
			d.Traversal(callback)
		}
	}
}

// traversal all tree node with pre-order.
// if the return value of callback function is false, then the traversal will break.
func (depTree *DependencyTree) TraversalPreOrder(callback func(*DependencyTree) bool) bool {
	if r := callback(depTree); r == false {
		return false
	}
	// if this node has children
	if depTree.Dependencies == nil || len(depTree.Dependencies) == 0 {
		return true
	} else {
		for _, d := range depTree.Dependencies {
			if r := d.TraversalPreOrder(callback); r == false {
				return false
			}
		}
	}
	return true
}

// traversal all tree node(including the root node) by deep first strategy.
// if return value of callback is false, then the traversal will break.
func (depTree *DependencyTree) TraversalDeep(callback func(*DependencyTree) error) error {
	// if this node has children
	if depTree.Dependencies != nil && len(depTree.Dependencies) != 0 {
		for _, d := range depTree.Dependencies {
			if err := d.TraversalDeep(callback); err != nil {
				return err
			}
		}
	}
	return callback(depTree)
}
