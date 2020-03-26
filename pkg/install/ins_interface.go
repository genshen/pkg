package install

import "github.com/genshen/pkg"

// instruction interface
type InsInterface interface {
	// setup the building
	Setup() error
	PkgPreInstall(meta *pkg.PackageMeta) (*pkg.PackageEnvs, error)
	PkgPostInstall(meta *pkg.PackageMeta) error
	// files copy
	InsCp(triple pkg.InsTriple, meta *pkg.PackageMeta) error
	// run a command
	InsRun(triple pkg.InsTriple, meta *pkg.PackageMeta) error
	// run cmake build
	InsCMake(triple pkg.InsTriple, meta *pkg.PackageMeta) error
	InsAutoPkg(triple pkg.InsTriple, meta *pkg.PackageMeta) error
}

// base instruction
type BaseInsExecutor struct {
	cmakeConfigArg string
	cmakeBuildArg  string
}
