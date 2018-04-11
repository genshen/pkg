package install

import (
	"os"
	"io"
	"log"
	"fmt"
	"flag"
	"errors"
	"net/http"
	"io/ioutil"
	"path/filepath"
	"encoding/json"
	"github.com/genshen/cmds"
	"github.com/genshen/pkg/utils"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

var getCommand = &cmds.Command{
	Name:        "install",
	Summary:     "install packages from existed file pkg.json",
	Description: "install packages(zip,cmake,makefile,.etc format) existed file pkg.json.",
	CustomFlags: false,
	HasOptions:  true,
}

var (
	pkgFilePath string
)

func init() {
	getCommand.Runner = &get{}
	fs := flag.NewFlagSet("install", flag.ContinueOnError)
	getCommand.FlagSet = fs
	getCommand.FlagSet.StringVar(&pkgFilePath, "p", "./", "base path of file pkg.json")
	getCommand.FlagSet.Usage = getCommand.Usage // use default usage provided by cmds.Command.
	cmds.AllCommands = append(cmds.AllCommands, getCommand)
}

type get struct{}

func (v *get) PreRun() error {
	jsonPath := filepath.Join(pkgFilePath, utils.PkgFileName)
	// check pkg.json file existence.
	if fileInfo, err := os.Stat(jsonPath); err != nil {
		return err
	} else if fileInfo.IsDir() {
		return fmt.Errorf("%s is not a file", utils.PkgFileName)
	}

	return nil
	// check .vendor  and some related directory, if not exists, create it.
	//return utils.CheckVendorPath(pkgFilePath)
}

func (v *get) Run() error {
	if pkgJsonPath, err := os.Open(filepath.Join(pkgFilePath, utils.PkgFileName)); err != nil { // open file
		return err
	} else {
		if bytes, err := ioutil.ReadAll(pkgJsonPath); err != nil { // read file contents
			return err
		} else {
			pkgs := utils.Pkg{}
			if err := json.Unmarshal(bytes, &pkgs); err != nil { // unmarshal json to struct
				return err
			}
			return v.install(pkgFilePath, &pkgs.Packages)
		}
	}
	return nil
}

/**
install a package to destination refer to installPath, including source code and installed files.
usually src files are located at 'vendor/src/PackageName/', installed files are located at 'vendor/pkg/PackageName/'.
installPath is where the file pkg.json is located.
*/
func (v *get) install(installPath string, packages *utils.Packages) error {
	//todo packages have dependencies.
	// todo check install.
	// download archive src package.
	for key, pkg := range packages.ArchivePackages {
		if err := v.archiveSrc(installPath, key, pkg.Path); err != nil {
			// todo roolback, clean src.
			return err
		} else {
			// if source code downloading succeed, then compile and install it;
			// besides, you can just use source code in your project (e.g. use cmake package in cmake project).
		}
	}
	// download files src, and install it.
	for key, pkg := range packages.FilesPackages {
		srcDes := utils.GetPackageSrcPath(installPath, key)
		if _, err := os.Stat(srcDes); os.IsNotExist(err) {
			if err := v.filesSrc(srcDes, key, pkg.Path, pkg.Files); err != nil {
				// todo roolback, clean src.
				return err
			} else {
				// if source code downloading succeed, then compile and install it;
				// besides, you can just use source code in your project (e.g. use cmake package in cmake project).
				// do post install
				if err := v.postInstall(installPath, key, pkg.Package.Build); err != nil {
					return err;
				}
			}
		} else if err != nil {
			return err
		} else {
			log.Printf("skiped %s in %s, because it already exists.\n", key, srcDes)
		}
	}
	// download git src and install it.
	for key, pkg := range packages.GitPackages {
		repositoryPrefix := utils.GetPackageSrcPath(installPath, key)
		// check directory, if not exists, then create it.
		if _, err := os.Stat(repositoryPrefix); os.IsNotExist(err) {
			if err := v.gitSrc(repositoryPrefix, key, pkg.Path, pkg.Hash, pkg.Branch, pkg.Tag); err != nil {
				// todo roolback, clean src.
				return err
			} else {
				// if source code downloading succeed, then compile and install it;
				// besides, you can just use source code in your project (e.g. use cmake package in cmake project).
				// do post install
				if err := v.postInstall(installPath, key, pkg.Package.Build); err != nil {
					return err;
				}
			}
		} else if err != nil {
			return err;
		} else {
			log.Printf("skiped %s in %s, because it already exists.\n", key, repositoryPrefix)
		}
		// install dependency for this package.
		if err := v.installSubDependency(repositoryPrefix); err != nil {
			return err;
		}
	}
	return nil
}

// install dependency in a dependency, installPath is the path of sub-dependency.
// todo circle detect
func (v *get) installSubDependency(installPath string) error {
	if pkgJsonPath, err := os.Open(filepath.Join(installPath, utils.PkgFileName)); err == nil { // pkg.json not exists.
		if bytes, err := ioutil.ReadAll(pkgJsonPath); err != nil { // read file contents
			return err
		} else {
			pkgs := utils.Pkg{}
			if err := json.Unmarshal(bytes, &pkgs); err != nil { // unmarshal json to struct
				return err
			}
			return v.install(installPath, &pkgs.Packages)
		}
	} else {
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	}
}

// download archived package source code to destination directory, usually its 'vendor/src/PackageName/'.
// installPath is where parent project's file pkg.json is located.
func (get *get) archiveSrc(des string, packageName string, path string) error {
	if err := os.MkdirAll(des, 0744); err != nil {
		return err
	}

	log.Printf("downloading %s to %s\n", packageName, des)

	res, err := http.Get(path)
	if err != nil {
		return err // todo fallback
	}
	if res.StatusCode >= 400 {
		return errors.New("Http response code is not ok.")
	}

	// save file.
	zipName := filepath.Join(des, packageName+".zip")
	if fp, err := os.Create(zipName); err != nil { //todo create dir if file includes father dirs.
		return err // todo fallback
	} else {
		if _, err = io.Copy(fp, res.Body); err != nil {
			return err // todo fallback
		}
	}
	log.Printf("downloaded %s to %s\n", packageName, des)

	// unzip
	log.Printf("unziping %s to %s\n", zipName, des)
	err = utils.Unzip(zipName, des)
	if err != nil {
		return err
	}
	log.Printf("finished unziping %s to %s\n", zipName, des)
	return nil
}

// files: just download files specified by map files.
func (get *get) filesSrc(des string, packageName string, baseUrl string, files map[string]string) error {
	// check packageName dir, if not exists, then create it.
	if err := os.MkdirAll(des, 0744); err != nil {
		return err
	}

	// download files:
	for k, file := range files {
		log.Printf("downloading %s to %s\n", packageName, filepath.Join(des, file))
		res, err := http.Get(utils.UrlJoin(baseUrl, k))
		if err != nil {
			return err // todo fallback
		}
		if res.StatusCode >= 400 {
			return errors.New("Http response code is not ok.")
		}
		// todo create dir
		if fp, err := os.Create(filepath.Join(des, file)); err != nil { //todo create dir if file includes father dirs.
			return err // todo fallback
		} else {
			if _, err = io.Copy(fp, res.Body); err != nil {
				return err // todo fallback
			}
			log.Printf("downloaded %s to %s\n", packageName, filepath.Join(des, file))
		}
	}

	return nil
}

// params:
//  gitPath:  package remote path, usually its a url.
//  hash: git commit hash.
//  branch: git branch.
//  tag:  git tag.
func (get *get) gitSrc(repositoryPrefix string, packageName, gitPath, hash, branch, tag string) error {
	if err := os.MkdirAll(repositoryPrefix, 0744); err != nil {
		return err
	}

	// init ReferenceName using branch and tag.
	var referenceName plumbing.ReferenceName
	if branch != "" { // checkout to a specify branch.
		log.Printf("cloning %s repository from %s to %s with branch: %s\n", packageName, gitPath, repositoryPrefix, branch)
		referenceName = plumbing.ReferenceName("refs/heads/" + branch)
	} else if tag != "" { // checkout to specify tag.
		log.Printf("cloning %s repository from %s to %s with tag: %s\n", packageName, gitPath, repositoryPrefix, tag)
		referenceName = plumbing.ReferenceName("refs/tags/" + tag)
	} else {
		log.Printf("cloning %s repository from %s to %s\n", packageName, gitPath, repositoryPrefix)
	}

	// clone repository.
	if repos, err := git.PlainClone(repositoryPrefix, false, &git.CloneOptions{
		URL:           gitPath,
		Progress:      os.Stdout,
		ReferenceName: referenceName, // specific branch or tag.
		//RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	}); err != nil {
		return err
	} else { // clone succeed.
		if hash != "" { // if hash is not empty, then checkout to some commit.
			worktree, err := repos.Worktree()
			if err != nil {
				return err
			}
			log.Printf("checkout %s repository to commit: %s\n", packageName, hash)
			// do checkout
			err = worktree.Checkout(&git.CheckoutOptions{
				Hash: plumbing.NewHash(hash),
			})
			if err != nil {
				return err
			}
		}

		// remove .git directory.
		err = os.RemoveAll(filepath.Join(repositoryPrefix, ".git"))
		if err != nil {
			return err
		}
	}
	return nil
}

func (get *get) postInstall(installPath string, packageName string, build []string) error {
	srcPath := utils.GetPackageSrcPath(installPath, packageName)
	log.Println("installing package ", packageName)
	for _, ins := range build {
		if err := utils.RunIns(installPath, srcPath, ins); err != nil {
			return err
		}
	}
	return nil
}
