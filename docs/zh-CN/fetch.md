
## fetch

## 缓存机制
缓存策略如下：
- 如果系统缓存和 vendor/src 下均没有包，会从互联网下载包，放到系统缓存然后拷贝到vencor/src 下。
- 如果仅在系统缓存中有包，则将系统缓存中的包拷贝到 vendor/src 下。
- 如果在vendor/src 下有包，则不会进行额外操作。即使缓存目录中没有相关的依赖包，使用 pkg fetch 命令也不会重新下载，也不会将依赖包代码拷贝到缓存目录。

注意：联网下载的依赖包会放到系统缓存，其位于 `~/.pkg/registry/default-pkg/src/`下。

## 使用代理
pkg 从v0.6.0 开始，支持使用 http 和 https 代理，只需要设置环境变量 `https_proxy` 和 `http_proxy` 即可，这样就可通过代理从网络上下载依赖包了。
```bash
export https_proxy=http://127.0.0.1:7890 http_proxy=http://127.0.0.1:7890
pkg fetch
```

## 选项
- `optional`: If the `optional` flag of a package is set to true, this package will not be downloaded and built.

## 命令行选项
- `no-cache`: 跳过本地全局缓存，直接从互联网上下载包。
- `features`: 选择启用的features，其中feature再 pkg.yaml 中的`features`定义。多个feature 用逗号分隔。
- `cmake-find-package-arg`: 如果一个包是CMake构建的，该选项可用。它表示这个包被添加到整个工程中，生成的CMake `find_package`语句的选项。
