package utils

import "path/filepath"

const (
	VendorName    = "vendor"
	VendorCache   = "cache"
	VendorSrc     = "src"
	VendorPkg     = "pkg"
	VendorScripts = "scripts"
	VendorInclude = "include"
	VendorLib     = "lib"
	VendorLib64   = "lib64"
)

const (
	PkgFileName     = "pkg.json"
	PkgSumFileName     = "pkg.sum.json"
	CMakeDep        = "pkg.dep.cmake"
	CMakeVendorPath = "${VENDOR_PATH}"
)

type Pkg struct {
	Command   map[string]string   `json:"command"`
	Compilers map[string]string   `json:"compilers"`
	Packages  Packages            `json:"packages"`
	Build     map[string][]string `json:"build"` // todo platform
	CMakeLib  string              `json:"cmake_lib"`
}

type Packages struct {
	GitPackages     map[string]GitPackage     `json:"git"`
	FilesPackages   map[string]FilesPackage   `json:"files"`
	ArchivePackages map[string]ArchivePackage `json:"archive"`
}

type Package struct {
	Path     string `json:"path"`
	Override bool   `json:"override"` // override package self build.
	//	Dependencies []string          `json:"dependencies"`
	Build            []string `json:"build"`
	CMakeLib         string   `json:"cmake_lib"`
	CMakeLibOverride bool     `json:"cmake_lib_override"`
}

type GitPackage struct {
	Package
	Tag    string `json:"tag"`    // git tag
	Branch string `json:"branch"` // git branch
	Hash   string `json:"hash"`   // git commit hash
}

type FilesPackage struct {
	Package
	Files map[string]string `json:"files"`
}

type ArchivePackage struct {
	Package
}

func GetVendorPath(base string) string {
	return filepath.Join(base, VendorName);
}

func GetPackageSrcPath(base, packageName string) (path string) {
	return filepath.Join(base, VendorName, VendorSrc, packageName)
}

// return @base/vendor/pkg/@packageName
func GetPkgPath(base string, packageName string) (path string) {
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
