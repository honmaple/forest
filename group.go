package forest

import (
	"fmt"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
)

type (
	Group struct {
		host        string
		prefix      string
		forest      *Forest
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

func (g *Group) Named(name string) *Group {
	prefix := ""
	if g.parent != nil && g.parent.Name != "" {
		prefix = g.parent.Name + "."
	}
	g.Name = prefix + name
	return g
}

func (g *Group) Host(host string, prefix string, middlewares ...HandlerFunc) *Group {
	n := &Group{
		host:         host,
		prefix:       g.prefix + prefix,
		parent:       g,
		forest:       g.forest,
		middlewares:  combineHandlers(g.middlewares, middlewares),
		Logger:       g.Logger,
		Renderer:     g.Renderer,
		ErrorHandler: g.ErrorHandler,
	}
	if g.Name != "" {
		n.Name = fmt.Sprintf("%s.%d", g.Name, len(g.children))
	}
	if g.children == nil {
		g.children = make([]*Group, 0)
	}
	g.children = append(g.children, n)
	return n
}

func (g *Group) Group(prefix string, middlewares ...HandlerFunc) *Group {
	return g.Host(g.host, prefix, middlewares...)
}

func (g *Group) Use(middlewares ...HandlerFunc) *Group {
	g.middlewares = append(g.middlewares, middlewares...)
	return g
}

func (g *Group) Add(method string, path string, handlers ...HandlerFunc) *Route {
	if len(handlers) == 0 {
		panic("no handler found: " + path)
	}
	if matches, err := regexp.MatchString("^[A-Z]+$", method); !matches || err != nil {
		panic("http method " + method + " is not valid")
	}
	route := g.forest.addRoute(g.host, method, g.prefix+path)
	route.group = g
	route.Name = handlerName(handlers[len(handlers)-1])
	route.handlers = combineHandlers(g.middlewares, handlers)
	return route
}

func (g *Group) TRACE(path string, handlers ...HandlerFunc) *Route {
	return g.Add(http.MethodTrace, path, handlers...)
}

func (g *Group) CONNECT(path string, handlers ...HandlerFunc) *Route {
	return g.Add(http.MethodConnect, path, handlers...)
}

func (g *Group) OPTIONS(path string, handlers ...HandlerFunc) *Route {
	return g.Add(http.MethodOptions, path, handlers...)
}

func (g *Group) HEAD(path string, handlers ...HandlerFunc) *Route {
	return g.Add(http.MethodHead, path, handlers...)
}

func (g *Group) GET(path string, handlers ...HandlerFunc) *Route {
	return g.Add(http.MethodGet, path, handlers...)
}

func (g *Group) POST(path string, handlers ...HandlerFunc) *Route {
	return g.Add(http.MethodPost, path, handlers...)
}

func (g *Group) PUT(path string, handlers ...HandlerFunc) *Route {
	return g.Add(http.MethodPut, path, handlers...)
}

func (g *Group) PATCH(path string, handlers ...HandlerFunc) *Route {
	return g.Add(http.MethodPatch, path, handlers...)
}

func (g *Group) DELETE(path string, handlers ...HandlerFunc) *Route {
	return g.Add(http.MethodDelete, path, handlers...)
}

func (g *Group) Any(path string, handlers ...HandlerFunc) Routes {
	routes := make(Routes, len(methods))
	for i, m := range methods {
		routes[i] = g.Add(m, path, handlers...)
	}
	return routes
}

func (g *Group) Static(path, root string, middlewares ...HandlerFunc) *Route {
	return g.StaticFS(path, http.Dir(root), middlewares...)
}

func (g *Group) StaticFS(path string, fs http.FileSystem, middlewares ...HandlerFunc) *Route {
	if strings.Contains(path, ":") || strings.Contains(path, "*") || strings.Contains(path, "{") {
		panic("URL parameters can not be used when serving a static file")
	}
	handler := func(c Context) error {
		return c.FileFromFS(c.Param("*"), fs)
	}
	path = filepath.Join(path, "/*")
	if len(middlewares) == 0 {
		return g.GET(path, handler)
	}
	return g.GET(path, append(middlewares, handler)...)
}

func (g *Group) StaticFile(path, file string, middlewares ...HandlerFunc) *Route {
	if strings.Contains(path, ":") || strings.Contains(path, "*") || strings.Contains(path, "{") {
		panic("URL parameters can not be used when serving a static file")
	}
	handler := func(c Context) error {
		return c.File(file)
	}
	if len(middlewares) == 0 {
		return g.GET(path, handler)
	}
	return g.GET(path, append(middlewares, handler)...)
}

func (g *Group) Mount(prefix string, child *Group) {
	if g.forest == child.forest {
		panic("forest: can't mount with same forest")
	}
	g.children = append(g.children, child)

	host := child.host
	if host == "" || g.host != "" {
		host = g.host
	}
	for _, r := range child.forest.Routes() {
		route := g.Add(r.method, prefix+r.path, r.handlers...)
		route.host = host
		route.group = r.group
	}
}

func (g *Group) newRoute(method, path string, handlers []HandlerFunc) *Route {
	return &Route{
		group:    g,
		host:     g.host,
		path:     g.prefix + path,
		method:   method,
		handlers: combineHandlers(g.middlewares, handlers),
	}
}

func NewHost(host string, opts ...Option) *Group {
	e := New(opts...)
	e.host = host
	return e.rootGroup
}

func NewGroup(opts ...Option) *Group {
	e := New(opts...)
	return e.rootGroup
}
