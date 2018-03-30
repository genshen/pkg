package utils

import "path/filepath"

const (
	PkgFileName   = "pkg.json"
	VendorName    = "vendor"
	VendorSrc     = "src"
	VendorPkg     = "pkg"
	VendorScripts = "scripts"
	VendorInclude = "include"
	VendorLib     = "lib"
	VendorLib64   = "lib64"
)

type Pkg struct {
	Command   map[string]string `json:"command"`
	Compilers map[string]string `json:"compilers"`
	Packages  Packages          `json:"packages"`
}

type Packages struct {
	GitPackages     map[string]GitPackage     `json:"git"`
	FilesPackages   map[string]FilesPackage   `json:"files"`
	ArchivePackages map[string]ArchivePackage `json:"archive"`
}

type Package struct {
	Path string `json:"path"`
	//	Installed    bool              `json:"_"`
	//	Dependencies []string          `json:"dependencies"`
	Build []string `json:"build"`
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

func GetPackageSrcPath(base, packageName string) (path string) {
	return filepath.Join(base, VendorName, VendorSrc, packageName)
}

func GetPkgIncludePath(base string) (path string) {
	return filepath.Join(base, VendorName, VendorPkg, VendorInclude)
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
