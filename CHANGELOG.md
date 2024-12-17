
<a name="v0.6.0"></a>
## [v0.6.0](https://github.com/genshen/pkg/compare/v0.6.0..v0.5.0)

> 2024-12-17

### Build

* **deps:** bump github.com/go-git/go-git/v5 from 5.5.2 to 5.11.0
* **go-module:** bump dependency packages to latest version
* **go-module:** rename github.com/mholt/archiver/v4 to github.com/mholt/archives
* **go-module:** bump package github.com/otiai10/copy to v1.14.0
* **go-module:** bump to github.com/mholt/archiver/v4
* **go-module:** bump min go version in go.mod to 1.19
* **go-module:** bump version of dependency packages

### Ci

* **github-action:** bump action upload-artifact, download-artifact to v4 and action-gh-release to v2
* **github-action:** bump go version in github-action to 1.22

### Docs

* **changelog:** update changelog for v0.6.0

### Feat

* migrate to go 1.16+: change package io/ioutil => io or os
* **fetch:** min pkg version constraint in yaml file and check it while fetching
* **fetch:** support `features`: group optional packages into `features` and enables/disable them
* **fetch:** add `optional` flag to package for skipping a package
* **fetch:** support to check and use system proxy to fetch file/archive/git packages
* **fetch:** add more types support for fetching archive package
* **fetch:** make find package options as a cli flag
* **fetch:** add option `NO_DEFAULT_PATH` for find_package when generating file pkg.dep.cmake
* **fetch:** let user select a package version when package conflict occurs
* **install:** add `-j` argument to allow parallel building
* **list:** add sub-command `list` for listing all packages information
* **version:** print git commit id and build time when running `pkg version`
* **version:** bump pkg version to 0.6.0 and file format version to 3

### Fix

* **export:** fix the nil go context passed to `archives.FilesFromDisk` for exporting package archive
* **fetch:** add the missing package copying after remote package downloading finished
* **fetch:** fix incorrect branch and return value in cache determining algorithm
* **install:** keep the work dir before running instruction
* **install:** fix incorrect trim of quotation mark in install instruction
* **version:** keep format version checking backward compatibility to version 1

### Merge

