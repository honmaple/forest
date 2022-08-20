package forest

import (
	stdcontext "context"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"sync"
)

type (
	rootGroup = Group
	Forest    struct {
		*rootGroup
		mu                    sync.Mutex
		contextPool           sync.Pool
		node                  *node
		nodes                 map[string]*node
		routes                map[string]*Route
		notFound              []HandlerFunc
		methodNotAllowed      []HandlerFunc
		notFoundRoute         *Route
		methodNotAllowedRoute *Route
		maxParam              int
		debug                 bool
		hostMatch             func(string, string) bool
		Server                *http.Server
	}
	HandlerFunc      func(Context) error
	ErrorHandlerFunc func(error, Context)
)

var (
	ErrNotFound            = NewError(http.StatusNotFound)
	ErrMethodNotAllowed    = NewError(http.StatusMethodNotAllowed)
	ErrInternalServerError = NewError(http.StatusInternalServerError)

	NotFoundMessage         = []byte(ErrNotFound.Error())
	MethodNotAllowedMessage = []byte(ErrMethodNotAllowed.Error())

	NotFoundHandler = func(c Context) error {
		return c.Bytes(http.StatusNotFound, NotFoundMessage)
	}
	MethodNotAllowedHandler = func(c Context) error {
		return c.Bytes(http.StatusMethodNotAllowed, MethodNotAllowedMessage)
	}
	ErrorHandler = func(err error, c Context) {
		if err == nil {
			return
		}
		e, ok := err.(*Error)
		if !ok {
			e = ErrInternalServerError
		}
		if resp := c.Response(); !resp.Written() {
			c.String(e.Code, e.Error())
		}
	}
)

func handlerName(h HandlerFunc) string {
	t := reflect.ValueOf(h).Type()
	if t.Kind() == reflect.Func {
		return runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
	}
	return t.String()
}

func combineHandlers(m1 []HandlerFunc, m2 []HandlerFunc) []HandlerFunc {
	m := make([]HandlerFunc, 0, len(m1)+len(m2))
	m = append(m, m1...)
	m = append(m, m2...)
	return m
}

func sprintf(format string, args ...interface{}) string {
	if len(args) == 0 {
		return format
	}
	return fmt.Sprintf(format, args...)
}

func debugPrint(msg string, args ...interface{}) {
	fmt.Fprint(os.Stdout, sprintf(msg, args...))
}

type Option func(*Forest)

func (opt Option) Group() GroupOption {
	return func(g *Group) {
		opt(g.forest)
	}
}

func Debug() Option {
	return func(e *Forest) {
		e.debug = true
	}
}

func HostMatch(matcher func(string, string) bool) Option {
	return func(e *Forest) {
		e.hostMatch = matcher
	}
}

func Middlewares(handlers ...HandlerFunc) Option {
	return func(e *Forest) {
		e.middlewares = handlers
	}
}

func New(opts ...Option) *Forest {
	e := &Forest{
		node:   &node{},
		routes: make(map[string]*Route),
	}
	e.rootGroup = &Group{
		forest:      e,
		middlewares: make([]HandlerFunc, 0),
	}
	e.contextPool = sync.Pool{
		New: func() interface{} {
			return e.NewContext(nil, nil)
		},
	}
	e.Logger = newLogger()
	e.ErrorHandler = ErrorHandler
	e.NotFound(NotFoundHandler)
	e.MethodNotAllowed(MethodNotAllowedHandler)
	e.SetOptions(opts...)
	return e
}

func HostMatcher(host, dst string) bool {
	c := len(host) - len(dst)
	if c < 0 {
		return false
	}
	index := strings.IndexRune(dst, '*')
	if index < 0 {
		return false
	}
	return dst[:index] == host[:index] && dst[index+1:] == host[c+index+1:]
}

func WrapHandler(h http.Handler) HandlerFunc {
	return func(c Context) error {
		h.ServeHTTP(c.Response(), c.Request())
		return nil
	}
}

func WrapHandlerFunc(h http.HandlerFunc) HandlerFunc {
	return func(c Context) error {
		h(c.Response(), c.Request())
		return nil
	}
}

func (e *Forest) addRoute(host, method, path string) *Route {
	key := host + method + path
	if route, ok := e.routes[key]; ok {
		return route
	}

	root := e.node
	if host != "" {
		if e.nodes == nil {
			e.nodes = make(map[string]*node, 0)
		}
		h, ok := e.nodes[host]
		if !ok {
			h = &node{}
			e.nodes[host] = h
		}
		root = h
	}

	route := &Route{
		host:   host,
		method: method,
		path:   path,
	}
	root.insert(route.Path(), route)

	// maybe should check route.pnames can't be repeated
	if l := len(route.pnames); l > e.maxParam {
		e.maxParam = l
	}

	e.routes[key] = route
	return route
}

