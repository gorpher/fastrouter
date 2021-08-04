package fastrouter

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/valyala/fasthttp"
)

func TestSplitPath(t *testing.T) {
	tests := map[string][]string{
		// basic
		"/":        {"/"},
		"/a":       {"/a"},
		"/a/":      {"/a", "/"},
		"/a/b":     {"/a", "/b"},
		"/a/b/":    {"/a", "/b", "/"},
		"/a/b/c/d": {"/a", "/b", "/c", "/d"},

		// no root
		"":       {"/"},
		"a/b":    {"/a", "/b"},
		"a/b/c":  {"/a", "/b", "/c"},
		"a/b/c/": {"/a", "/b", "/c", "/"},

		// Remove doubled slash
		"//":         {"/"},
		"/z//":       {"/z", "/"},
		"/q/w//":     {"/q", "/w", "/"},
		"/a/b/c//":   {"/a", "/b", "/c", "/"},
		"/a//b//c":   {"/a", "/b", "/c"},
		"/a//b///c/": {"/a", "/b", "/c", "/"},
		"//a":        {"/a"},
		"///a":       {"/a"},
		"//a//":      {"/a", "/"},
	}
	for key := range tests {
		values := tests[key]
		l := splitPath(key)
		if len(l) != len(values) {
			t.Fatalf("parse [%s] error, want:%v ,but %v", key, values, l)
		}
		for i := range values {
			if values[i] != l[i] {
				t.Fatalf("parse [%s] error, want:%s ,but %s", key, values, l)
			}
		}
	}
}

