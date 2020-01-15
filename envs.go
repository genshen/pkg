package pkg

import (
	"bytes"
	"reflect"
	"runtime"
	"strconv"
	"text/template"
)

const pkgTagName = "pkg"

// paths env variable used in instruction
type PackageEnvs struct {
	PkgRoot             string `pkg:"PKG_ROOT"`              // the path running pkg
	VendorPath          string `pkg:"VENDOR_PATH"`           // vendor
	PkgInCPath          string `pkg:"INCLUDE"`               // vendor/include
	PackageCacheDir     string `pkg:"CACHE"`                 // vendor/cache/@pkg
	PackagePkgDir       string `pkg:"PKG_DIR"`               // vendor/pkg/@pkg
	PackagePkgIncDir    string `pkg:"PKG_INC"`               // vendor/pkg/@pkg/include
	PackageSrcDir       string `pkg:"SRC_DIR"`               // vendor/src/@pkg
	CMakePackageFindDir string `pkg:"CMAKE_VENDOR_PATH_PKG"` // vendor/pkg/@pkg
}

// pkgRoot: the root directory of project
// packageName: package name/path
// packageSrcPath: path of package source
func NewPackageEnvs(pkgRoot, packageName, packageSrc string) *PackageEnvs {
	return &PackageEnvs{
		PkgRoot:             pkgRoot,
		VendorPath:          GetVendorPath(pkgRoot),
		PkgInCPath:          GetIncludePath(pkgRoot),
		PackageCacheDir:     GetCachePath(pkgRoot, packageName),
		PackagePkgDir:       GetPackagePkgPath(pkgRoot, packageName),
		PackagePkgIncDir:    GetPkgIncludePath(pkgRoot, packageName),
		PackageSrcDir:       packageSrc,
		CMakePackageFindDir: GetCMakeVendorPkgPath(packageName),
	}
}

// replace origin string with args values.
func ExpandEnv(origin string, envs *PackageEnvs) (string, error) {
	var vars = make(map[string]string)
	// add global envs
	vars["CORES"] = strconv.Itoa(runtime.NumCPU())
	// convert struct to map (key is the tag).
	t := reflect.TypeOf(*envs)
	v := reflect.ValueOf(*envs)
	for i := 0; i < t.NumField(); i++ {
		// Get the field tag value
		tag := t.Field(i).Tag.Get(pkgTagName)
		vars[tag] = v.Field(i).Interface().(string)
	}

	// template rendering
	if t, err := template.New("o").Parse(origin); err != nil {
		return "", err
	} else {
		sb := bytes.NewBufferString("")
		if err := t.Execute(sb, vars); err != nil {
			return "", err
		}
		return sb.String(), nil
	}
}
