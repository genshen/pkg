package install

import (
	"bufio"
	"errors"
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
func RunIns(pkgHome, pkgName, srcPath, ins string, verbose bool) error {
	ins = strings.Trim(ins, " ")
	triple, err := pkg.ParseIns(ins)
	if err != nil {
		return err
	}

	switch triple.First {
	case "CP":
		if triple.Second == "" || triple.Third == "" {
			return errors.New("CP instruction must have src and des")
		}
		// run copy.
		if err := runInsCopy(filepath.Join(srcPath, triple.Second), triple.Third); err != nil {
			return err
		}

	case "RUN":
		if triple.Second == "" || triple.Third == "" {
			return errors.New("RUN instruction must be a triple")
		}
		workDir := triple.Second // fixme path not contains space.
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
		// run the command
		if err := involveShell(pkgHome, workDir, triple.Third, verbose); err != nil {
			return err
		}
	case "CMAKE": // run cmake commands, format CMAKE {config args} {build args}
		packageCacheDir := pkg.GetCachePath(pkgHome, pkgName)
		// remove old work dir files.
		if _, err := os.Stat(packageCacheDir); err != nil {
			if !os.IsNotExist(err) {
				return err
			}
		} else {
			if err := os.RemoveAll(packageCacheDir); err != nil {
				return err
			}
		}
		// make dirs
		if err := os.MkdirAll(packageCacheDir, 0744); err != nil {
			return err
		}
		// create script
		var configCmd = fmt.Sprintf("cmake -S \"%s\" -B \"%s\" -DCMAKE_BUILD_TYPE=Release -DCMAKE_INSTALL_PREFIX=\"%s\" %s",
			srcPath, packageCacheDir, pkg.GetPackagePkgPath(pkgHome, pkgName), triple.Second)
		var buildCmd = fmt.Sprintf("cmake --build \"%s\" --target install %s", packageCacheDir, triple.Third)
		// todo user customized config
		if err := involveShell(pkgHome, pkgHome, configCmd, verbose); err != nil {
			return err
		}
		if err := involveShell(pkgHome, pkgHome, buildCmd, verbose); err != nil {
			return err
		}
	}

	return nil
}

func involveShell(pkgHome, workDir, script string, verbose bool) error {
	if verbose {
		log.Println("running [", script, "] in directory ", workDir)
	}

	cmd := exec.Command("sh", "-c", script) // todo only for linux OS or OSX.
	cmd.Dir = workDir
	cmakeBuildEnv := fmt.Sprintf("PKG_VENDOR_PATH=%s", pkg.GetVendorPath(pkgHome))
	cmd.Env = append(os.Environ(), cmakeBuildEnv)
	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Run(); err != nil {
		return err
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

// w: writer
// pkgHome: path of project
// packageSrcPath: path of the source code in user home
func WriteIns(w *bufio.Writer, pkgHome, pkgName, packageSrcPath, ins string) error {
	ins = strings.Trim(ins, " ")
	triple, err := pkg.ParseIns(ins)
	if err != nil {
		return err
	}

	if triple.Second != "" && triple.Third != "" {
		switch triple.First {
		case "CP":
			// run copy.
			if _, err := w.WriteString(fmt.Sprintf("mkdir -p \"%s\"\n", pkg.GetIncludePath(pkgHome))); err != nil {
				return err
			}
			if _, err := w.WriteString(fmt.Sprintf("cp -r \"%s\" \"%s\"\n",
				filepath.Join(packageSrcPath, triple.Second), triple.Third)); err != nil {
				return err
			}
		case "RUN":
			if _, err := w.WriteString(fmt.Sprintf("mkdir -p \"%s\"\ncd \"%s\"\n%s\n",
				triple.Second, triple.Second, triple.Third)); err != nil {
				return err
			}
		}
	}

	if triple.First == "CMAKE" {
		var configCmd = fmt.Sprintf("cmake -S \"%s\" -B \"%s\" -DCMAKE_BUILD_TYPE=Release -DCMAKE_INSTALL_PREFIX=\"%s\" %s",
			packageSrcPath, pkg.GetCachePath(pkgHome, pkgName), pkg.GetPackagePkgPath(pkgHome, pkgName), triple.Second)
		var buildCmd = fmt.Sprintf("cmake --build \"%s\" --target install %s",
			pkg.GetCachePath(pkgHome, pkgName), triple.Second)
		if _, err := w.WriteString(fmt.Sprintf("cd \"%s\"\n%s\n%s\n", pkgHome, configCmd, buildCmd)); err != nil {
			return err
		}
	}

	return nil
}
