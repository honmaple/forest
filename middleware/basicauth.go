package middleware

import (
	"fmt"
	"net/http"

	"github.com/honmaple/forest"
)

type (
	BasicAuthConfig struct {
		Skipper   Skipper
		Realm     string
		Validator func(string, string) bool
	}
)

var (
	DefaultBasicAuthConfig = BasicAuthConfig{
		Realm: "Restricted",
		Validator: func(user, pass string) bool {
			return false
		},
	}
)

func BasicAuth(valid func(string, string) bool) forest.HandlerFunc {
	return BasicAuthWithConfig(BasicAuthConfig{Validator: valid})
}

func BasicAuthWithConfig(config BasicAuthConfig) forest.HandlerFunc {
	if config.Realm == "" {
		config.Realm = DefaultBasicAuthConfig.Realm
	}
	if config.Validator == nil {
		config.Validator = DefaultBasicAuthConfig.Validator
	}
	return func(c forest.Context) error {
		if config.Skipper != nil && config.Skipper(c) {
			return c.Next()
		}
		user, pass, ok := c.Request().BasicAuth()
		if ok && config.Validator(user, pass) {
			return c.Next()
		}
		w := c.Response()
		w.Header().Add("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, config.Realm))
		w.WriteHeader(http.StatusUnauthorized)
		return nil
	}
}
