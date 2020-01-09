package pkg

import (
	"os"
	"path/filepath"
)

const (
	VendorName        = "vendor"
	VendorCache       = "cache"
	VendorSrc         = "src"
	VendorPkg         = "pkg"
	VendorScripts     = "scripts"
	VendorInclude     = "include"
	VendorLib         = "lib"
	VendorLib64       = "lib64"
	VendorUserHome    = ".pkg"
	VendorUserHomeSrc = "registry/default-pkg/src"
)

const (
	PkgFileName         = "pkg.yaml"
	PurgePkgSumFileName = "pkg.sum.yaml"
	PkgSumFileName      = VendorName + "/" + PurgePkgSumFileName
	VendorSrcDir        = VendorName + "/" + "src"
	BuildShellName      = "pkg.build.sh"
	CMakeDep            = "pkg.dep.cmake"
	DepGraph            = "pkg.graph"
	CMakeVendorPath     = "${VENDOR_PATH}"
)

const RootPKG = "root"

func GetVendorPath(base string) string {
	return filepath.Join(base, VendorName)
}

func GetPkgBuildPath(base string) string {
	return filepath.Join(base, VendorName, BuildShellName)
}

func GetDepGraphPath(base string) string {
	return filepath.Join(base, VendorName, DepGraph)
}

func GetPkgSumPath(base string) string {
	return filepath.Join(base, PkgSumFileName)
}

func GetPkgSrcPath(base string) string {
	return filepath.Join(base, VendorSrcDir)
}

func getPackageVendorSrcPath(base string, packageName string, version string) string {
	return filepath.Join(base, VendorSrcDir, packageName+"@"+version)
}

func GetCachedPackageSrcPath(packageName string, version string) (string, error) {
	if path, err := GetPkgUserHomeFile(filepath.Join(VendorUserHomeSrc, packageName+"@"+version)); err != nil {
		return "", err
	} else {
		return path, nil
	}
}

// return @base/vendor/pkg/@packageName
func GetPackagePkgPath(base string, packageName string) (path string) {
	return filepath.Join(base, VendorName, VendorPkg, packageName)
}

// return ${VENDOR_PATH}/pkg/@packageName
func GetCMakeVendorPkgPath(packageName string) (path string) {
	return filepath.Join(CMakeVendorPath, VendorPkg, packageName)
}

// return @base/vendor/pkg/@packageName/include
func GetPkgIncludePath(base string, packageName string) (path string) {
	return filepath.Join(base, VendorName, VendorPkg, packageName, VendorInclude)
}

// return $HOME/.pkg/registry/default-pkg/src
func GetHomeSrcPath() (string, error) {
	if path, err := GetPkgUserHomeFile(VendorUserHomeSrc); err != nil {
		return "", err
	} else {
		return path, nil
	}
}

func GetPkgUserHomeFile(suffixPath string) (string, error) {
	if home, err := os.UserHomeDir(); err != nil {
		return "", err
	} else {
		return filepath.Join(home, VendorUserHome, suffixPath), nil
	}
}

// return @base/vendor/pkg
func GetPkgPath(base string) (path string) {
	return filepath.Join(base, VendorName, VendorPkg)
}

// return @base/vendor/include
func GetIncludePath(base string) (path string) {
	return filepath.Join(base, VendorName, VendorInclude)
}

// return @base/vendor/cache
func GetCachePath(base, packageName string) (path string) {
	return filepath.Join(base, VendorName, VendorCache, packageName)
}

//
//func CheckVendorPath(base string) error {
//	if err := CheckDirectoryLists(
//		filepath.Join(base, VendorName), // ".vendor"
//		//filepath.Join(base, VendorName, "bin"),     // ".vendor/bin"
//		//filepath.Join(base, VendorName, "sbin"),    // ".vendor/sbin"
//		//filepath.Join(base, VendorName, "include"), // ".vendor/include"
//		//filepath.Join(base, VendorName, "lib"),     // ".vendor/lib"
//		filepath.Join(base, VendorName, VendorSrc), // ".vendor/src"
//		//filepath.Join(base, VendorName, VendorScripts), // ".vendor/scripts"
//		filepath.Join(base, VendorName, VendorPkg), // ".vendor/pkg"
//	); err != nil {
//		return err
//	}
//	return nil
//}
