package forest

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
)

type (
	routeParam struct {
		start int
		end   int
		name  string
	}
	Route struct {
		group    *Group
		pnames   []routeParam
		Name     string
		Desc     string
		Host     string
		Path     string
		Method   string
		Handlers []HandlerFunc
	}
	Routes []*Route
	Router struct {
		host     *node
		hosts    map[string]*node
		routes   map[string]*Route
		maxParam int
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

func (r *Router) Insert(host, method, path string) *Route {
	key := host + method + path
	if route, ok := r.routes[key]; ok {
		return route
	}

	root := r.host
	if host != "" {
		if r.hosts == nil {
			r.hosts = make(map[string]*node, 0)
		}
		h, ok := r.hosts[host]
		if !ok {
			h = &node{}
			r.hosts[host] = h
		}
		root = h
	}

	route := &Route{
		Host:   host,
		Method: method,
		Path:   path,
	}
	root.insert(route)

	if l := len(route.pnames); l > r.maxParam {
		r.maxParam = l
	}

	r.routes[key] = route
	return route
}

func (r *Router) Find(host, method, path string, pvalues []string) (route *Route, found bool) {
	root := r.host
	if host != "" {
		if h, ok := r.hosts[host]; ok {
			root = h
		}
	}
	n := root.find(path, 0, pvalues)
	if n == nil || n.routes == nil {
		return nil, false
	}
	return n.routes.find(method), len(n.routes) > 0
}

func NewRouter() *Router {
	return &Router{host: &node{}, routes: make(map[string]*Route)}
}

func (r *Route) Named(name string, desc ...string) *Route {
	prefix := ""
	if r.group.Name != "" {
		prefix = r.group.Name + "."
	}
	r.Name = prefix + name
	if len(desc) > 0 {
		r.Desc = desc[0]
	}
	return r
}

func (r *Route) Forest() *Forest {
	return r.group.forest
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
	return r.group.forest.notFoundRoute.Handle(c)
}

func (r *Route) Handle(c Context) error {
	return r.Last()(c)
}

func (r *Route) Last() HandlerFunc {
	if len(r.Handlers) == 0 {
		return nil
	}
	return r.Handlers[len(r.Handlers)-1]
}

func (r *Route) URL(args ...interface{}) string {
	if len(args) == 0 {
		return r.Path
	}
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
	if lstart < len(path) {
		uri.WriteString(path[lstart:])
	}
	return uri.String()
}

func (r *Route) String() string {
	return fmt.Sprintf("[DEBUG] %-6s %s%-30s --> %s (%d handlers)\n", r.Method, r.Host, r.Path, r.Name, len(r.Handlers))
}
