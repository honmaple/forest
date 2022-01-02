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
		pool                    sync.Pool
		routers                 map[string]*router
		Debug                   bool
		NotFoundHandler         HandlerFunc
		MethodNotAllowedHandler HandlerFunc
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
		c.JSON(e.Code, H{"message": e.Message})
	}
)

func New() *Engine {
	e := &Engine{
		routers: make(map[string]*router),
		Debug:   true,
	}
	e.rootGroup = &Group{
		engine:      e,
		middlewares: make([]HandlerFunc, 0),
	}
	e.Logger = newLogger()
	e.ErrorHandler = ErrorHandler
	e.NotFoundHandler = NotFoundHandler
	e.MethodNotAllowedHandler = MethodNotAllowedHandler
	e.pool.New = func() interface{} {
		return NewContext(e, nil, nil)
	}
	return e
}

func (e *Engine) findRouter(host string) *router {
	if r, ok := e.routers[host]; ok {
		return r
	}
	return e.routers[""]
}

func (e *Engine) addRoute(route *Route) {
	if e.Debug {
		debugPrint(route)
	}

	r, ok := e.routers[route.Host]
	if !ok {
		r = newrouter()
		e.routers[route.Host] = r
	}
	r.Add(route)
	return
}

func (e *Engine) Routes() []*Route {
	routes := make([]*Route, 0)
	for _, router := range e.routers {
		for _, r := range router.routes {
			routes = append(routes, r)
		}
	}
	return routes
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := e.pool.Get().(*context)
	c.reset(r, w)
	defer e.pool.Put(c)

	router := e.findRouter(r.Host)
	route, found := router.Find(r.Method, r.URL.EscapedPath(), c.params)
	if found && route != nil {
		c.route = route
	} else if found {
		c.route = &Route{Handler: e.MethodNotAllowedHandler, Middlewares: e.middlewares, group: e.rootGroup}
	} else {
		c.route = &Route{Handler: e.NotFoundHandler, Middlewares: e.middlewares, group: e.rootGroup}
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

func debugPrint(r *Route) {
	fmt.Fprintf(os.Stdout, "[DEBUG] %-6s %s%-30s --> %s (%d handlers)\n", r.Method, r.Host, r.Path, r.Name, len(r.Middlewares)+1)
}
