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
	Forest    struct {
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

type Option func(e *Forest)

func Debug() Option {
	return func(e *Forest) {
		e.debug = true
	}
}

func New(opts ...Option) *Forest {
	e := &Forest{
		router: NewRouter(),
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

func WrapHandlerFunc(h http.HandlerFunc) HandlerFunc {
	return func(c Context) error {
		h(c.Response(), c.Request())
		return nil
	}
}

func (e *Forest) addRoute(host, method, path string) *Route {
	return e.router.Insert(host, method, path)
}

func (e *Forest) rebuild(route *Route) {
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

func (e *Forest) URL(name string, args ...interface{}) string {
	if r := e.Route(name); r != nil {
		return r.URL(args...)
	}
	return ""
}

func (e *Forest) Route(name string) *Route {
	for _, r := range e.router.routes {
		if r.Name == name {
			return r
		}
	}
	return nil
}

func (e *Forest) Router() *Router {
	return e.router
}

func (e *Forest) Routes() []*Route {
	routes := make([]*Route, 0, len(e.router.routes))
	for _, r := range e.router.routes {
		routes = append(routes, r)
	}
	return routes
}

func (e *Forest) Use(middlewares ...HandlerFunc) *Forest {
	e.rootGroup.Use(middlewares...)
	e.rebuild(e.notFoundRoute)
	e.rebuild(e.methodNotAllowedRoute)
	return e
}

func (e *Forest) Mount(prefix string, child *Forest) {
	e.rootGroup.Mount(prefix, child.rootGroup)
}

func (e *Forest) MountGroup(prefix string, child *Group) {
	e.rootGroup.Mount(prefix, child)
}

func (e *Forest) NotFound(h HandlerFunc) *Route {
	if e.notFoundRoute == nil {
		e.notFoundRoute = &Route{Handlers: make([]HandlerFunc, 1), group: e.rootGroup}
	}
	e.notFoundRoute.Name = handlerName(h)
	e.notFoundRoute.Handlers[len(e.notFoundRoute.Handlers)-1] = h
	return e.notFoundRoute
}

func (e *Forest) MethodNotAllowed(h HandlerFunc) *Route {
	if e.methodNotAllowedRoute == nil {
		e.methodNotAllowedRoute = &Route{Handlers: make([]HandlerFunc, 1), group: e.rootGroup}
	}
	e.methodNotAllowedRoute.Name = handlerName(h)
	e.methodNotAllowedRoute.Handlers[len(e.methodNotAllowedRoute.Handlers)-1] = h
	return e.methodNotAllowedRoute
}

func (e *Forest) NewContext(w http.ResponseWriter, r *http.Request) *context {
	c := &context{
		pvalues:  make([]string, e.router.maxParam),
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
	route, found := e.router.Find(r.Host, r.Method, path, c.pvalues)
	if found && route != nil {
		c.route = route
	} else if found {
		c.route = e.methodNotAllowedRoute
	} else {
		c.route = e.notFoundRoute
	}
	c.Next()
}

func (e *Forest) configure(addr string) error {
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
