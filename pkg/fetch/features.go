package fetch

import (
	"fmt"
	"log"

	"github.com/genshen/pkg"
)

const FEATURE_DEBUG = true

const DefaultFeatureName = "default"

// true for keeping this package building.
// false for ignore this package building.
func checkOptionalPackageFeatureMatches(context pkg.PackageMeta, featPkgList []string) bool {
	if FEATURE_DEBUG {
		log.Printf("feature matching debug. package name: %s, candidate list: %v", context.PackageName, featPkgList)
	}

	if featPkgList == nil || len(featPkgList) == 0 {
		return false
	}
	for _, feat := range featPkgList {
		if feat == context.PackageName {
			return true
		}
	}
	return false
}

// activeFeatureOptionalPackages returns all active packages listed in allFeatures.
// First, it filters all active features in all available features,
// then the active packages in each active features is selected.
// Please note, active features is specified by cli flags.
func activeFeatureOptionalPackages(allFeatures map[string]pkg.YamlFeatures, activeFeatures []string) (error, []string) {
	featVisitMap := make(map[string]bool)
	if err, activePackages := dfsSearchAllFeaturePackages(allFeatures, featVisitMap, activeFeatures); err != nil {
		return err, nil
	} else {
		return nil, activePackages
	}
}

func dfsSearchAllFeaturePackages(allFeatures map[string]pkg.YamlFeatures, featVisitMap map[string]bool, activeFeatures []string) (error, []string) {
	if allFeatures == nil {
		return nil, nil
	}

	localActivePackages := make([]string, 0) // active package in current scope and deeper scope (specified by `needed`)

	for _, featName := range activeFeatures {
		if feat, ok := allFeatures[featName]; !ok {
			if featName == DefaultFeatureName {
				continue // If `default` feature is not specified in yaml file, it is also ok.
			} else {
				return fmt.Errorf("feature %s is an invalid feature", featName), nil
			}
		} else {
			// if this feature is not added before, add it.
			if _, ok2 := featVisitMap[featName]; !ok2 {
				featVisitMap[featName] = true
				// append packages in current level.
				localActivePackages = append(localActivePackages, feat.Deps...)
				if err, pkgList := dfsSearchAllFeaturePackages(allFeatures, featVisitMap, feat.Needs); err != nil {
					return err, nil
				} else {
					// append packages in deeper level.
					if pkgList != nil && len(pkgList) > 0 {
						localActivePackages = append(localActivePackages, pkgList...)
					}
				}
			}
		}
	}
	if FEATURE_DEBUG {
		log.Println("search active packages (active features and result packages):", activeFeatures, localActivePackages)
	}
	return nil, localActivePackages
}
