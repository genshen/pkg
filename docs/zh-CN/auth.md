## pkg 的仓库身份认证
如果相关的依赖库需要登录认证权限，则需要配置相关的身份认证。  

这里以 gitlab 上（包括官方或者自己搭建的实例）存储的仓库为例，假设某个仓库是私有的，下面开展 pkg 身份认证的相关配置。  

1. 在 gitlab 用户的的 profile 页面的 Access Tokens 子页面，新建一个 Access Tokens。权限勾选 `read_repository` 即可。  
2. 在本地需要下载依赖包的环境中，编辑 pkg 的配置文件 `~/.pkg/pkg.config.yaml`，添加刚刚生成的 Access Token：
   ```yaml
   auth:
     git.private.org: # 私有仓库对地址对应的域名
       user: genshen # gitlab 的用户名
       token: ****** # 上面生成的 access token
   ```
3. 保存 pkg.config.yaml，然后在执行 pkg fetch 命令的时候，pkg 工具就会自动通过配置的 access token 去拉去私有的仓库代码。
