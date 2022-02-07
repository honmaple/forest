package middleware

import (
	"github.com/honmaple/forest"
)

type NoCacheConfig struct {
	Skipper Skipper
	Headers map[string]string
}

var (
	DefaultNoCacheConfig = NoCacheConfig{
		Headers: map[string]string{
			"Cache-Control": "no-cache;no-store",
		},
	}
)

func NoCache() forest.HandlerFunc {
	return NoCacheWithConfig(DefaultNoCacheConfig)
}

func NoCacheWithConfig(config NoCacheConfig) forest.HandlerFunc {
	if len(config.Headers) == 0 {
		config.Headers = DefaultNoCacheConfig.Headers
	}
	return func(c forest.Context) error {
		if config.Skipper != nil && config.Skipper(c) {
			return c.Next()
		}
		header := c.Request().Header
		for k, v := range config.Headers {
			header.Set(k, v)
		}
		return c.Next()
	}
}
