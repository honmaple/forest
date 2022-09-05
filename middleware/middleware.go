package middleware

import (
	"net/http"

	"github.com/honmaple/forest"
)

type (
	Skipper func(forest.Context) bool
)

func WrapHandler(h http.Handler) forest.HandlerFunc {
	return func(c forest.Context) error {
		h.ServeHTTP(c.Response(), c.Request())
		return c.Next()
	}
}

func Options() forest.HandlerFunc {
	return func(c forest.Context) error {
		if c.Request().Method == http.MethodOptions {
			return c.Status(http.StatusNoContent)
		}
		return c.Next()
	}
}

func Skip(m forest.HandlerFunc, skips ...Skipper) forest.HandlerFunc {
	return func(c forest.Context) error {
		for _, skip := range skips {
			if skip(c) {
				return c.Next()
			}
		}
		return m(c)
	}
}
