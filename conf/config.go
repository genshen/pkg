package conf

import (
	"github.com/genshen/pkg"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path/filepath"
)

const ConfigFileName = "pkg.config.yaml"

type PkgConfig struct {
	Auth map[string]Auth `yaml:"auth"`
}

func ParseConfig(projectHome string) (*PkgConfig, error) {
	var config PkgConfig
	// config in user home
	var userHomeConfig string
	if _conf, err := pkg.GetPkgUserHomeFile(ConfigFileName); err != nil {
		return nil, err
	} else {
		userHomeConfig = _conf
	}
	// project config file
	var projectConfig = filepath.Join(projectHome, ConfigFileName)

	for _, configFile := range []string{userHomeConfig, projectConfig} {
		if _, err := os.Stat(configFile); err != nil {
			if !os.IsNotExist(err) {
				return nil, err
			}
		} else { // file exists
			if confData, err := ioutil.ReadFile(configFile); err != nil {
				return nil, err
			} else {
				// auth map will be merged
				if err := yaml.Unmarshal(confData, &config); err != nil {
					return nil, err
				}
			}
		}
	}

	// parse auth from env
	if pkgAuthEnv := os.Getenv(AuthEnvName); pkgAuthEnv != "" {
		if auths, err := parseAuthEnv(pkgAuthEnv); err != nil {
			return nil, err
		} else {
			// merge env auth
			if config.Auth == nil {
				config.Auth = make(map[string]Auth)
			}
			for key, auth := range auths {
				config.Auth[key] = auth
			}
		}
	}
	return &config, nil
}
