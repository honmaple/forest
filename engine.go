package forest

import (
	stdcontext "context"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"sync"
)

type (
	rootGroup = Group
	Engine    struct {
		*rootGroup
		mu                    sync.Mutex
		contextPool           sync.Pool
		router                *Router
		notFoundRoute         *Route
		methodNotAllowedRoute *Route
		debug                 bool
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

func mergeHandlers(m1 []HandlerFunc, m2 []HandlerFunc) []HandlerFunc {
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

type Option func(e *Engine)

func Debug() Option {
	return func(e *Engine) {
		e.debug = true
	}
}

func New(opts ...Option) *Engine {
	e := &Engine{
		router: newRouter(),
	}
	e.rootGroup = &Group{
		engine:      e,
		middlewares: make([]HandlerFunc, 0),
	}
	e.contextPool = sync.Pool{
		New: func() interface{} {
			return NewContext(nil, nil)
		},
	}
	e.Logger = newLogger()
	e.ErrorHandler = ErrorHandler
	e.NotFound(NotFoundHandler)
	e.MethodNotAllowed(MethodNotAllowedHandler)
	for _, opt := range opts {
		opt(e)
	}
	return e
}

func WrapHandler(h http.Handler) HandlerFunc {
	return func(c Context) error {
		h.ServeHTTP(c.Response(), c.Request())
		return nil
	}
}

func (e *Engine) addRoute(route *Route) {
	e.router.Insert(route)
}

func (e *Engine) rebuild(route *Route) {
	rlen := len(route.Handlers)
	if rlen > 1 {
		rlen = 1
	}
	handlers := make([]HandlerFunc, len(e.middlewares)+rlen)
	copy(handlers, e.middlewares)

	if rlen > 0 {
		handlers[len(e.middlewares)] = route.Last()
	}
	route.Handlers = handlers
}

func (e *Engine) URL(name string, args ...interface{}) string {
	if r := e.Route(name); r != nil {
		return r.URL(args...)
	}
	return ""
}

func (e *Engine) Route(name string) *Route {
	for _, r := range e.router.routes {
		if r.Name == name {
			return r
		}
	}
	return nil
}

func (e *Engine) Router() *Router {
	return e.router
}

func (e *Engine) Routes() []*Route {
	routes := make([]*Route, 0, len(e.router.routes))
	for _, r := range e.router.routes {
		routes = append(routes, r)
	}
	return routes
}

func (e *Engine) Use(middlewares ...HandlerFunc) *Engine {
	e.rootGroup.Use(middlewares...)
	e.rebuild(e.notFoundRoute)
	e.rebuild(e.methodNotAllowedRoute)
	return e
}

func (e *Engine) Mount(prefix string, child *Engine) {
	e.rootGroup.Mount(prefix, child.rootGroup)
}

func (e *Engine) MountGroup(prefix string, child *Group) {
	e.rootGroup.Mount(prefix, child)
}

func (e *Engine) NotFound(h HandlerFunc) *Route {
	if e.notFoundRoute == nil {
		e.notFoundRoute = &Route{Handlers: make([]HandlerFunc, 1), group: e.rootGroup}
	}
	e.notFoundRoute.Name = handlerName(h)
	e.notFoundRoute.Handlers[len(e.notFoundRoute.Handlers)-1] = h
	return e.notFoundRoute
}

func (e *Engine) MethodNotAllowed(h HandlerFunc) *Route {
	if e.methodNotAllowedRoute == nil {
		e.methodNotAllowedRoute = &Route{Handlers: make([]HandlerFunc, 1), group: e.rootGroup}
	}
	e.methodNotAllowedRoute.Name = handlerName(h)
	e.methodNotAllowedRoute.Handlers[len(e.methodNotAllowedRoute.Handlers)-1] = h
	return e.methodNotAllowedRoute
}

func (e *Engine) Context(w http.ResponseWriter, r *http.Request) *context {
	c := e.contextPool.Get().(*context)
	c.reset(r, w)
	defer e.contextPool.Put(c)

	path := r.URL.RawPath
	if path == "" {
		path = r.URL.Path
	}

	route, found := e.router.Find(r.Host, r.Method, path, c)
	if found && route != nil {
		c.route = route
	} else if found {
		c.route = e.methodNotAllowedRoute
	} else {
		c.route = e.notFoundRoute
	}
	return c
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e.Context(w, r).Next()
}

func (e *Engine) configure(addr string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.debug {
		for _, r := range e.router.routes {
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

func (e *Engine) Start(addr string) error {
	e.configure(addr)
	return e.Server.ListenAndServe()
}

func (e *Engine) StartTLS(addr string, certFile, keyFile string) error {
	e.configure(addr)
	return e.Server.ListenAndServeTLS(certFile, keyFile)
}

func (e *Engine) Shutdown(ctx stdcontext.Context) error {
	return e.Server.Shutdown(ctx)
}
