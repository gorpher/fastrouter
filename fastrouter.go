package fastrouter

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/valyala/fasthttp"
)

type PreHandler func(ctx *fasthttp.RequestCtx) bool

type FastRouter struct {
	mu          sync.Mutex
	indexRoutes map[string][]*route
	routes      []*route
	NotFound    fasthttp.RequestHandler
	NotAllowed  fasthttp.RequestHandler
	Recover     func(ctx *fasthttp.RequestCtx, p interface{})
	preHandlers []PreHandler
}

func defaultRecover(ctx *fasthttp.RequestCtx, p interface{}) {
	ctx.Error(fmt.Sprintf("%v", p), http.StatusInternalServerError)
}

type route struct {
	deepPath        []string
	varsN           map[string]int
	prefix          string
	urlPath         string
	method          string
	allowMethods    map[string]struct{}
	isPrefixHandler bool
	preHandlers     []PreHandler
	handler         fasthttp.RequestHandler
}

const (
	SplitPathMAXSize = 100
	URLSep           = "/"
	PathMaxSize      = 8182
)

// splitPath 分割URL路径，最大分割100次,支持基本的路径分割清理，不支持去除dot符号.
func splitPath(s string) []string {
	n := SplitPathMAXSize
	sep := URLSep
	if len(s) == 0 {
		return []string{URLSep}
	}
	if s[0] != '/' {
		s = sep + s
	}
	a := make([]string, n)
	n--
	i := 0
	for i < n {
		m := strings.Index(s, sep)
		if m < 0 {
			break
		}
		if m != 0 && i > 0 {
			a[i-1] += s[:m]
		}
		if i > 0 && a[i-1] == sep {
			s = s[m+len(sep):]

			continue
		}
		a[i] = sep
		s = s[m+len(sep):]
		i++
	}
	if i > 0 && len(s) > 0 {
		a[i-1] += s
	}
	return a[:i]
}

func (a *FastRouter) serve(ctx *fasthttp.RequestCtx, v *route, method, urlPath string) (bool, bool) {
	deepPath := splitPath(urlPath)
	var m, p bool

	p = len(v.deepPath) == len(deepPath) &&
		((v.deepPath[len(v.deepPath)-1] == URLSep && deepPath[len(deepPath)-1] == URLSep) ||
			(v.deepPath[len(v.deepPath)-1] != URLSep && deepPath[len(deepPath)-1] != URLSep))
	if v.isPrefixHandler && len(v.deepPath) <= len(deepPath) {
		p = true
	}
	m = v.method == method
	var allows []string
	if len(v.allowMethods) != 0 {
		for key := range v.allowMethods {
			allows = append(allows, strings.ToUpper(key))
		}
		_, ok := v.allowMethods["OPTIONS"]
		if !ok {
			allows = append(allows, "OPTIONS")
		}
		ctx.Response.Header.Set("Allow", strings.Join(allows, ","))
	}

	if !(m && p) {
		return m, p
	}
	for key, index := range v.varsN {
		ctx.SetUserValue(key, deepPath[index][1:])
	}
	for j := range a.preHandlers {
		if !a.preHandlers[j](ctx) {
			return true, true
		}
	}
	for j := range v.preHandlers {
		if !v.preHandlers[j](ctx) {
			return true, true
		}
	}
	v.handler(ctx)
	return m, p
}

func (a *FastRouter) genRoute(method, urlPath string, isPrefixHandler bool,
	handler fasthttp.RequestHandler, preHandler ...PreHandler) route {
	deepPath := splitPath(urlPath)
	varN := map[string]int{}
	var prefix string
	for i := range deepPath {
		if deepPath[i] != "/" {
			if deepPath[i][1] == ':' {
				key := deepPath[i][2:]
				if v, ok := varN[key]; ok && v != 0 {
					panic(fmt.Sprintf("路由变量变量名重复：%s %s", urlPath, deepPath[i]))
				}
				varN[key] = i
			}
		}
		if len(varN) == 0 {
			prefix += deepPath[i]
		}
	}
	if len(prefix) == 0 {
		prefix = "/"
	}
	return route{
		deepPath:        deepPath,
		varsN:           varN,
		prefix:          prefix,
		method:          method,
		handler:         handler,
		urlPath:         urlPath,
		preHandlers:     preHandler,
		isPrefixHandler: isPrefixHandler,
		allowMethods:    map[string]struct{}{method: {}},
	}
}