func TestRouter(t *testing.T) {
	router := NewRouter()
	routed := false
	router.Handle("GET", "/user/:name", func(ctx *fasthttp.RequestCtx) {
		routed = true
		want := map[string]string{"name": "gopher"}

		if ctx.UserValue("name") != want["name"] {
			t.Fatalf("wrong wildcard values: want %v, got %v", want["name"], ctx.UserValue("name"))
		}
		ctx.Success("foo/bar", []byte("success"))
	})

	s := &fasthttp.Server{
		Handler: router.Handler(),
	}

	rw := &readWriter{}
	rw.r.WriteString("GET /user/gopher?baz HTTP/1.1\r\n\r\n")

	ch := make(chan error)
	go func() {
		ch <- s.ServeConn(rw)
	}()

	select {
	case err := <-ch:
		if err != nil {
			t.Fatalf("return error %s", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("timeout")
	}

	if !routed {
		t.Fatal("routing failed")
	}
}

func TestRouterAPI(t *testing.T) {
	var get, head, options, post, put, patch, deleted bool

	router := NewRouter()
	router.Get("/GET", func(ctx *fasthttp.RequestCtx) {
		get = true
	})
	router.Head("/GET", func(ctx *fasthttp.RequestCtx) {
		head = true
	})
	router.Options("/GET", func(ctx *fasthttp.RequestCtx) {
		options = true
	})
	router.Post("/POST", func(ctx *fasthttp.RequestCtx) {
		post = true
	})
	router.Put("/PUT", func(ctx *fasthttp.RequestCtx) {
		put = true
	})
	router.Patch("/PATCH", func(ctx *fasthttp.RequestCtx) {
		patch = true
	})
	router.Delete("/DELETE", func(ctx *fasthttp.RequestCtx) {
		deleted = true
	})

	s := &fasthttp.Server{
		Handler: router.Handler(),
	}

	rw := &readWriter{}
	ch := make(chan error)

	rw.r.WriteString("GET /GET HTTP/1.1\r\n\r\n")
	go func() {
		ch <- s.ServeConn(rw)
	}()
	select {
	case err := <-ch:
		if err != nil {
			t.Fatalf("return error %s", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("timeout")
	}
	if !get {
		t.Error("routing GET failed")
	}

	rw.r.WriteString("HEAD /GET HTTP/1.1\r\n\r\n")
	go func() {
		ch <- s.ServeConn(rw)
	}()
	select {
	case err := <-ch:
		if err != nil {
			t.Fatalf("return error %s", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("timeout")
	}
	if !head {
		t.Error("routing HEAD failed")
	}

	rw.r.WriteString("OPTIONS /GET HTTP/1.1\r\n\r\n")
	go func() {
		ch <- s.ServeConn(rw)
	}()
	select {
	case err := <-ch:
		if err != nil {
			t.Fatalf("return error %s", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("timeout")
	}
	if !options {
		t.Error("routing OPTIONS failed")
	}

	rw.r.WriteString("POST /POST HTTP/1.1\r\n\r\n")
	go func() {
		ch <- s.ServeConn(rw)
	}()
	select {
	case err := <-ch:
		if err != nil {
			t.Fatalf("return error %s", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("timeout")
	}
	if !post {
		t.Error("routing POST failed")
	}

	rw.r.WriteString("PUT /PUT HTTP/1.1\r\n\r\n")
	go func() {
		ch <- s.ServeConn(rw)
	}()
	select {
	case err := <-ch:
		if err != nil {
			t.Fatalf("return error %s", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("timeout")
	}
	if !put {
		t.Error("routing PUT failed")
	}

	rw.r.WriteString("PATCH /PATCH HTTP/1.1\r\n\r\n")
	go func() {
		ch <- s.ServeConn(rw)
	}()
	select {
	case err := <-ch:
		if err != nil {
			t.Fatalf("return error %s", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("timeout")
	}
	if !patch {
		t.Error("routing PATCH failed")
	}

	rw.r.WriteString("DELETE /DELETE HTTP/1.1\r\n\r\n")
	go func() {
		ch <- s.ServeConn(rw)
	}()
	select {
	case err := <-ch:
		if err != nil {
			t.Fatalf("return error %s", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("timeout")
	}
	if !deleted {
		t.Error("routing DELETE failed")
	}
}

func TestRouterRoot(t *testing.T) {
	t.Parallel()
	router := NewRouter()

	recv := catchPanic(func() {
		router.Get("noSlashRoot", nil)
	})
	if recv == nil {
		t.Fatal("registering path not beginning with '/' did not panic")
	}
}

func TestRouterChaining(t *testing.T) {
	router1 := NewRouter()
	router2 := NewRouter()
	router1.NotFound = router2.Handler()

	fooHit := false
	router1.Post("/foo", func(ctx *fasthttp.RequestCtx) {
		fooHit = true
		ctx.SetStatusCode(fasthttp.StatusOK)
	})

	barHit := false
	router2.Post("/bar", func(ctx *fasthttp.RequestCtx) {
		barHit = true
		ctx.SetStatusCode(fasthttp.StatusOK)
	})

	s := &fasthttp.Server{
		Handler: router1.Handler(),
	}

	rw := &readWriter{}
	ch := make(chan error)

	rw.r.WriteString("POST /foo HTTP/1.1\r\n\r\n")
	go func() {
		ch <- s.ServeConn(rw)
	}()
	select {
	case err := <-ch:
		if err != nil {
			t.Fatalf("return error %s", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("timeout")
	}
	br := bufio.NewReader(&rw.w)
	var resp fasthttp.Response
	if err := resp.Read(br); err != nil {
		t.Fatalf("Unexpected error when reading response: %s", err)
	}
	if !(resp.Header.StatusCode() == fasthttp.StatusOK && fooHit) {
		t.Errorf("Regular routing failed with router chaining.")
		t.FailNow()
	}

	rw.r.WriteString("POST /bar HTTP/1.1\r\n\r\n")
	go func() {
		ch <- s.ServeConn(rw)
	}()
	select {
	case err := <-ch:
		if err != nil {
			t.Fatalf("return error %s", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("timeout")
	}
	if err := resp.Read(br); err != nil {
		t.Fatalf("Unexpected error when reading response: %s", err)
	}
	if !(resp.Header.StatusCode() == fasthttp.StatusOK && barHit) {
		t.Errorf("Chained routing failed with router chaining.")
		t.FailNow()
	}

	rw.r.WriteString("POST /qax HTTP/1.1\r\n\r\n")
	go func() {
		ch <- s.ServeConn(rw)
	}()
	select {
	case err := <-ch:
		if err != nil {
			t.Fatalf("return error %s", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("timeout")
	}
	if err := resp.Read(br); err != nil {
		t.Fatalf("Unexpected error when reading response: %s", err)
	}
	if !(resp.Header.StatusCode() == fasthttp.StatusNotFound) {
		t.Errorf("NotFound behavior failed with router chaining.")
		t.FailNow()
	}
}

func TestRouterNotAllowed(t *testing.T) {
	handlerFunc := func(_ *fasthttp.RequestCtx) {}

	router := NewRouter()

	router.Post("/path", handlerFunc)

	// Test not allowed
	s := &fasthttp.Server{
		Handler: router.Handler(),
	}

	rw := &readWriter{}
	ch := make(chan error)

	rw.r.WriteString("GET /path HTTP/1.1\r\n\r\n")
	go func() {
		ch <- s.ServeConn(rw)
	}()
	select {
	case err := <-ch:
		if err != nil {
			t.Fatalf("return error %s", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("timeout")
	}
	br := bufio.NewReader(&rw.w)
	var resp fasthttp.Response
	if err := resp.Read(br); err != nil {
		t.Fatalf("Unexpected error when reading response: %s", err)
	}
	if !(resp.Header.StatusCode() == fasthttp.StatusMethodNotAllowed) {
		t.Errorf("NotAllowed handling failed: Code=%d", resp.Header.StatusCode())
	}

	// add another method
	router.Delete("/path", handlerFunc)
	router.Options("/path", handlerFunc) // must be ignored

	// test again
	rw.r.WriteString("GET /path HTTP/1.1\r\n\r\n")
	go func() {
		ch <- s.ServeConn(rw)
	}()
	select {
	case err := <-ch:
		if err != nil {
			t.Fatalf("return error %s", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("timeout")
	}
	if err := resp.Read(br); err != nil {
		t.Fatalf("Unexpected error when reading response: %s", err)
	}
	if !(resp.Header.StatusCode() == fasthttp.StatusMethodNotAllowed) {
		t.Errorf("NotAllowed handling failed: Code=%d", resp.Header.StatusCode())
	}
}

func TestRouterNotFound(t *testing.T) {
	handlerFunc := func(_ *fasthttp.RequestCtx) {}

	router := NewRouter()

	router.Get("/path", handlerFunc)
	router.Get("/dir/", handlerFunc)
	router.Get("/", handlerFunc)

	testRoutes := []struct {
		route string
		code  int
	}{
		{"/nope", 404}, // NotFound
	}

	s := &fasthttp.Server{
		Handler: router.Handler(),
	}

	rw := &readWriter{}
	br := bufio.NewReader(&rw.w)
	var resp fasthttp.Response
	ch := make(chan error)
	for _, tr := range testRoutes {
		rw.r.WriteString(fmt.Sprintf("GET %s HTTP/1.1\r\n\r\n", tr.route))
		go func() {
			ch <- s.ServeConn(rw)
		}()
		select {
		case err := <-ch:
			if err != nil {
				t.Fatalf("return error %s", err)
			}
		case <-time.After(100 * time.Millisecond):
			t.Fatalf("timeout")
		}
		if err := resp.Read(br); err != nil {
			t.Fatalf("Unexpected error when reading response: %s", err)
		}
		if !(resp.Header.StatusCode() == tr.code) {
			t.Errorf("NotFound handling genRoute %s failed: Code=%d want=%d",
				tr.route, resp.Header.StatusCode(), tr.code)
		}
	}

	// Test custom not found handler
	var notFound bool
	router.NotFound = func(ctx *fasthttp.RequestCtx) {
		ctx.SetStatusCode(404)
		notFound = true
	}
	rw.r.WriteString("GET /nope HTTP/1.1\r\n\r\n")
	go func() {
		ch <- s.ServeConn(rw)
	}()
	select {
	case err := <-ch:
		if err != nil {
			t.Fatalf("return error %s", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("timeout")
	}
	if err := resp.Read(br); err != nil {
		t.Fatalf("Unexpected error when reading response: %s", err)
	}
	if !(resp.Header.StatusCode() == 404 && notFound == true) {
		t.Errorf("Custom NotFound handler failed: Code=%d, Header=%v", resp.Header.StatusCode(), string(resp.Header.Peek("Location")))
	}

	// Test special case where no node for the prefix "/" exists
	router = NewRouter()
	router.Get("/a", handlerFunc)
	s.Handler = router.Handler()
	rw.r.WriteString("GET / HTTP/1.1\r\n\r\n")
	go func() {
		ch <- s.ServeConn(rw)
	}()
	select {
	case err := <-ch:
		if err != nil {
			t.Fatalf("return error %s", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("timeout")
	}
	if err := resp.Read(br); err != nil {
		t.Fatalf("Unexpected error when reading response: %s", err)
	}
	if !(resp.Header.StatusCode() == 404) {
		t.Errorf("NotFound handling genRoute / failed: Code=%d", resp.Header.StatusCode())
	}
}

func TestRouterPanicHandler(t *testing.T) {
	router := NewRouter()

	panicHandled := false

	router.Recover = func(ctx *fasthttp.RequestCtx, p interface{}) {
		panicHandled = true
	}

	router.Handle("PUT", "/user/:name", func(_ *fasthttp.RequestCtx) {
		panic("oops!")
	})

	defer func() {
		if rcv := recover(); rcv != nil {
			t.Fatal("handling panic failed")
		}
	}()

	s := &fasthttp.Server{
		Handler: router.Handler(),
	}

	rw := &readWriter{}
	ch := make(chan error)

	rw.r.WriteString(string("PUT /user/gopher HTTP/1.1\r\n\r\n"))
	go func() {
		ch <- s.ServeConn(rw)
	}()
	select {
	case err := <-ch:
		if err != nil {
			t.Fatalf("return error %s", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("timeout")
	}

	if !panicHandled {
		t.Fatal("simulating failed")
	}
}

type readWriter struct {
	net.Conn
	r bytes.Buffer
	w bytes.Buffer
}

var zeroTCPAddr = &net.TCPAddr{
	IP: net.IPv4zero,
}

func (rw *readWriter) Close() error {
	return nil
}

func (rw *readWriter) Read(b []byte) (int, error) {
	return rw.r.Read(b)
}

func (rw *readWriter) Write(b []byte) (int, error) {
	return rw.w.Write(b)
}

func (rw *readWriter) RemoteAddr() net.Addr {
	return zeroTCPAddr
}

func (rw *readWriter) LocalAddr() net.Addr {
	return zeroTCPAddr
}

func (rw *readWriter) SetReadDeadline(t time.Time) error {
	return nil
}

func (rw *readWriter) SetWriteDeadline(t time.Time) error {
	return nil
}

func catchPanic(testFunc func()) (recv interface{}) {
	defer func() {
		recv = recover()
	}()

	testFunc()
	return
}
