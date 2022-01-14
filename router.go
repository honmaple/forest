package forest

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"strconv"
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

type Params map[string]string

func (p Params) Has(key string) bool {
	_, ok := p[key]
	return ok
}

func (p Params) String(key string) (val string) {
	if val, ok := p[key]; ok {
		return val
	}
	return
}

func (p Params) Int(key string) (val int) {
	if str, ok := p[key]; ok {
		val, err := strconv.Atoi(str)
		if err == nil {
			return val
		}
	}
	return
}

func (p Params) Uint(key string) (val uint) {
	if str, ok := p[key]; ok {
		val, err := strconv.ParseUint(str, 10, 0)
		if err == nil {
			return uint(val)
		}
	}
	return
}

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
