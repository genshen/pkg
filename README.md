# pkg
> a c/c++ package manager based on source code.

## Usage
- Quick Start
```
# generate a new "package.yaml"
$ pkg init

# install a package,type can be "git", "tar", "files" for now. 
$ pkg fetch <type> <packagename>

# build and install packages from "package.yaml" file.
$ pkg install

# build and install a package specified by argument --pkg.
$ pkg install -pkg=<package_name>  # or: pkg install --pkg <package_name>
```
