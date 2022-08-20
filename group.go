package forest

import (
	"fmt"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/honmaple/forest/render"
)

type (
	Group struct {
		name        string
		host        string
		prefix      string
		forest      *Forest
		parent      *Group
		children    []*Group
		middlewares []HandlerFunc

		Logger       Logger
		Renderer     render.TemplateRenderer
		ErrorHandler ErrorHandlerFunc
	}
	GroupOption func(*Group)
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

func (opt GroupOption) Forest() Option {
	return func(e *Forest) {
		opt(e.rootGroup)
	}
}

func WithName(name string) GroupOption {
	return func(g *Group) {
		prefix := ""
		if g.parent != nil && g.parent.name != "" {
			prefix = g.parent.name + "."
		}
		g.name = prefix + name
	}
}

func WithHost(host string) GroupOption {
	return func(g *Group) {
		g.host = host
	}
}

func WithPrefix(prefix string) GroupOption {
	return func(g *Group) {
		if g.parent != nil && g.parent.prefix != "" {
			g.prefix = g.parent.prefix + prefix
		} else {
			g.prefix = prefix
		}
	}
}

func WithMiddlewares(handlers ...HandlerFunc) GroupOption {
	return func(g *Group) {
		g.middlewares = combineHandlers(g.middlewares, handlers)
	}
}

func (g *Group) SetOptions(opts ...GroupOption) {
	for _, opt := range opts {
		opt(g)
	}
}

func (g *Group) Name() string {
	return g.name
}

func (g *Group) Group(opts ...GroupOption) *Group {
	n := &Group{
		parent:       g,
		host:         g.host,
		prefix:       g.prefix,
		forest:       g.forest,
		middlewares:  combineHandlers(g.middlewares, nil),
		Logger:       g.Logger,
		Renderer:     g.Renderer,
		ErrorHandler: g.ErrorHandler,
	}
	n.SetOptions(opts...)
	if g.name != "" && n.name == "" {
		n.name = fmt.Sprintf("%s.%d", g.name, len(g.children))
	}
	if g.children == nil {
		g.children = make([]*Group, 0)
	}
	g.children = append(g.children, n)
	return n
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

func (g *Group) Mount(child *Group, opts ...GroupOption) {
	if g.forest == child.forest {
		panic("forest: can't mount with same forest")
	}
	g.children = append(g.children, child)

	n := &Group{}
	n.SetOptions(opts...)

	host := child.host
	if host == "" || g.host != "" {
		host = g.host
	}
	for _, r := range child.forest.Routes() {
		route := g.Add(r.method, n.prefix+r.path, r.handlers...)
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

func NewGroup(opts ...GroupOption) *Group {
	e := New()
	e.rootGroup.SetOptions(opts...)
	return e.rootGroup
}
