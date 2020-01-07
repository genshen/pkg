package pkg

// for pkg yaml file parsing
type YamlPkg struct {
	Version  int                 `yaml:"version"`
	Args     map[string]string   `yaml:"args"`
	PkgName  string              `yaml:"pkg"`
	Deps     YamlDependencies    `yaml:"dependencies"`
	Build    map[string][]string `yaml:"build"`
	CMakeLib string              `yaml:"cmake_lib"`
}

type YamlDependencies struct {
	GitPackages     map[string]YamlGitPackage     `yaml:"packages"`
	FilesPackages   map[string]YamlFilesPackage   `yaml:"files"`
	ArchivePackages map[string]YamlArchivePackage `yaml:"archives"`
}

type YamlPackage struct {
	Path             string   `yaml:"path"`     // todo this is not used
	Override         bool     `yaml:"override"` // override package self build.
	Build            []string `yaml:"build"`
	CMakeLib         string   `yaml:"cmake_lib"`
	CMakeLibOverride bool     `yaml:"cmake_lib_override"`
}

type YamlGitPackage struct {
	YamlPackage `yaml:",inline"`
	Version     string `yaml:"version"`
	Target      string `yaml:"target"`
}

type YamlFilesPackage struct {
	YamlPackage `yaml:",inline"`
	Files       map[string]string `yaml:"files"`
}

type YamlArchivePackage struct {
	YamlPackage `yaml:",inline"`
}
