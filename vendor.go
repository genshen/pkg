package pkg

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
	PkgFileName     = "pkg.yaml"
	PkgSumFileName  = VendorName + "/pkg.sum.json"
	BuildShellName  = "pkg.build.sh"
	CMakeDep        = "pkg.dep.cmake"
	DepGraph        = "pkg.graph"
	CMakeVendorPath = "${VENDOR_PATH}"
)

type Pkg struct {
	Version  int                 `yaml:"version"`
	Args     map[string]string   `yaml:"args"`
	Packages Packages            `yaml:"packages"`
	Build    map[string][]string `yaml:"build"`
	CMakeLib string              `yaml:"cmake_lib"`
}

type Packages struct {
	GitPackages     map[string]GitPackage     `yaml:"git"`
	FilesPackages   map[string]FilesPackage   `yaml:"files"`
	ArchivePackages map[string]ArchivePackage `yaml:"archive"`
}

type Package struct {
	Path     string `yaml:"path"`
	Override bool   `yaml:"override"` // override package self build.
	//	Dependencies []string          `yaml:"dependencies"`
	Build            []string `yaml:"build"`
	CMakeLib         string   `yaml:"cmake_lib"`
	CMakeLibOverride bool     `yaml:"cmake_lib_override"`
}

type GitPackage struct {
	Package `yaml:",inline"`
	Tag     string `yaml:"tag"`    // git tag
	Branch  string `yaml:"branch"` // git branch
	Hash    string `yaml:"hash"`   // git commit hash
}

type FilesPackage struct {
	Package `yaml:",inline"`
	Files   map[string]string `yaml:"files"`
}

type ArchivePackage struct {
	Package `yaml:",inline"`
}

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
	return filepath.Join(base, BuildShellName)
}

// return @base/vendor/src/@packageName
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

// return @base/vendor/src
func GetSrcPath(base string) (path string) {
	return filepath.Join(base, VendorName, VendorSrc)
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
