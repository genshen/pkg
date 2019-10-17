package pkg

import (
	"bytes"
	"errors"
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
func AddVendorPathEnv(pkgRoot string) {
	vars[PKGROOT] = pkgRoot
	vendorPath := GetVendorPath(pkgRoot)
	vars["VENDOR_PATH"] = vendorPath
	vars["INCLUDE"] = GetIncludePath(pkgRoot) // vendor/include
}

// pkgRoot: the root directory of pkg.yaml
func AddPathEnv(packageName string, packageSrcPath string) error {
	if root, ok := vars[PKGROOT]; !ok {
		return errors.New("pkg root variable not set")
	} else {
		vars["CACHE"] = GetCachePath(root, packageName)        // vendor/cache/@pkg
		vars["PKG_DIR"] = GetPkgPath(root, packageName)        // vendor/pkg/@pkg
		vars["PKG_INC"] = GetPkgIncludePath(root, packageName) // vendor/pkg/@pkg/include
		// CMAKE_VENDOR_PATH_PKG
		vars["CMAKE_VENDOR_PATH_PKG"] = GetCMakeVendorPkgPath(packageName) // ${VENDOR_PATH}/pkg/@pkg
		// todo vars["PKG_SRC"] = pkg.GetPackageHomeSrcPath(root, packageName)
		vars["SRC_DIR"] = packageSrcPath
	}
	return nil
}

// replace origin string with args values.
func ProcessEnv(origin string) string {
	t := template.New("o")
	t.Parse(origin)
	sb := bytes.NewBufferString("")
	t.Execute(sb, vars)
	return sb.String()
}
