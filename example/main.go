package main

import (
	"encoding/json"
	"fmt"

	"github.com/gorpher/fastrouter"
	"github.com/valyala/fasthttp"
)

func main() {
	a := fastrouter.NewRouter()
	a.Use(fastrouter.CorsHandler)
	a.Use(fastrouter.BasicAuth("golang", "siki"))
	a.Get("/", func(ctx *fasthttp.RequestCtx) {
		json.NewEncoder(ctx).Encode(a.Routers())
	})
	a.Get("/a", func(ctx *fasthttp.RequestCtx) {
		fmt.Fprintln(ctx, "success")
	})
	a.Get("/:a/:b", func(ctx *fasthttp.RequestCtx) {
		value := ctx.UserValue("a")
		value2 := ctx.UserValue("b")
		fmt.Fprintln(ctx, value, value2)
	})
	a.Get("/a/:y", func(ctx *fasthttp.RequestCtx) {
		fmt.Fprintln(ctx, ctx.UserValue("y"))
	})
	a.PrefixHandler("GET", "/a/:b/:c", func(ctx *fasthttp.RequestCtx) {
		fmt.Fprintln(ctx, "prefix mini")
	})
	a.Static("/a/", "/var/tmp/mini")
	fmt.Println("start: http://localhost:8080")
	err := fasthttp.ListenAndServe(":8080", a.Handler())
	if err != nil {
		panic(err)
	}
}
