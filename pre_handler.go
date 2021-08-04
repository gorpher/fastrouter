package fastrouter

import (
	"encoding/base64"
	"strings"

	"github.com/valyala/fasthttp"
)

func CorsHandler(ctx *fasthttp.RequestCtx) bool {
	origin := ctx.Request.Header.Peek("Origin")
	ctx.Response.Header.Set("Access-Control-Allow-Origin", strings.Join([]string{string(origin)}, ","))
	ctx.Response.Header.Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS,PATCH")
	ctx.Response.Header.Set("Access-Control-Allow-Headers", "*")
	ctx.Response.Header.Set("Access-Control-Expose-Headers", "*")
	ctx.Response.Header.Set("Access-Control-Allow-Credentials", "true")
	return !ctx.IsOptions()
}

// BasicAuth is the basic auth handler.
func BasicAuth(requiredUser, requiredPassword string) PreHandler {
	return func(ctx *fasthttp.RequestCtx) bool {
		auth := ctx.Request.Header.Peek("Authorization")
		if auth == nil {
			ctx.Error(fasthttp.StatusMessage(fasthttp.StatusUnauthorized), fasthttp.StatusUnauthorized)
			ctx.Response.Header.Set("WWW-Authenticate", "Basic realm=Restricted")
			return false
		}
		user, password, hasAuth := parseBasicAuth(string(auth))

		if hasAuth && user == requiredUser && password == requiredPassword {
			return true
		}
		ctx.Error(fasthttp.StatusMessage(fasthttp.StatusUnauthorized), fasthttp.StatusUnauthorized)
		ctx.Response.Header.Set("WWW-Authenticate", "Basic realm=Restricted")
		return false
	}
}

// parseBasicAuth parses an HTTP Basic Authentication string.
// "Basic Z29sYW5nOnNpa2k=" returns ("golang", "siki", true).
func parseBasicAuth(auth string) (username, password string, ok bool) {
	const prefix = "Basic "
	if !strings.HasPrefix(auth, prefix) {
		return
	}
	c, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
	if err != nil {
		return
	}
	s := string(c)
	i := strings.IndexByte(s, ':')
	if i < 0 {
		return
	}
	return s[:i], s[i+1:], true
}
