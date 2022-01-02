package main

import (
	"net/http"

	"github.com/honmaple/forest"
	"github.com/honmaple/forest/middleware"
)

func main() {
	r := forest.New()
	r.Use(middleware.Recover())
	r.Use(middleware.Logger())
	r.GET("/", func(c forest.Context) error {
		return c.HTML(http.StatusOK, "<h1>Hello Forest</h1>")
	})

	v1 := r.Group("/v1")
	{
		v1.GET("/posts/{title}", func(c forest.Context) error {
			return c.JSON(http.StatusOK, forest.H{"title": c.Param("title")})
		})
		v1.GET("/posts/{titleId:int}", func(c forest.Context) error {
			return c.JSON(http.StatusOK, forest.H{"title": c.Param("titleId")})
		})
	}

	v2 := r.Host("v2.localhost:8000", "/v2")
	{
		v2.GET("/posts/{title:^test-\\w+}", func(c forest.Context) error {
			return c.JSON(http.StatusOK, forest.H{"title": c.Param("title")})
		})
	}

	v3 := newGroup()
	r.Mount("/v3", v3)

	r.Run("127.0.0.1:8000")
}

func newGroup() *forest.Group {
	r := forest.NewGroup()
	r.GET("/test", func(c forest.Context) error {
		return c.String(200, "test\n")
	})
	return r
}
