package pkg

import (
	"errors"
	"os"
	"strings"
)

type Auth struct {
	Host     string
	Username string
	Token    string
}

const DEFAULT_AUTH_FILE = "pkg.auth-config.yaml"
const AuthEnvName = "PKG_AUTH"

const ErrorEnvFormat = "wrong format of env, format should be: <username1>?<token1>@example1.com:<username2>?<token2>@example2.com"

// parse git repo access auth file (see example directory for example file)
func ParseAuth(home string) ([]Auth, error) {
	// check env first
	if pkgAuthEnv := os.Getenv(AuthEnvName); pkgAuthEnv != "" {
		return parseAuthEnv(pkgAuthEnv) // todo also parse auth file (but env has high priority)
	}
	// todo parsing file
	return nil, nil
}

func parseAuthEnv(pkgAuthEnv string) ([]Auth, error) {
	var localAuths []Auth
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
					localAuths = append(localAuths, Auth{Username: preUsername[0], Token: preToken[0], Host: preToken[1]})
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
