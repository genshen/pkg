package conf

import (
	"errors"
	"strings"
)

type Auth struct {
	Username string `yaml:"user"`
	Token    string `yaml:"token"`
}

const AuthEnvName = "PKG_AUTH"

const ErrorEnvFormat = "wrong format of env, format should be: <username1>?<token1>@example1.com:<username2>?<token2>@example2.com"

func parseAuthEnv(pkgAuthEnv string) (map[string]Auth, error) {
	var localAuths = make(map[string]Auth)
	// format: <username>?<token>@example.com:<username2>?<token2>@example2.com
	auths := strings.Split(pkgAuthEnv, ":")
	if len(auths) == 0 {
		return nil, nil
	} else {
		for _, authItl := range auths {
			// split  username
			preUsername := strings.Split(authItl, "?")
			if len(preUsername) == 2 {
				//  split token and domain.
				preToken := strings.Split(preUsername[1], "@")
				if len(preToken) == 2 {
					localAuths[preToken[1]] = Auth{Username: preUsername[0], Token: preToken[0]}
				} else {
					return nil, errors.New(ErrorEnvFormat)
				}
			} else {
				return nil, errors.New(ErrorEnvFormat)
			}
		}
	}
	return localAuths, nil
}
