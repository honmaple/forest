package forest

import (
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
		pool                  sync.Pool
		router                *Router
		notFoundRoute         *Route
		methodNotAllowedRoute *Route
		Debug                 bool
	}
	Error struct {
		Code    int         `json:"-"`
		Message interface{} `json:"message"`
	}
	HandlerFunc      func(Context) error
	ErrorHandlerFunc func(error, Context)
)

var (
	ErrNotFound            = NewError(http.StatusNotFound)
	ErrMethodNotAllowed    = NewError(http.StatusMethodNotAllowed)
	ErrInternalServerError = NewError(http.StatusInternalServerError)

	NotFoundHandler = func(c Context) error {
		return ErrNotFound
	}
	MethodNotAllowedHandler = func(c Context) error {
		return ErrMethodNotAllowed
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

func New() *Engine {
	e := &Engine{
		router: newRouter(),
	}
	e.rootGroup = &Group{
		engine:      e,
		middlewares: make([]HandlerFunc, 0),
	}
	e.Debug = true
	e.Logger = newLogger()
	e.ErrorHandler = ErrorHandler
	e.pool.New = func() interface{} {
		return NewContext(e, nil, nil)
	}
	e.NotFound(NotFoundHandler)
	e.MethodNotAllowed(MethodNotAllowedHandler)
	return e
}

func (e *Engine) addRoute(route *Route) {
	if e.Debug {
		debugPrint(route.String())
	}

	e.router.Insert(route)
	return
}

func (e *Engine) Routes() []*Route {
	routes := make([]*Route, 0)
	for _, r := range e.router.routes {
		routes = append(routes, r)
	}
	return routes
}

func (e *Engine) URL(name string, args ...interface{}) string {
	for _, r := range e.router.routes {
		if r.Name == name {
			return r.URL(args...)
		}
	}
	return ""
}

func (e *Engine) Use(middlewares ...HandlerFunc) {
	e.rootGroup.Use(middlewares...)
	e.notFoundRoute.Handlers = append(e.middlewares, e.notFoundRoute.Handlers[len(e.notFoundRoute.Handlers)-1])
	e.methodNotAllowedRoute.Handlers = append(e.middlewares, e.methodNotAllowedRoute.Handlers[len(e.methodNotAllowedRoute.Handlers)-1])
}

func (e *Engine) NotFound(h HandlerFunc) {
	if e.notFoundRoute == nil {
		e.notFoundRoute = &Route{Handlers: make([]HandlerFunc, 1), group: e.rootGroup}
	}
	e.notFoundRoute.Handlers[len(e.notFoundRoute.Handlers)-1] = h
}

func (e *Engine) MethodNotAllowed(h HandlerFunc) {
	if e.methodNotAllowedRoute == nil {
		e.methodNotAllowedRoute = &Route{Handlers: make([]HandlerFunc, 1), group: e.rootGroup}
	}
	e.methodNotAllowedRoute.Handlers[len(e.methodNotAllowedRoute.Handlers)-1] = h
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := e.pool.Get().(*context)
	c.reset(r, w)
	defer e.pool.Put(c)

	path := r.URL.EscapedPath()
	if path == "" {
		path = "/"
	}

	route, found := e.router.Find(r.Host, r.Method, path, c)
	if found && route != nil {
		c.route = route
	} else if found {
		c.route = e.methodNotAllowedRoute
	} else {
		c.route = e.notFoundRoute
	}
	c.Next()
}

func (e *Engine) Run(addr string) error {
	server := http.Server{Addr: addr, Handler: e}
	return server.ListenAndServe()
}

func (e *Error) Error() string {
	return fmt.Sprintf("code=%d, message=%v", e.Code, e.Message)
}

func NewError(code int, message ...interface{}) *Error {
	e := &Error{
		Code: code,
	}
	if len(message) > 0 {
		e.Message = message[0]
	} else {
		e.Message = http.StatusText(code)
	}
	return e
}

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

func debugPrint(msg string) {
	fmt.Fprint(os.Stdout, msg)
}

func sprintf(format string, args ...interface{}) string {
	if len(args) == 0 {
		return format
	}
	return fmt.Sprintf(format, args)
}