func (a *FastRouter) handle(method string, urlPath string, isPrefixHandler bool,
	handler fasthttp.RequestHandler, preHandler ...PreHandler) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if urlPath == "" {
		panic("'URL Path' Cannot be an empty string")
	}
	if urlPath[0] != '/' {
		panic("'URL Path' must start with '/'")
	}
	r := a.genRoute(method, urlPath, isPrefixHandler, handler, preHandler...)
	h, ok := a.indexRoutes[r.prefix]
	if !ok {
		a.indexRoutes[r.prefix] = []*route{&r}
		a.routes = append(a.routes, &r)
		return
	}
	for i := range h {
		if len(h[i].deepPath) == len(r.deepPath) {
			if h[i].method == r.method {
				panic(fmt.Sprintf("route already exist : %s %s", r.urlPath, r.method))
			}
			h[i].allowMethods[r.method] = struct{}{}
		}
	}
	h = append(h, &r)
	a.indexRoutes[r.prefix] = h
	a.routes = append(a.routes, &r)
}

func (a *FastRouter) PrefixHandler(method string, prefixPath string,
	handler fasthttp.RequestHandler, preHandler ...PreHandler) {
	a.handle(method, prefixPath, true, handler, preHandler...)
}

func (a *FastRouter) Handle(method string, urlPath string, handler fasthttp.RequestHandler, preHandler ...PreHandler) {
	a.handle(method, urlPath, false, handler, preHandler...)
}

func (a *FastRouter) Post(urlPath string, handler fasthttp.RequestHandler, preHandler ...PreHandler) {
	a.handle(http.MethodPost, urlPath, false, handler, preHandler...)
}

func (a *FastRouter) Get(urlPath string, handler fasthttp.RequestHandler, preHandler ...PreHandler) {
	a.handle(http.MethodGet, urlPath, false, handler, preHandler...)
}

func (a *FastRouter) Patch(urlPath string, handler fasthttp.RequestHandler, preHandler ...PreHandler) {
	a.handle(http.MethodPatch, urlPath, false, handler, preHandler...)
}

func (a *FastRouter) Put(urlPath string, handler fasthttp.RequestHandler, preHandler ...PreHandler) {
	a.handle(http.MethodPut, urlPath, false, handler, preHandler...)
}

func (a *FastRouter) Head(urlPath string, handler fasthttp.RequestHandler, preHandler ...PreHandler) {
	a.handle(http.MethodHead, urlPath, false, handler, preHandler...)
}

func (a *FastRouter) Options(urlPath string, handler fasthttp.RequestHandler, preHandler ...PreHandler) {
	a.handle(http.MethodOptions, urlPath, false, handler, preHandler...)
}

func (a *FastRouter) Delete(urlPath string, handler fasthttp.RequestHandler, preHandler ...PreHandler) {
	a.handle(http.MethodDelete, urlPath, false, handler, preHandler...)
}

func (a *FastRouter) Connect(urlPath string, handler fasthttp.RequestHandler, preHandler ...PreHandler) {
	a.handle(http.MethodConnect, urlPath, false, handler, preHandler...)
}

func (a *FastRouter) Trace(urlPath string, handler fasthttp.RequestHandler, preHandler ...PreHandler) {
	a.handle(http.MethodTrace, urlPath, false, handler, preHandler...)
}

func (a *FastRouter) Any(urlPath string, handler fasthttp.RequestHandler, preHandler ...PreHandler) {
	a.handle(http.MethodGet, urlPath, false, handler, preHandler...)
	a.handle(http.MethodPost, urlPath, false, handler, preHandler...)
	a.handle(http.MethodPut, urlPath, false, handler, preHandler...)
	a.handle(http.MethodPatch, urlPath, false, handler, preHandler...)
	a.handle(http.MethodHead, urlPath, false, handler, preHandler...)
	a.handle(http.MethodOptions, urlPath, false, handler, preHandler...)
	a.handle(http.MethodDelete, urlPath, false, handler, preHandler...)
	a.handle(http.MethodConnect, urlPath, false, handler, preHandler...)
	a.handle(http.MethodTrace, urlPath, false, handler, preHandler...)
}

