package main

import (
	"net/http"
	"net/http/pprof"
	_ "net/http/pprof"

	"github.com/honmaple/forest"
	"github.com/honmaple/forest/middleware"
)

func profile() *forest.Forest {
	r := forest.New()
	r.GET("/pprof/", forest.WrapHandlerFunc(pprof.Index))
	r.GET("/pprof/*", forest.WrapHandler(http.DefaultServeMux))
	r.POST("/pprof/symbol", forest.WrapHandlerFunc(pprof.Symbol))
	return r
}

func main() {
	r := forest.New(forest.Debug())
	r.Use(middleware.Recover())
	r.Use(middleware.Logger())
	r.GET("/", func(c forest.Context) error {
		return c.HTML(http.StatusOK, "<h1>PPROF</h1>")

	})
	r.Mount("/debug", profile())
	r.Start("127.0.0.1:8000")
}
