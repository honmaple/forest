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
	Engine struct {
		Router
		pool            sync.Pool
		routers         map[string]*router
		Debug           bool
		NotFoundHandler HandlerFunc
	}
	HandlerFunc      func(Context) error
	ErrorHandlerFunc func(error, Context)
)

var (
	ErrorHandler = func(err error, c Context) {
		c.Logger().Errorln(c.Request().URL.Path, err.Error())
	}
	NotFoundHandler = func(c Context) error {
		return c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Request().URL.Path)
	}
)

func New() *Engine {
	e := &Engine{
		routers: make(map[string]*router),
		Debug:   true,
	}
	e.Router = &Group{
		engine:       e,
		middlewares:  make([]HandlerFunc, 0),
		Logger:       newLogger(),
		ErrorHandler: ErrorHandler,
	}
	e.NotFoundHandler = NotFoundHandler
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
	route := router.Find(r.Method, r.URL.Path, c.params)
	if route != nil {
		c.route = route
		if err := c.Next(); err != nil {
			route.ErrorHandler(err, c)
		}
	} else {
		e.NotFoundHandler(c)
	}
}

func (e *Engine) Run(addr string) error {
	server := http.Server{Addr: addr, Handler: e}
	return server.ListenAndServe()
}

func handlerName(h HandlerFunc) string {
	t := reflect.ValueOf(h).Type()
	if t.Kind() == reflect.Func {
		return runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
	}
	return t.String()
}

func mergeMiddlewares(m1 []HandlerFunc, m2 []HandlerFunc) []HandlerFunc {
	m := make([]HandlerFunc, 0, len(m1)+len(m2))
	m = append(m, m1...)
	m = append(m, m2...)
	return m
}

func debugPrint(r *Route) {
	fmt.Fprintf(os.Stdout, "[DEBUG] %-6s %s%-30s --> %s (%d handlers)\n", r.Method, r.Host, r.Path, r.Name, len(r.Middlewares)+1)
}
