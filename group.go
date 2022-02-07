package forest

import (
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
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
	return g.Host(g.host, prefix, middlewares...)
}

func (g *Group) Use(middlewares ...HandlerFunc) *Group {
	g.middlewares = append(g.middlewares, middlewares...)
	return g
}

func (g *Group) Add(method string, path string, handler HandlerFunc, middlewares ...HandlerFunc) *Route {
	if matches, err := regexp.MatchString("^[A-Z]+$", method); !matches || err != nil {
		panic("http method " + method + " is not valid")
	}
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

func (g *Group) TRACE(path string, handler HandlerFunc, middlewares ...HandlerFunc) *Route {
	return g.Add(http.MethodTrace, path, handler, middlewares...)
}

func (g *Group) CONNECT(path string, handler HandlerFunc, middlewares ...HandlerFunc) *Route {
	return g.Add(http.MethodConnect, path, handler, middlewares...)
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

func (g *Group) Any(path string, handler HandlerFunc, middlewares ...HandlerFunc) Routes {
	routes := make(Routes, len(methods))
	for i, m := range methods {
		routes[i] = g.Add(m, path, handler, middlewares...)
	}
	return routes
}

func (g *Group) StaticFile(path, file string, middlewares ...HandlerFunc) *Route {
	if strings.Contains(path, ":") || strings.Contains(path, "*") || strings.Contains(path, "{") {
		panic("URL parameters can not be used when serving a static file")
	}
	handler := func(c Context) error {
		return c.File(file)
	}
	return g.GET(path, handler, middlewares...)
}

func (g *Group) Static(path, root string, middlewares ...HandlerFunc) *Route {
	return g.StaticFS(path, http.Dir(root), middlewares...)
}

func (g *Group) StaticFS(path string, fs http.FileSystem, middlewares ...HandlerFunc) *Route {
	if strings.Contains(path, ":") || strings.Contains(path, "*") || strings.Contains(path, "{") {
		panic("URL parameters can not be used when serving a static file")
	}
	const indexPage = "/index.html"
	handler := func(c Context) error {
		fpath := c.Param("*")
		fname := filepath.Clean("/" + fpath)

		file, err := fs.Open(fname)
		if err != nil {
			return NotFoundHandler(c)
		}
		defer file.Close()

		fi, err := file.Stat()
		if err != nil {
			return NotFoundHandler(c)
		}
		url := c.Request().URL.Path
		if fi.IsDir() {
			if url == "" || url[len(url)-1] != '/' {
				return c.Redirect(http.StatusMovedPermanently, url+"/")
			}
			index, err := fs.Open(strings.TrimSuffix(fname, "/") + indexPage)
			if err != nil {
				return NotFoundHandler(c)
			}
			defer index.Close()

			indexfi, err := file.Stat()
			if err != nil {
				return NotFoundHandler(c)
			}
			file, fi = index, indexfi
		}
		http.ServeContent(c.Response(), c.Request(), fi.Name(), fi.ModTime(), file)
		return nil
	}
	path = filepath.Join(path, "/*")
	return g.GET(path, handler, middlewares...)
}

func (g *Group) Mount(prefix string, child *Group) {
	if g.engine == child.engine {
		panic("forest: can't mount with same engine")
	}
	if child.host == "" {
		child.host = g.host
	}
	child.parent = g
	child.prefix = g.prefix + prefix + child.prefix
	child.middlewares = mergeHandlers(g.middlewares, child.middlewares)
	g.children = append(g.children, child)
	g.engine.Mount(prefix, child.engine)
	child.engine = g.engine
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
