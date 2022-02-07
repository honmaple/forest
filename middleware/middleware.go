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
