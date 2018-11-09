package install

import (
	"errors"
	"github.com/genshen/pkg/utils"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// download source code packages.
// files: just download files specified by map files.
func filesSrc(srcDes string, packageName string, baseUrl string, files map[string]string) error {
	// check packageName dir, if not exists, then create it.
	if err := os.MkdirAll(srcDes, 0744); err != nil {
		return err
	}

	// download files:
	for k, file := range files {
		log.Printf("downloading %s to %s\n", packageName, filepath.Join(srcDes, file))
		res, err := http.Get(utils.UrlJoin(baseUrl, k))
		if err != nil {
			return err // todo rollback
		}
		if res.StatusCode >= 400 {
			return errors.New("http response code is not ok (200)")
		}
		// todo create dir
		if fp, err := os.Create(filepath.Join(srcDes, file)); err != nil { //todo create dir if file includes father dirs.
			return err // todo fallback
		} else {
			if _, err = io.Copy(fp, res.Body); err != nil {
				return err // todo fallback
			}
			log.Printf("downloaded %s to %s\n", packageName, filepath.Join(srcDes, file))
		}
	}
	return nil
}

// download archived package source code to destination directory, usually its 'vendor/src/PackageName/'.
// srcPath is the src location of this package (vendor/src/packageName).
func archiveSrc(srcPath string, packageName string, path string) error {
	if err := os.MkdirAll(srcPath, 0744); err != nil {
		return err
	}

	log.Printf("downloading %s to %s\n", packageName, srcPath)

	res, err := http.Get(path)
	if err != nil {
		return err // todo fallback
	}
	if res.StatusCode >= 400 {
		return errors.New("http response code is not ok (200)")
	}

	// save file.
	zipName := filepath.Join(srcPath, packageName+".zip")
	if fp, err := os.Create(zipName); err != nil { //todo create dir if file includes father dirs.
		return err // todo fallback
	} else {
		if _, err = io.Copy(fp, res.Body); err != nil {
			return err // todo fallback
		}
	}
	log.Printf("downloaded %s to %s\n", packageName, srcPath)

	// unzip
	log.Printf("unziping %s to %s\n", zipName, srcPath)
	err = utils.Unzip(zipName, srcPath)
	if err != nil {
		return err
	}
	log.Printf("finished unziping %s to %s\n", zipName, srcPath)
	return nil
}

// params:
// gitPath:  package remote path, usually its a url.
// hash: git commit hash.
// branch: git branch.
// tag:  git tag.
func gitSrc(repositoryPrefix string, packageName, gitPath, hash, branch, tag string) error {
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
