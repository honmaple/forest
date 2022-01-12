package forest

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
)

type (
	routeParam struct {
		key   string
		start int
		end   int
	}
	Route struct {
		pnames   []routeParam
		group    *Group
		Name     string        `json:"name"`
		Host     string        `json:"host"`
		Path     string        `json:"path"`
		Method   string        `json:"method"`
		Handlers []HandlerFunc `json:"-"`
	}
	Routes []*Route
	Router struct {
		hosts  map[string]*node
		routes Routes
	}
)

func (rs Routes) find(method string) *Route {
	for _, r := range rs {
		if r.Method == method {
			return r
		}
	}
	return nil
}

func (r *Router) Insert(route *Route) {
	root, ok := r.hosts[route.Host]
	if !ok {
		root = &node{}
		r.hosts[route.Host] = root
	}
	root.insert(route)
	r.routes = append(r.routes, route)
}

func (r *Router) Find(host, method, path string, c *context) (*Route, bool) {
	if root, ok := r.hosts[host]; ok {
		return root.search(method, path, c)
	}
	return r.hosts[""].search(method, path, c)
}

func newRouter() *Router {
	return &Router{hosts: make(map[string]*node)}
}

func (r *Route) Logger() Logger {
	group := r.group
	for group != nil {
		if group.Logger != nil {
			return group.Logger
		}
		group = group.parent
	}
	return nil
}

func (r *Route) Render(w http.ResponseWriter, name string, data interface{}) error {
	group := r.group
	for group != nil {
		if group.Renderer != nil {
			return group.Renderer.Render(w, name, data)
		}
		group = group.parent
	}
	return errors.New("renderer is nil")
}

func (r *Route) ErrorHandler(err error, c Context) {
	group := r.group
	for group != nil {
		if group.ErrorHandler != nil {
			group.ErrorHandler(err, c)
			return
		}
		group = group.parent
	}
}

func (r *Route) NotFoundHandler(c Context) error {
	return r.group.engine.notFoundRoute.Handle(c)
}

func (r *Route) Handle(c Context) error {
	if len(r.Handlers) == 0 {
		return nil
	}
	return r.Handlers[len(r.Handlers)-1](c)
}

func (r *Route) URL(args ...interface{}) string {
	uri := new(bytes.Buffer)
	path := r.Path
	lstart := 0
	for i, arg := range args {
		pname := r.pnames[i]
		if pname.start > lstart {
			uri.WriteString(path[lstart:pname.start])
		}
		uri.WriteString(fmt.Sprintf("%v", arg))
		lstart = pname.end
	}
	return uri.String()
}

func (r *Route) String() string {
	return fmt.Sprintf("[DEBUG] %-6s %s%-30s --> %s (%d handlers)\n", r.Method, r.Host, r.Path, r.Name, len(r.Handlers))
}
