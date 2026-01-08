# the `pkg` packager manager

`pkg` 是一款用于 C/C++ 的包管理工具，其基于源码和 git 进行依赖管理，基于 CMake 工具进行依赖的构建。

> 本文档适用于 pkg v0.3.0 及以上版本。

## 基本功能
具体功能包括：
- `pkg init` 子命令，用户初始化一个空的配置。
- `pkg fetch` 子命令，用于从远程/本地缓存拉取源代码。
- `pkg install` 子命令，用于构建从源代码编译库并安装。
- `pkg export` 子命令，用于将本项目的依赖打包，方便进行依赖的迁移。
- `pkg import` 子命令，用于导入`pkg export`所打包的依赖包。
- `pkg clean` 子命令，用于清除依赖包的编译缓存。
- `pkg version` 子命令，用于显示版本信息。

对应的子命令后加上`--help`参数，可以查看对应子命令对应的可用参数，如 `pkg fetch --help`。

## 配置文件
pkg 的配置文件采用 yaml 格式进行配置，默认文件名为`pkg.yaml`，具体可参考 example 目录下的文件。

- 代理功能
