package main

import (
	"net/http"

	"github.com/honmaple/forest"
	"github.com/honmaple/forest/middleware"
)

func main() {
	r := forest.New(forest.Debug())

	r.Use(middleware.Logger())
	r.Use(middleware.Recover())

	r.NotFound(func(c forest.Context) error {
		return c.JSON(444, forest.H{"message": "not found"})
	})

	// r.Use(middleware.BasicAuth(nil))
	r.GET("/*", func(c forest.Context) error {
		return c.HTML(http.StatusOK, "<h1>Hello Forest</h1>")

	})
	v1 := r.Group(forest.WithPrefix("/v1"), forest.WithName("v1"))
	{
		v1.GET("/posts/{title}", func(c forest.Context) error {
			return c.JSON(http.StatusOK, forest.H{"title": c.Param("title")})
		}).Named("vv")
		v1.GET("/posts/{titleId:int}", func(c forest.Context) error {
			return c.JSON(http.StatusOK, forest.H{"title": c.Param("titleId")})
		})
	}

	v2 := r.Group(forest.WithHost("v2.localhost:8000"))
	{
		v2.GET("/posts/{title:^test-\\w+}", func(c forest.Context) error {
			return c.JSON(http.StatusOK, forest.H{"title": c.Param("title")})
		})
	}

	v3 := newGroup()
	r.MountGroup(v3, forest.WithPrefix("/v3"))

	v2.Mount(v3, forest.WithPrefix("/v3"))

	r.Start("127.0.0.1:8000")
}

func newGroup() *forest.Group {
	r := forest.NewGroup()
	r.GET("/test", func(c forest.Context) error {
		return c.String(200, "test\n")
	})
	return r
}
