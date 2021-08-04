# FastRouter

FastRouter是一个快速、轻量级的 [fasthttp](https://github.com/valyala/fasthttp) 路由✨。


[![GoDoc](https://pkg.go.dev/badge/github.com/gorpher/fastrouter)](https://pkg.go.dev/github.com/gorpher/fastrouter)
[![Build Status](https://api.travis-ci.org/gorpher/fastrouter.svg?branch=main&status=passed)](https://travis-ci.org/gorpher/fastrouter)


目前使用的http路由器大多都是使用`tree`数据结构实现路由功能。 而且很多路由集成了不必要的功能，导致依赖越亮越臃肿。

### 路由说明

三种路由：明确路由、前缀路由和变量路由

1. 明确路由

```go
a.Get("/a", nil)
a.Get("/b", nil)
a.Get("/d", nil) 
```

2. 前缀路由

```go
a.PrefixHandler("GET", "/a", nil)
a.PrefixHandler("GET", "/b", nil)
a.PrefixHandler("GET", "/b/:c", nil) 
```

3. 变量路由

```go
a.Get( "/:a", nil)
a.Get("/:a/:b", nil)
a.Get( "/:a/:b/", nil)
```

核心规则：

1. 明确的路由定义不能重复
2. 可以定义变量路由和前缀路由的子路由，有限匹配子路由。

### 中间件

1. 全局中间件

```go
a := fastrouter.NewRouter()
a.Use(fastrouter.CorsHandler)
```

2. 子路由中间件

```go 
a.Get("/", func(ctx *fasthttp.RequestCtx) {
    fmt.Fprintln(ctx, "hello world")
},fastrouter.CorsHandler)
```