func (e *Forest) findHost(host string) *node {
	if host != "" && len(e.nodes) > 0 {
		if node, ok := e.nodes[host]; ok {
			return node
		}
		if e.hostMatch == nil {
			return e.node
		}
		for h, node := range e.nodes {
			if e.hostMatch(host, h) {
				return node
			}
		}
	}
	return e.node
}

func (e *Forest) findRoute(host, method, path string, params *contextParams) *Route {
	n := e.findHost(host).find(path, params)
	if n == nil || n.routes == nil || len(n.routes) == 0 {
		return e.notFoundRoute
	}
	for _, route := range n.routes {
		if route.Method() == method {
			return route
		}
	}
	return e.methodNotAllowedRoute
}

func (e *Forest) Route(name string) *Route {
	for _, route := range e.routes {
		if route.Name == name {
			return route
		}
	}
	return nil
}

func (e *Forest) Routes() []*Route {
	routes := make([]*Route, 0, len(e.routes))
	for _, r := range e.routes {
		routes = append(routes, r)
	}
	sort.Slice(routes, func(i, j int) bool {
		return routes[i].Path() < routes[j].Path()
	})
	return routes
}

func (e *Forest) URL(name string, args ...interface{}) string {
	if r := e.Route(name); r != nil {
		return r.URL(args...)
	}
	return ""
}

func (e *Forest) NotFound(handlers ...HandlerFunc) *Route {
	if e.notFoundRoute == nil {
		e.notFoundRoute = &Route{}
	}
	e.notFound = handlers
	e.notFoundRoute.handlers = combineHandlers(e.middlewares, e.notFound)
	return e.notFoundRoute
}

func (e *Forest) MethodNotAllowed(handlers ...HandlerFunc) *Route {
	if e.methodNotAllowedRoute == nil {
		e.methodNotAllowedRoute = &Route{}
	}
	e.methodNotAllowed = handlers
	e.methodNotAllowedRoute.handlers = combineHandlers(e.middlewares, e.methodNotAllowed)
	return e.methodNotAllowedRoute
}

func (e *Forest) Use(middlewares ...HandlerFunc) *Forest {
	e.rootGroup.Use(middlewares...)
	e.notFoundRoute.handlers = combineHandlers(e.middlewares, e.notFound)
	e.methodNotAllowedRoute.handlers = combineHandlers(e.middlewares, e.methodNotAllowed)
	return e
}

func (e *Forest) Mount(child *Forest, opts ...GroupOption) {
	e.rootGroup.Mount(child.rootGroup, opts...)
}

func (e *Forest) MountGroup(child *Group, opts ...GroupOption) {
	e.rootGroup.Mount(child, opts...)
}

func (e *Forest) NewContext(w http.ResponseWriter, r *http.Request) *context {
	c := &context{
		params: &contextParams{
			pvalues: make([]string, e.maxParam),
		},
		response: NewResponse(w),
	}
	return c
}

func (e *Forest) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := e.contextPool.Get().(*context)
	c.reset(r, w)
	defer e.contextPool.Put(c)

	path := r.URL.RawPath
	if path == "" {
		path = r.URL.Path
	}
	// pass []string is faster than *context than *([]string)
	c.route = e.findRoute(r.Host, r.Method, path, c.params)
	c.Next()
}

func (e *Forest) configure(addr string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.debug {
		for _, r := range e.Routes() {
			debugPrint(r.String())
		}
		debugPrint("Listening and serving HTTP on %s\n", addr)
	}
	if e.Server == nil {
		e.Server = &http.Server{Handler: e}
	}
	e.Server.Addr = addr
	return nil
}

func (e *Forest) SetOptions(opts ...Option) {
	for _, opt := range opts {
		opt(e)
	}
}

func (e *Forest) Start(addr string) error {
	e.configure(addr)
	return e.Server.ListenAndServe()
}

func (e *Forest) StartTLS(addr string, certFile, keyFile string) error {
	e.configure(addr)
	return e.Server.ListenAndServeTLS(certFile, keyFile)
}

func (e *Forest) Shutdown(ctx stdcontext.Context) error {
	return e.Server.Shutdown(ctx)
}