func (a *FastRouter) Static(prefixPath string, fileRootPath string) {
	fs := &fasthttp.FS{
		Root:               fileRootPath,
		GenerateIndexPages: true,
		PathRewrite: func(ctx *fasthttp.RequestCtx) []byte {
			// 由于默认的 今天文件会出现url重定向的问题，于是重写了静态文件路径。
			path := ctx.Path()
			hasTrailingSlash := len(path) > 0 && path[len(path)-1] == '/'
			prefixSize := len(prefixPath)
			if len(prefixPath) > 0 && prefixPath[len(prefixPath)-1] == '/' {
				prefixSize--
			}
			if len(path) >= prefixSize {
				path = path[prefixSize:]
			}
			if hasTrailingSlash {
				return path
			}

			return append(path, '/')
		},
		PathNotFound: func(ctx *fasthttp.RequestCtx) {
			ctx.NotFound()
		},
	}
	a.handle("GET", prefixPath, true, fs.NewRequestHandler())
}

func (a *FastRouter) Handler() func(ctx *fasthttp.RequestCtx) {
	return func(ctx *fasthttp.RequestCtx) {
		var m, p bool
		urlPath := string(ctx.Path())
		if len(urlPath) > PathMaxSize {
			ctx.SetStatusCode(fasthttp.StatusRequestURITooLong)
			return
		}
		method := string(ctx.Method())
		defer func() {
			if err := recover(); err != nil {
				if a.Recover != nil {
					a.Recover(ctx, err)
				}
				defaultRecover(ctx, err)
			}
		}()
		// 优先匹配明确的路由
		handlers, ok := a.indexRoutes[urlPath]
		if ok {
			// 明确的路由前缀都相同时，优先匹配先定义的路由。
			for i := range handlers {
				m, p = a.serve(ctx, handlers[i], method, urlPath)
				if m && p {
					return
				}
			}
		}
		var normalMaxDeep int
		var normalDeepLen int
		var prefixMaxDeep int
		var prefixDeepLen int

		// 没有确定前缀的路由，可能是变量路由，或前缀路由。
		for j := range a.routes {
			r := a.routes[j]
			if strings.HasPrefix(urlPath, r.prefix) {
				// 优先匹配前缀最相似的路由
				if len(r.prefix) > prefixMaxDeep && r.isPrefixHandler {
					prefixDeepLen = len(r.prefix)
					prefixMaxDeep = j
					continue
				}
				if len(r.prefix) > normalMaxDeep {
					normalDeepLen = len(r.prefix)
					normalMaxDeep = j
				}
			}
		}
		if prefixDeepLen > 0 {
			m, p = a.serve(ctx, a.routes[prefixMaxDeep], method, urlPath)
			if m && p {
				return
			}
		}
		if normalDeepLen > 0 {
			m, p = a.serve(ctx, a.routes[normalMaxDeep], method, urlPath)
			if m && p {
				return
			}
		}
		if !p {
			if a.NotFound != nil {
				a.NotFound(ctx)
				return
			}
			ctx.NotFound()
			return
		}
		if !m {
			if a.NotAllowed != nil {
				a.NotAllowed(ctx)
			}
			ctx.SetStatusCode(http.StatusMethodNotAllowed)
		}
	}
}

func (a *FastRouter) Routers() []string {
	var routers []string
	for key := range a.indexRoutes {
		for i := range a.indexRoutes[key] {
			handler := a.indexRoutes[key][i]
			routers = append(routers, handler.urlPath)
		}
	}
	return routers
}

func (a *FastRouter) Use(handler PreHandler) *FastRouter {
	a.preHandlers = append(a.preHandlers, handler)

	return a
}

func NewRouter() *FastRouter {
	return &FastRouter{
		indexRoutes: map[string][]*route{},
		routes:      []*route{},
		preHandlers: []PreHandler{},
		Recover:     defaultRecover,
	}
}
