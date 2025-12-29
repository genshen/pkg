package fetch

import (
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
// noCache: dont use the global cache
func determinePackageCacheStrategy(packageMeta pkg.PackageMeta, projectRoot string, noCache bool) (error, CacheStrategy) {
	pkgCachePath := packageMeta.HomeCacheSrcPath() // package global cache dir
	pkgVendorPath := packageMeta.VendorSrcPath(projectRoot)

	// copy only when global cache exists
	if _, err := os.Stat(pkgVendorPath); os.IsNotExist(err) { // vendor src does not exist.
		if _, err := os.Stat(pkgCachePath); err != nil { // global cache does not exist or other error.
			if os.IsNotExist(err) {
				return nil, CacheStrategyDownloadFromRemote
			}
			return err, CacheStrategySkip
		} else {         // vendor src does not exist, but global cache exist.
			if noCache { // if noCache, directly download. Don't use global cache.
				return err, CacheStrategyDownloadFromRemote
			}
			return nil, CacheStrategyCopyFromGlobalCache
		}
	} else if err != nil { // other error for vendor src stat
		return err, CacheStrategySkip
	} else { // vendor src exits.
		return nil, CacheStrategyUserLocalVendor
	}
}
