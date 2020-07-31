package fetch

import (
	"errors"
	"fmt"
	"github.com/genshen/pkg"
	"github.com/genshen/pkg/conf"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/storer"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

// download source code packages.
// files: just download files specified by map files.
func filesSrc(srcDes, packageName, baseUrl string, files map[string]string) error {
	// check packageName dir, if not exists, then create it.
	if err := os.MkdirAll(srcDes, 0744); err != nil {
		return err
	}

	// download files:
	for k, file := range files {
		log.WithFields(log.Fields{
			"pkg":     packageName,
			"storage": filepath.Join(srcDes, file),
		}).Info("downloading dependencies.")
		res, err := http.Get(pkg.UrlJoin(baseUrl, k))
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
			log.WithFields(log.Fields{
				"pkg": packageName,
			}).Info("downloaded dependencies.")
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

	log.WithFields(log.Fields{
		"pkg":     packageName,
		"storage": srcPath,
	}).Info("downloading dependency package.")

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
	log.WithFields(log.Fields{
		"pkg": packageName,
	}).Info("downloaded dependency package.")

	// unzip
	log.WithFields(log.Fields{
		"pkg":     zipName,
		"storage": srcPath,
	}).Println("extracting package.")
	err = pkg.Unzip(zipName, srcPath)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"pkg":     zipName,
		"storage": srcPath,
	}).Println("finished extracting package.")
	return nil
}

// params:
// repositoryPrefix: the directory to store the git repo.
// packageCacheDir: cache location to store this package source.
// packagePath: package path.
// packageUrl:  package remote path, usually its a url.
// version: git commit hash or git tag or git branch.
func gitSrc(auths map[string]conf.Auth, packageCacheDir, packagePath, packageUrl, version string) error {
	if err := os.MkdirAll(packageCacheDir, 0744); err != nil {
		return err
	}

	// generate auth repository url.
	repoUrl := packageUrl
	if gitUrl, err := url.Parse(packageUrl); err != nil {
		return err
	} else {
		if hostAuth, ok := auths[gitUrl.Host]; ok {
			gitUrl.User = url.UserPassword(hostAuth.Username, hostAuth.Token)
			repoUrl = gitUrl.String()
		}
	}

	// init ReferenceName using branch and tag.
	var checkoutOpt git.CheckoutOptions
	// clone repository.
	if repos, err := git.PlainClone(packageCacheDir, false, &git.CloneOptions{
		URL:      repoUrl,
		Progress: os.Stdout,
		//ReferenceName: referenceName, // specific branch or tag.
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	}); err != nil {
		log.Println("Error here", err)
		return err
	} else { // clone succeed.
		// fetch all branches references from remote
		if err := repos.Fetch(&git.FetchOptions{
			RefSpecs: []config.RefSpec{"refs/*:refs/*", "HEAD:refs/heads/HEAD"},
		}); err != nil {
			return err
		}

		var found = false
		// check branches and tags
		if refIter, err := repos.Storer.IterReferences(); err != nil {
			return err
		} else {
			refIter := storer.NewReferenceFilteredIter(func(r *plumbing.Reference) bool {
				return r.Name().IsTag() || r.Name().IsBranch()
			}, refIter)

			for {
				if t, err := refIter.Next(); err != nil {
					if err == io.EOF {
						break
					} else {
						return err
					}
				} else {
					if (t.Name().String() == "refs/tags/"+version) || (t.Name().String() == "refs/heads/"+version) {
						checkoutOpt.Branch = t.Name()
						found = true
						break
					}
				}
			}
		}

		if !found {
			// checkout to hash, if hash is not empty, then checkout to some commit.
			checkoutOpt.Hash = plumbing.NewHash(version)
			if checkoutOpt.Hash.IsZero() {
				return fmt.Errorf("invalid commit hash: %s", version)
			}
		}

		worktree, err := repos.Worktree()
		if err != nil {
			return err
		}
		log.WithFields(log.Fields{
			"pkg":     packagePath,
			"version": version,
		}).Println("checkout repository to reference.")
		// do checkout
		if err = worktree.Checkout(&checkoutOpt); err != nil {
			return err
		}
	}

	// remove .git directory.
	err := os.RemoveAll(filepath.Join(packageCacheDir, ".git"))
	if err != nil {
		return err
	}

	return nil
}