* Merge pull request [#25](https://github.com/genshen/pkg/issues/25/) from genshen/bump-go-packages
* Merge pull request [#21](https://github.com/genshen/pkg/issues/21/) from genshen/ci-bump-artifact-actions-and-release-action
* Merge pull request [#22](https://github.com/genshen/pkg/issues/22/) from genshen/feature-archive-package-more-types-support
* Merge pull request [#23](https://github.com/genshen/pkg/issues/23/) from genshen/feature-fetch-via-system-proxy
* **export:** Merge pull request [#27](https://github.com/genshen/pkg/issues/27/) from genshen/fix-export-gen-archive-context-nil
* **fetch:** Merge pull request [#26](https://github.com/genshen/pkg/issues/26/) from genshen/feature-min-pkg-version-check
* **fetch:** Merge pull request [#24](https://github.com/genshen/pkg/issues/24/) from genshen/feature-fetch-by-features
* **fetch:** Merge pull request [#28](https://github.com/genshen/pkg/issues/28/) from genshen/fix-fetch-package-cache-strategy
* **fetch:** Merge pull request [#14](https://github.com/genshen/pkg/issues/14/) from genshen/feature-find-package-options
* **fetch:** Merge pull request [#11](https://github.com/genshen/pkg/issues/11/) from genshen/feature-fetch-check-packages-conflict
* **go-module:** Merge pull request [#15](https://github.com/genshen/pkg/issues/15/) from genshen/bump-go-1.19+
* **go-module:** Merge pull request [#12](https://github.com/genshen/pkg/issues/12/) from genshen/dependabot/go_modules/github.com/go-git/go-git/v5-5.11.0
* **install:** Merge pull request [#30](https://github.com/genshen/pkg/issues/30/) from genshen/fix-rm-workdir-before-instruction-run
* **install:** Merge pull request [#29](https://github.com/genshen/pkg/issues/29/) from genshen/fix-install-instruction-trim
* **install:** Merge pull request [#10](https://github.com/genshen/pkg/issues/10/) from genshen/feature-install-n-jobs
* **list:** Merge pull request [#20](https://github.com/genshen/pkg/issues/20/) from genshen/feature-package-list-subcommand
* **version:** Merge pull request [#32](https://github.com/genshen/pkg/issues/32/) from genshen/feature-version-with-commit-id
* **version:** Merge pull request [#31](https://github.com/genshen/pkg/issues/31/) from genshen/fix-config-format-version-check

### Refactor

* **fetch:** refactor the code of checking for global and vendor package cache

### Style

* code and comments formatting


<a name="v0.5.0"></a>
## [v0.5.0](https://github.com/genshen/pkg/compare/v0.5.0..v0.4.1)

> 2022-04-29

### Build

* **go-module:** bump version of dependency packages
* **go-module:** bump version of dependency packages
* **go-module:** update dependency packages, including go-git, logrus
* **go-module:** update package version of github.com/go-git/go-git

### Ci

* **github-action:** add github action config for building and releasing
* **makefile:** add darwin arm64 support

### Docs

* **changelog:** update changelog
* **changelog:** add changelog

### Feat

* improve error handling while parsing sub command in main.go
* **clean:** add clean sub-command for cleaning building cache files of dependency packages
* **fetch:** download package src using mirror specified by `git-replace` in pkg.yaml and pkg config
* **version:** bump version to 0.5.0

### Merge

* Merge pull request [#9](https://github.com/genshen/pkg/issues/9/) from genshen/ci-build-and-release


<a name="v0.4.1"></a>
## [v0.4.1](https://github.com/genshen/pkg/compare/v0.4.1..v0.4.0)

> 2020-07-24

### Build

* **docker:** update version of go,cmake,LibreSSL. And base image of pkg with mpi(for next version)
* **docker:** fix image building error for mpi version pkg (erroe of symbolic link: 'file exists')
* **docker:** update version of clang/cmake/go in dockerfile
* **go-module:** update go dependencies

### Feat

* **install:** specific cmake config and building arguments in command line of pkg install
* **version:** update version to 0.4.1


<a name="v0.4.0"></a>
## [v0.4.0](https://github.com/genshen/pkg/compare/v0.4.0..v0.3.3)

> 2020-03-25

### Feat

* **env:** add env PKG_FIND_PATH (vendor/deps/[@pkg](https://github.com/pkg/)(inner build) or vendor/pkg/[@pkg](https://github.com/pkg/)(outer build))
* **fetch:** add support of rendering `file pkg.dep.cmake` when AUTO_PKG is used
* **fetch:** add `features` into package metadata in yaml config file
* **fetch:** add parsing of AUTO_PKG instruction if builder commands and cmake lib are not specified
* **fetch:** add support for package fallback/default builder commands
* **install:** handle AUTO_PKG instruction of `pkg build` command: convert to CMAKE instruction
* **version:** update version to 0.4.0

### Fix

* **fetch:** remove find_package in cmake package template(for pkg.dep.cmake file) in inner building
* **fetch:** fix "if version does not match repo's tag/branch/hash, `fetch` still success"
* **fetch:** rename package name of fetch sub-command from `install` to `fetch`
* **install:** add missing AUTO_PKG instruction calling when executing building instructions

### Improvement

* **fetch:** use cmake variable as prefix(not absolute path) when generating pkg.dep.cmake
* **install:** use ${PROJECT_HOME} as path prefix when create building shell srcipt

### Refactor

* **fetch:** add DepsDir field to struct cmakeDepData as cmake binary dir for packages inner building
* **install:** move func RunIns from command.go to build_pkg.go
* **install:** refactor instruction env expanding: use local struct instead of global map
* **install:** move AutoPkg redirection to implemantation of interface InsInterface.InsAutoPkg
* **install:** add more installation interface functions: PkgPreInstall, PkgPostInstall
* **install:** refactor installation shell script generation and cmd building: use InsInterface


<a name="v0.3.3"></a>
## [v0.3.3](https://github.com/genshen/pkg/compare/v0.3.3..v0.3.2)

> 2020-01-10

### Chore

* **fetch:** add log before fetching and copies package, and correct some spelling mistakes

### Feat

* **fetch:** global cache strategy and global->vendor copy strategy
* **fetch:** only fetch package from remote if package doesn't exist in global cache and vendor src
* **fetch:** add feature of checkout to a branch or commit hash when cloning a git package
* **fetch:** add compatibility for fetching packages from pkg.yaml file version 1
* **fetch:** support parsing version and target from package path and package description
* **fetch:** redesign pkg.yaml file format, use path as package's name
* **install:** add cmake option and build option for CMAKE instruction
* **install:** remove cache directory before running cmake in CMAKE instruction
* **install:** add install instruction parser
* **install:** add CMAKE instruction to build cmake based dependencies
* **version:** update version to 0.3.3

### Fix

* **fetch:** fetching all branches from remote first to fix missing branch checking out
* **fetch:** fix `object not found` error when checking out to an annotated tag
* **install:** fix wrong condition of instruction triple when performing shell script generation

### Improvement

* **fetch:** cache fetched packages to user home and then copy to project vendor for building

### Merge

* Merge pull request [#6](https://github.com/genshen/pkg/issues/6/) from genshen/feature-package-path
* **install:** Merge pull request [#5](https://github.com/genshen/pkg/issues/5/) from genshen/feature-instruction-cmake

### Refactor

* **fetch:** remove recursive calling when generating pkg.dep.cmake file
* **fetch:** add PackageFetcher interface and implementation for files and git packages fetching

### BREAKING CHANGE


pkg.yaml file format is changed and use path as package's name.


<a name="v0.3.2"></a>
## [v0.3.2](https://github.com/genshen/pkg/compare/v0.3.2..v0.3.1)

> 2019-12-29

### Build

* **docker:** add Dockerfile of pkg docker image with mpi env
* **docker:** add necessary tools(cmake, make, clang toolchain) to docker image
* **docker:** add ENTRYPOINT for Dockerfile
* **go-module:** update yaml dependency to v3

### Feat

* **build:** generate build script with short path (using environment variable as base path)
* **config:** read authentication config from project config file and user home config file
* **export:** export dependencies packages in user home to tar file
* **fetch:** use yaml format as sum file
* **fetch:** remove srcPath in sum file and get srcPath by calling GetPackageHomeSrcPath
* **fetch:** remove package source files when threr is a fetching error
* **fetch:** decide a version when there are multiple versions for a package
* **fetch:** fetch dependencies to user home dir, not vendor dir.
* **import:** import packages to '.pkg' directory in user home from tar file
* **install:** use env 'PKG_VENDOR_PATH' to find installed libs in project vendor dir while building
* **version:** update version to 0.3.2

### Fix

* **build:** correct src path when writing CP instruction to shell srcipt
* **fetch:** check vendor directory in pre-run for fetch command
* **import:** check vendor and user-home src directory in pre-run for import command
* **vendor:** fix wrong returning of func GetPkgSumPath

### Style

* **build:** remove used parameters, variables or functions in vendor.go and envs.go


<a name="v0.3.1"></a>
## [v0.3.1](https://github.com/genshen/pkg/compare/v0.3.1..v0.3.0)

> 2019-09-29

### Build

* **docker:** update docker go images to 1.13.1

### Feat

* **fetch:** set the root package's name as "root", not empty.
* **fetch:** dump dependencies graph for all packages.
* **install:** build and install pacakges from dependencies graph file.
* **version:** update version to 0.3.1.

### Fix

* **install:** fix bug of package not found while installing a specific package.


<a name="v0.3.0"></a>
## [v0.3.0](https://github.com/genshen/pkg/compare/v0.3.0..v0.3.0-alpha)

> 2019-09-12

### Build

* **go-module:** update dependency packages.

### Chore

* update binary package name.
* remove build_all.sh file.

### Feat

* **install:** add --sh option for install subcommand to generate building shell script file.
* **install:** add "--dry" flag to generate cmake files only (not build lib).
* **version:** update version to v0.3.0.

### Merge

* Merge pull request [#2](https://github.com/genshen/pkg/issues/2/) from genshen/dev

### Refactor

* move cmds directory to pkg directory.

### BREAKING CHANGE


remove --dry and --skipdep option in install step; move cmake generation to fetch
step.


<a name="v0.3.0-alpha"></a>
## [v0.3.0-alpha](https://github.com/genshen/pkg/compare/v0.3.0-alpha..v0.2.0)

> 2019-06-12

### Build

* **go-module:** update dependency packages.

### Feat

* **fetch:** add http authentication support for git clone.
* **version:** update version to 0.3.0-alpha.

### Fix

* **cmds:** call log.Fatal when it has error (os.Exit(1)).
* **cmds:** change flag ErrorHandling from ContinueOnError to ExitOnError.
* **fetch:** fix the bug of opening file pkg.sum.json failed if not executing pkg command in pkg home.


<a name="v0.2.0"></a>
## [v0.2.0](https://github.com/genshen/pkg/compare/v0.2.0..v0.2.0-beta)

> 2019-03-14

### Build

* **go-mod:** update go dependency
* **makefile:** fix Makefile problem: use tab not space for commands.

### Chore

* **dockerfile:** add cmake pkg to docker image.

### Feat

* **export:** add automatic output file name for exported vendor tar.
* **init:** add init command
* **install:** add --verbose option for package installing.

### Merge

* Merge pull request [#1](https://github.com/genshen/pkg/issues/1/) from genshen/dev

### Style

* reorganize the code structural.


<a name="v0.2.0-beta"></a>
## [v0.2.0-beta](https://github.com/genshen/pkg/compare/v0.2.0-beta..v0.2.0-alpha)

> 2018-11-13

### Feat

* change pkg.json to pkg.yaml
* **install:** add feature of building and installing a specific package.
* **install:** add install env feature.

### Fix

* **fetch:** fix bug of reading empty command line argument in fetch sub-command.

### Refactor

* rename sub-command directories: fetch->load, install->fetch, build->install.
* move packages build fucntions to build sub-command.


<a name="v0.2.0-alpha"></a>
## [v0.2.0-alpha](https://github.com/genshen/pkg/compare/v0.2.0-alpha..v0.1.0)

> 2018-08-03

### Chore

* **example:** add an example of pkg.json

### Feat

* **install:** expand variable such as {PKG_DIR} to real path in generated dep.cmake.
* **install:** add more information to generated dep.cmake file.
* **install:** add package build and cmake generating after installed package.
* **install:** removed nesting vendor, now all dependency in the same director.

### Fix

* **install:** add hard coding of VENDOR_PATH in pkg.dep.cmake file to find vendor folder.
* **install:** fixed dependency lib cannot be found problem in pkg.dep.cmake file in v0.2.0.

### Refactor

* **install:** add dependency tree to build package.

### Revert

* **install:** fix some spell errors.


<a name="v0.1.0"></a>
## v0.1.0

> 2018-07-30

### Build

* **cross_platform:** add build script for windows/linux platform.

### Docs

* **README:** add README

### Feat

* **all:** first commit
* **docker:** add dockerfile
* **install:** add post install implement
* **vendor:** add dependency github.com/genshen/cmds

### Fix

* **docker:** update Dockerfile

