package fetch

import (
	"fmt"
	"os"

	"github.com/genshen/pkg"
)

const (
	CacheStrategyDownloadFromRemote  CacheStrategy = iota                            // package does not exist in global cache or project vendor dir. Need download from remote.
	CacheStrategyCopyFromGlobalCache CacheStrategy = iota                            // package exits at user home's global cache directory.
	CacheStrategyUserLocalVendor     CacheStrategy = iota                            // package exist in project's vendor/src directory.
	CacheStrategySkip                CacheStrategy = iota                            // skip package downloading or local copying
	CacheStrategyDefault             CacheStrategy = CacheStrategyDownloadFromRemote // package exist in project's vendor/src directory.
)

type CacheStrategy int

// determinePackageCacheStrategy determines package cache strategy via package filesystem path.
func determinePackageCacheStrategy(packageMeta pkg.PackageMeta, projectRoot string) (error, CacheStrategy) {
	pkgVendorPath := packageMeta.HomeCacheSrcPath() // package global cache dir
	pkgCachePath := packageMeta.VendorSrcPath(projectRoot)

	// copy only when global cache exists
	if _, err := os.Stat(pkgVendorPath); os.IsNotExist(err) { // vendor src does not exist.
		if _, err := os.Stat(pkgCachePath); err != nil { // global cache does not exist or other error.
			if os.IsNotExist(err) {
				return fmt.Errorf("cache and vendor path of package `%s` does not exists", packageMeta.PackageName), CacheStrategyDefault
			}
			return err, CacheStrategySkip
		} else { // vendor src does not exist, but global cache exist.
			return nil, CacheStrategyCopyFromGlobalCache
		}
	} else if err != nil { // other error for vendor src stat
		return err, CacheStrategySkip
	} else { // vendor src exits.
		return nil, CacheStrategyUserLocalVendor
	}
}
