package main

import (
	"fmt"

	"github.com/honmaple/forest"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

func main() {
	r := forest.New()

	r.GET("/", func(c forest.Context) error {
		return c.String(200, "OK\n")
	})
	fmt.Println(fasthttp.ListenAndServe("127.0.0.1:9000", fasthttpadaptor.NewFastHTTPHandler(r)))
}
