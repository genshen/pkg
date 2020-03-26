package install

import (
	"errors"
	"fmt"
	"github.com/genshen/pkg"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// run the instruction
type InsExecutor struct {
	BaseInsExecutor
	pkgHome string // home directory of running pkg command
	verbose bool   // flag to show building logs when running a command
}

func NewInsExecutor(pkgHome string, verbose bool, cmakeConfigArg, cmakeBuildArg string) *InsExecutor {
	return &InsExecutor{
		BaseInsExecutor: BaseInsExecutor{
			cmakeConfigArg: cmakeConfigArg,
			cmakeBuildArg:  cmakeBuildArg,
		},
		pkgHome: pkgHome,
		verbose: verbose,
	}
}

func (in *InsExecutor) Setup() error {
	return nil
}

func (in *InsExecutor) PkgPreInstall(meta *pkg.PackageMeta) (*pkg.PackageEnvs, error) {
	log.WithFields(log.Fields{
		"pkg": meta.PackageName,
	}).Info("installing package.")
	// package env
	packageEnv := pkg.NewPackageEnvs(in.pkgHome, meta.PackageName, meta.VendorSrcPath(in.pkgHome))
	return packageEnv, nil
}

func (in *InsExecutor) PkgPostInstall(meta *pkg.PackageMeta) error {
	log.WithFields(log.Fields{
		"pkg": meta.PackageName,
	}).Info("package built and installed.")
	return nil
}

func (in *InsExecutor) InsCp(triple pkg.InsTriple, meta *pkg.PackageMeta) error {
	if triple.Second == "" || triple.Third == "" {
		return errors.New("CP instruction must have src and des")
	}
	// run copy.
	srcPath := meta.VendorSrcPath(in.pkgHome)
	if err := runInsCopy(filepath.Join(srcPath, triple.Second), triple.Third); err != nil {
		return err
	}
	return nil
}

func (in *InsExecutor) InsRun(triple pkg.InsTriple, meta *pkg.PackageMeta) error {
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
	if err := involveShell(in.pkgHome, workDir, triple.Third, in.verbose); err != nil {
		return err
	}
	return nil
}

func (in *InsExecutor) InsCMake(triple pkg.InsTriple, meta *pkg.PackageMeta) error {
	packageCacheDir := pkg.GetCachePath(in.pkgHome, meta.PackageName)
	srcPath := meta.VendorSrcPath(in.pkgHome)
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
	// prepare cmake config and building command arguments
	if in.cmakeConfigArg != "" {
		triple.Second = triple.Second + " " + in.cmakeConfigArg
	}
	if in.cmakeBuildArg != "" {
		triple.Third = triple.Third + " " + in.cmakeBuildArg
	}
	// create script
	var configCmd = fmt.Sprintf("cmake -S \"%s\" -B \"%s\" -DCMAKE_BUILD_TYPE=Release -DCMAKE_INSTALL_PREFIX=\"%s\" %s",
		srcPath, packageCacheDir, pkg.GetPackagePkgPath(in.pkgHome, meta.PackageName), triple.Second)
	var buildCmd = fmt.Sprintf("cmake --build \"%s\" --target install %s", packageCacheDir, triple.Third)
	// todo user customized config
	if err := involveShell(in.pkgHome, in.pkgHome, configCmd, in.verbose); err != nil {
		return err
	}
	if err := involveShell(in.pkgHome, in.pkgHome, buildCmd, in.verbose); err != nil {
		return err
	}
	return nil
}

func (in *InsExecutor) InsAutoPkg(triple pkg.InsTriple, meta *pkg.PackageMeta) error {
	// if it is auto pkg and outer build mode
	if pkgEnvInc := os.Getenv("PKG_INNER_BUILD"); pkgEnvInc == "" {
		// use cmake instruction with features (features as cmake options)
		triple.First = pkg.InsCmake
		triple.Second = featuresToOptions(meta.Features)
		triple.Third = ""
		return in.InsCMake(triple, meta)
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

// writer: writer
// pkgHome: path of project
// packageSrcPath: path of the source code in user home
//func WriteIns(inst InsInterface, writer *bufio.Writer)
