package middleware

import (
	"github.com/honmaple/forest"
	"runtime/debug"
)

func Recover() forest.HandlerFunc {
	return func(c forest.Context) (err error) {
		logger := c.Logger()
		defer func() {
			if r := recover(); r != nil {
				if e, ok := r.(error); ok {
					logger.Errorln(string(debug.Stack()))
					logger.Errorln(e.Error())
				} else {
					logger.Errorf("%v", r)
				}
			}
		}()
		return c.Next()
	}
}
