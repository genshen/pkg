package install

import (
	"bytes"
	"errors"
	"github.com/genshen/pkg"
	"runtime"
	"strconv"
	"text/template"
)

/* envs used in pkg template. */
var vars map[string]string;

const PKGROOT = "PKG_ROOT"

func init() {
	vars = make(map[string]string)
	vars["CORES"] = strconv.Itoa(runtime.NumCPU())
}

// pkgRoot: the root directory of pkg.yaml
func addVendorPathEnv(pkgRoot string) {
	vars[PKGROOT] = pkgRoot
	vendorPath := pkg.GetVendorPath(pkgRoot)
	vars["VENDOR_PATH"] = vendorPath
	vars["INCLUDE"] = pkg.GetIncludePath(pkgRoot)                 // vendor/include
}

// pkgRoot: the root directory of pkg.yaml
func addPathEnv(packageName string) error {
	if root, ok := vars[PKGROOT]; !ok {
		return errors.New("pkg root variable not set")
	} else {
		vars["CACHE"] = pkg.GetCachePath(root, packageName)        // vendor/cache/@pkg
		vars["PKG_DIR"] = pkg.GetPkgPath(root, packageName)        // vendor/pkg/@pkg
		vars["SRC_DIR"] = pkg.GetPackageSrcPath(root, packageName) // vendor/src/@pkg
		// todo vars["PKG_SRC"] = pkg.GetPackageSrcPath(root, packageName)
		vars["PKG_INC"] = pkg.GetPkgIncludePath(root, packageName) // vendor/pkg/@pkg/include
		// CMAKE_VENDOR_PATH_PKG
		vars["CMAKE_VENDOR_PATH_PKG"] = pkg.GetCMakeVendorPkgPath(packageName) // ${VENDOR_PATH}/pkg/@pkg
	}
	return nil
}

// replace origin string with args values.
func processEnv(origin string) string {
	t := template.New("o")
	t.Parse(origin)
	sb := bytes.NewBufferString("")
	t.Execute(sb, vars)
	return sb.String()
}
