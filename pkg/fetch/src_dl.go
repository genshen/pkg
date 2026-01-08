package fetch

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/genshen/pkg"
	"github.com/genshen/pkg/conf"
	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/config"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/storer"
	"github.com/go-git/go-git/v6/plumbing/transport"
	"github.com/mholt/archives"
	cp "github.com/otiai10/copy"
	log "github.com/sirupsen/logrus"
)

// getProxyOptionFromEnvVars returns proxy options from environment variables
// by checking `https_proxy` and `http_proxy`.
// todo: proxy username and password support.
// todo: check and use socks proxy.
func getProxyOptionFromEnvVars(reqUrl string) string {
	proxyUrl := ""
	if strings.HasPrefix(reqUrl, "https") {
		// case-sensitive
		proxyUrl = os.Getenv("https_proxy")
		if proxyUrl == "" {
			proxyUrl = os.Getenv("HTTPS_PROXY")
		}
	} else if strings.HasPrefix(reqUrl, "http") {
		proxyUrl = os.Getenv("http_proxy")
		if proxyUrl == "" {
			proxyUrl = os.Getenv("HTTP_PROXY")
		}
	}
	return proxyUrl
}

func getHttpClientProxy(reqUrl string) *url.URL {
	proxyUrl := getProxyOptionFromEnvVars(reqUrl)
	if proxyUrl == "" {
		return nil
	}

	urli := url.URL{}
	if proxy, err := urli.Parse(proxyUrl); err != nil {
		return nil
	} else {
		return proxy
	}
}

// download source code packages.
// files: just download files specified by map files.
func filesSrc(srcDes, packageName, baseUrl string, files map[string]string) error {
	// create temp dir for saving downloaded files.
	tempPath, err := pkg.MakeGlobalPackageSrcDlTempPath()
	if err != nil {
		return err
	}
	log.WithFields(log.Fields{"pkg": packageName, "temp path": tempPath}).
		Debugln("downloading dependency to temporary directory.")

	// setup proxy if possible
	proxyUrl := getHttpClientProxy(baseUrl)
	if proxyUrl != nil {
		log.WithFields(log.Fields{"pkg": packageName, "proxy": proxyUrl}).
			Println("use proxy for package downloading.")
	}
	client := http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		},
	}

	// download files:
	for k, file := range files {
		log.WithFields(log.Fields{
			"pkg":     packageName,
			"storage": filepath.Join(tempPath, file),
		}).Info("downloading dependencies.")
		res, err := client.Get(pkg.UrlJoin(baseUrl, k))
		if err != nil {
			return err // todo rollback
		}
		if res.StatusCode >= 400 {
			return errors.New("http response code is not ok (200)")
		}
		// todo create dir
		if fp, err := os.Create(filepath.Join(tempPath, file)); err != nil { //todo create dir if file includes father dirs.
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

	// move dir from temp dir to real source file location in postDownloadStep.
	if err := postDownloadStep(packageName, tempPath, srcDes); err != nil {
		return err
	}
	return nil
}

// archivePackageFilename returns the archive file storage path by given
// srcPath (base path, e.g. $cache/src/packageName) and packageName and archiveType.
func archivePackageFilepath(srcPath, packageName, archiveType string) string {
	packageName = strings.ReplaceAll(packageName, "/", "_")
	packageName = strings.ReplaceAll(packageName, string(filepath.Separator), "_")
	return filepath.Join(srcPath, packageName+"."+archiveType)
}

// download archived package source code to destination directory, usually its 'vendor/src/PackageName/'.
// srcPath is the src location of this package ($cache/src/packageName).
func archiveSrc(archiveType string, srcPath string, packageName string, remoteUrl string) error {
	tempPath, err := pkg.MakeGlobalPackageSrcDlTempPath()
	if err != nil {
		return err
	}
	log.WithFields(log.Fields{"pkg": packageName, "temp path": tempPath}).
		Debugln("downloading dependency to temporary directory.")

	// setup proxy if possible
	proxyUrl := getHttpClientProxy(remoteUrl)
	if proxyUrl != nil {
		log.WithFields(log.Fields{"pkg": packageName, "proxy": proxyUrl}).
			Println("use proxy for package downloading.")
	}
	client := http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		},
	}

	res, err := client.Get(remoteUrl)
	if err != nil {
		return err // todo fallback
	}
	if res.StatusCode >= 400 {
		return errors.New("http response code is not ok (200)")
	}

	if archiveType == "" {
		archiveType = pkg.DefaultArchiveFormatType
		log.WithFields(log.Fields{
			"pkg": packageName,
		}).Info("set archive package format type to default type: %s.", archiveType)
	}

	// save file.
	zipName := archivePackageFilepath(tempPath, packageName, archiveType)
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
		"pkg":     packageName,
		"storage": tempPath,
	}).Println("extracting package.")

	// open archive file
	if f, err := os.Open(zipName); err != nil {
		return err
	} else {
		defer f.Close()

		// todo use archives.Identify(), instead of if-else.
		ac := archives.CompressedArchive{}
		if archiveType == "tar.bz2" {
			ac.Extraction = archives.Tar{}
			ac.Compression = archives.Bz2{}
		} else if archiveType == "tar.gz" {
			ac.Extraction = archives.Tar{}
			ac.Compression = archives.Gz{}
		} else if archiveType == "zip" {
			ac.Extraction = archives.Zip{}
		} else {
			return errors.New("unsupported type error")
		}

		handle := func(ctx context.Context, file archives.FileInfo) error {
			if err := pkg.Unzip(file, tempPath); err != nil {
				return err
			}
			return nil
		}

		if err := ac.Extract(context.Background(), f, handle); err != nil {
			return err
		}
	}

	log.WithFields(log.Fields{
		"pkg":     packageName,
		"storage": tempPath,
	}).Println("finished extracting package.")

	// move dir from temp dir to real source file location in postDownloadStep.
	if err := postDownloadStep(packageName, tempPath, srcPath); err != nil {
		return err
	}
	return nil
}

