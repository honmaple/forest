package forest

import (
	"net/http"
)

type (
	Group struct {
		host        string
		prefix      string
		engine      *Engine
		parent      *Group
		children    []*Group
		middlewares []HandlerFunc

		Name         string
		Logger       Logger
		Renderer     Renderer
		ErrorHandler ErrorHandlerFunc
	}
	Renderer interface {
		Render(http.ResponseWriter, string, interface{}) error
	}
)

var methods = [...]string{
	http.MethodConnect,
	http.MethodDelete,
	http.MethodGet,
	http.MethodHead,
	http.MethodOptions,
	http.MethodPatch,
	http.MethodPost,
	http.MethodPut,
	http.MethodTrace,
}

func NewContext(e *Engine, r *http.Request, w http.ResponseWriter) Context {
	c := &context{engine: e, response: NewResponse(w)}
	c.reset(r, w)
	return c
}

func (g *Group) Host(host string, prefix string, middlewares ...HandlerFunc) *Group {
	n := &Group{
		host:         host,
		prefix:       g.prefix + prefix,
		parent:       g,
		engine:       g.engine,
		middlewares:  mergeHandlers(g.middlewares, middlewares),
		Logger:       g.Logger,
		Renderer:     g.Renderer,
		ErrorHandler: g.ErrorHandler,
	}
	if g.children == nil {
		g.children = make([]*Group, 0)
	}
	g.children = append(g.children, n)
	return n
}

func (g *Group) Group(prefix string, middlewares ...HandlerFunc) *Group {
	return g.Host("", prefix, middlewares...)
}

func (g *Group) Use(middlewares ...HandlerFunc) {
	g.middlewares = append(g.middlewares, middlewares...)
}

func (g *Group) Add(method string, path string, handler HandlerFunc, middlewares ...HandlerFunc) *Route {
	route := &Route{
		Host:   g.host,
		Path:   g.prefix + path,
		Method: method,
		group:  g,
	}
	route.Name = handlerName(handler)
	route.Handlers = append(mergeHandlers(g.middlewares, middlewares), handler)

	g.engine.addRoute(route)
	return route
}

func (g *Group) OPTIONS(path string, handler HandlerFunc, middlewares ...HandlerFunc) *Route {
	return g.Add(http.MethodOptions, path, handler, middlewares...)
}

func (g *Group) HEAD(path string, handler HandlerFunc, middlewares ...HandlerFunc) *Route {
	return g.Add(http.MethodHead, path, handler, middlewares...)
}

func (g *Group) GET(path string, handler HandlerFunc, middlewares ...HandlerFunc) *Route {
	return g.Add(http.MethodGet, path, handler, middlewares...)
}

func (g *Group) POST(path string, handler HandlerFunc, middlewares ...HandlerFunc) *Route {
	return g.Add(http.MethodPost, path, handler, middlewares...)
}

func (g *Group) PUT(path string, handler HandlerFunc, middlewares ...HandlerFunc) *Route {
	return g.Add(http.MethodPost, path, handler, middlewares...)
}

func (g *Group) PATCH(path string, handler HandlerFunc, middlewares ...HandlerFunc) *Route {
	return g.Add(http.MethodPatch, path, handler, middlewares...)
}

func (g *Group) DELETE(path string, handler HandlerFunc, middlewares ...HandlerFunc) *Route {
	return g.Add(http.MethodDelete, path, handler, middlewares...)
}

func (g *Group) Any(path string, handler HandlerFunc, middlewares ...HandlerFunc) []*Route {
	routes := make([]*Route, len(methods))
	for i, m := range methods {
		routes[i] = g.Add(m, path, handler, middlewares...)
	}
	return routes
}

func (g *Group) Mount(prefix string, group *Group) {
	if g.engine == group.engine {
		return
	}
	for _, r := range group.engine.Routes() {
		r.Path = prefix + r.Path
		r.Handlers = append(g.middlewares, r.Handlers...)
		g.engine.addRoute(r)
	}
	group.parent = g
	group.prefix = prefix + g.prefix
	group.engine = g.engine
	if group.Logger == nil {
		group.Logger = g.Logger
	}
	if group.Renderer == nil {
		group.Renderer = g.Renderer
	}
	if group.ErrorHandler == nil {
		group.ErrorHandler = g.ErrorHandler

	}
}

func NewHost(host string) *Group {
	g := &Group{
		host:        host,
		engine:      New(),
		middlewares: make([]HandlerFunc, 0),
	}
	g.engine.Debug = false
	return g
}

func NewGroup() *Group {
	g := &Group{
		engine:      New(),
		middlewares: make([]HandlerFunc, 0),
	}
	g.engine.Debug = false
	return g
}
