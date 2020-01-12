package install

import (
	"fmt"
	"github.com/genshen/pkg"
	"github.com/genshen/pkg/conf"
	"os"
)

type PackageFetcher interface {
	setPackageMeta(pkgPath string, meta *pkg.PackageMeta) error
	fetch(auth map[string]conf.Auth, srcDes string, meta pkg.PackageMeta) error
}

type YamlGitPkgFetcher pkg.YamlGitPackage
type YamlFilesPkgFetcher pkg.YamlFilesPackage

// fetcher interface implementation for git package
// pkgPath:
func (git *YamlGitPkgFetcher) setPackageMeta(pkgPath string, meta *pkg.PackageMeta) error {
	meta.Version = git.Version
	meta.TargetName = git.Target
	meta.Features = git.Features
	meta.CMakeLib = git.CMakeLib
	meta.Builder = git.Build[:]
	if meta.CMakeLib == "" && len(meta.Builder) == 0 {
		// if user specified cmake lib and build commands are not set,
		// we set self cmake lib and self build commands.
		// Note: this self cmake lib and self build commands can be overwrite
		// by package specified self cmake lib and self build commands.
		meta.SelfCMakeLib = pkg.InsAutoPkg
		bs := []string{pkg.InsAutoPkg}
		meta.SelfBuild = bs[:]
	}

	// parse package path(name), target and version from key and gitPkg
	if err := meta.SetPackageName(pkgPath); err != nil {
		return err
	}
	return nil
}

func (git *YamlGitPkgFetcher) fetch(auth map[string]conf.Auth, srcDes string, meta pkg.PackageMeta) error {
	if git.Path == "" {
		git.Path = fmt.Sprintf("https://%s.git", meta.PackageName)
	}
	if err := gitSrc(auth, srcDes, meta.PackageName, git.Path, meta.Version); err != nil {
		_ = os.RemoveAll(srcDes)
		return err
	}
	return nil
}

// fetcher interface implementation for files package
func (files *YamlFilesPkgFetcher) setPackageMeta(pkgPath string, meta *pkg.PackageMeta) error {
	meta.PackageName = pkgPath
	meta.TargetName = ""
	meta.Version = "latest"
	meta.CMakeLib = files.CMakeLib
	meta.Builder = files.Build[:]
	return nil
}

func (files *YamlFilesPkgFetcher) fetch(auth map[string]conf.Auth, srcDes string, meta pkg.PackageMeta) error {
	if err := filesSrc(srcDes, meta.PackageName, files.Path, files.Files); err != nil {
		_ = os.RemoveAll(srcDes)
		return err
	}
	return nil
}

func gitPkgsToInterface(pkgYaml map[string]pkg.YamlGitPackage) map[string]PackageFetcher {
	fetchers := make(map[string]PackageFetcher)
	for k, p := range pkgYaml {
		temp := p
		fetchers[k] = (*YamlGitPkgFetcher)(&temp)
	}
	return fetchers
}

func filesPkgsToInterface(pkgYaml map[string]pkg.YamlFilesPackage) map[string]PackageFetcher {
	fetchers := make(map[string]PackageFetcher)
	for k, p := range pkgYaml {
		temp := p
		fetchers[k] = (*YamlFilesPkgFetcher)(&temp)
	}
	return fetchers
}