// params:
// repositoryPrefix: the directory to store the git repo.
// packageCacheDir: cache location to store this package source.
// packagePath: package path.
// packageUrl:  package remote path, usually its a url.
// version: git commit hash or git tag or git branch.
func gitSrc(auths map[string]conf.Auth, packageCacheDir, packagePath, packageUrl, version string) error {
	tempPath, err := pkg.MakeGlobalPackageSrcDlTempPath()
	if err != nil {
		return err
	}
	log.WithFields(log.Fields{"pkg": packagePath, "temp path": tempPath}).
		Debugln("downloading dependency to temporary directory.")

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

	// setup proxy if possible
	proxyUrl := getProxyOptionFromEnvVars(repoUrl)
	if proxyUrl != "" {
		log.WithFields(log.Fields{"pkg": packagePath, "proxy": proxyUrl}).
			Println("use proxy for package downloading.")
	}

	// init ReferenceName using branch and tag.
	var checkoutOpt git.CheckoutOptions
	// clone repository.
	if repos, err := git.PlainClone(tempPath, &git.CloneOptions{
		URL:      repoUrl,
		Progress: os.Stdout,
		//ReferenceName: referenceName, // specific branch or tag.
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		ProxyOptions: transport.ProxyOptions{
			URL: proxyUrl,
		},
	}); err != nil {
		log.Println("Error here", err)
		return err
	} else { // clone succeed.
		// fetch all branches references from remote
		if err := repos.Fetch(&git.FetchOptions{
			Force:    true,
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

		if err := postDownloadStep(packagePath, tempPath, packageCacheDir); err != nil {
			return err
		}
	}

	// remove .git directory.
	if err := os.RemoveAll(filepath.Join(packageCacheDir, ".git")); err != nil {
		return err
	}

	return nil
}

func postDownloadStep(packageName, tempPath, packageCacheDir string) error {
	// move temp dir to global home source dir
	log.WithFields(log.Fields{"pkg": packageName, "temp path": tempPath, "src path": packageCacheDir}).
		Debugln("move dependency from temporary directory to source path.")
	// create parent dir first and then perform move.
	if srcParentDir, err := pkg.GetCachedPackageSrcPath(packageName, ".draft"); err != nil {
		return err
	} else {
		if err := os.MkdirAll(srcParentDir, 0744); err != nil {
			return err
		}
		// remove the old package directory if possible
		if err := os.RemoveAll(packageCacheDir); err != nil {
			return err
		}

		// perform renaming: move from tem dir to the system cache dir.
		if err := os.Rename(tempPath, packageCacheDir); err != nil { // try rename
			if err := cp.Copy(tempPath, packageCacheDir); err != nil { // try copy and remove
				return err
			}
			// remove the dir
			if err := os.RemoveAll(tempPath); err != nil {
				return err
			}
		}
	}
	return nil
}
