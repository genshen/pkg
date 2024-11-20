package pkg

import (
	"fmt"
	"runtime"
)

// for pkg yaml file parsing
type YamlPkg struct {
	Version    int                 `yaml:"version"`
	Args       map[string]string   `yaml:"args"`
	GitReplace map[string]string   `yaml:"git-replace"`
	PkgName    string              `yaml:"pkg"`
	Packages   V1Packages          `yaml:"packages"` // for pkg file version 1
	Deps       YamlDependencies    `yaml:"dependencies"`
	Build      map[string][]string `yaml:"build"`
	CMakeLib   string              `yaml:"cmake_lib"`
}

type YamlDependencies struct {
	GitPackages     map[string]YamlGitPackage     `yaml:"packages"`
	FilesPackages   map[string]YamlFilesPackage   `yaml:"files"`
	ArchivePackages map[string]YamlArchivePackage `yaml:"archives"`
}

type YamlPackage V1Package

type YamlGitPackage struct {
	YamlPackage `yaml:",inline"`
	Version     string   `yaml:"version"`
	Target      string   `yaml:"target"`
	Features    []string `yaml:"features"`
}

type YamlFilesPackage struct {
	YamlPackage `yaml:",inline"`
	Files       map[string]string `yaml:"files"`
}

const DefaultArchiveFormatType = "zip"

type YamlArchivePackage struct {
	YamlPackage `yaml:",inline"`
	Type        string `yaml:"type"` // archive type, support: zip, tar.gz, tar.bz2, tar. Default: zip.
}

// for pkg file version 1.
type V1Packages struct {
	GitPackages   map[string]V1GitPackage   `yaml:"git"`
	FilesPackages map[string]V1FilesPackage `yaml:"files"`
}

type V1Package struct {
	Path             string   `yaml:"path"`
	Override         bool     `yaml:"override"` // override package self build.
	Build            []string `yaml:"build"`
	CMakeLib         string   `yaml:"cmake_lib"`
	CMakeLibOverride bool     `yaml:"cmake_lib_override"`
}

type V1GitPackage struct {
	V1Package `yaml:",inline"`
	Tag       string `yaml:"tag"`    // git tag
	Branch    string `yaml:"branch"` // git branch
	Hash      string `yaml:"hash"`   // git commit hash
}

type V1FilesPackage struct {
	V1Package `yaml:",inline"`
	Files     map[string]string `yaml:"files"`
}

// find builder by os. If builder[os] is not found, return a fallback builder.
func (yamlPkg *YamlPkg) FindBuilder() []string {
	if _build, ok := yamlPkg.Build[runtime.GOOS]; ok {
		return _build[:] // builder can be empty if specified
	}
	if _build, ok := yamlPkg.Build["fallback"]; ok {
		return _build[:] // builder can be empty if specified
	}
	return nil
}

func (v1 *V1Packages) MigrateToV2(d *YamlDependencies) error {
	if d.GitPackages == nil {
		d.GitPackages = make(map[string]YamlGitPackage)
	}
	if d.FilesPackages == nil {
		d.FilesPackages = make(map[string]YamlFilesPackage)
	}

	if v1.GitPackages != nil {
		for name, gitPkg := range v1.GitPackages {
			// and not found in v2
			if _, ok := d.GitPackages[name]; !ok {
				var version = ""
				if gitPkg.Tag != "" {
					version = gitPkg.Tag
				} else if gitPkg.Branch != "" {
					version = gitPkg.Branch
				} else if gitPkg.Hash != "" {
					version = gitPkg.Hash
				} else {
					return fmt.Errorf("package %s version(tag/branch/hash) is not specified", name)
				}

				d.GitPackages[name] = YamlGitPackage{
					YamlPackage: YamlPackage{
						Path:             gitPkg.Path,
						Override:         gitPkg.Override,
						Build:            gitPkg.Build,
						CMakeLib:         gitPkg.CMakeLib,
						CMakeLibOverride: gitPkg.CMakeLibOverride,
					},
					Version: version,
					Target:  name,
				}
			}
		}
	}
	if v1.FilesPackages != nil {
		for name, filePkg := range v1.FilesPackages {
			// and not found in v2
			if _, ok := d.FilesPackages[name]; !ok {
				d.FilesPackages[name] = YamlFilesPackage{
					YamlPackage: YamlPackage{
						Path:             filePkg.Path,
						Override:         filePkg.Override,
						Build:            filePkg.Build,
						CMakeLib:         filePkg.CMakeLib,
						CMakeLibOverride: filePkg.CMakeLibOverride,
					},
					Files: filePkg.Files,
				}
			}
		}
	}
	return nil
}
