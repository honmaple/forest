package forest

import (
	"errors"
	"net/http"
)

type (
	Route struct {
		Name        string        `json:"name"`
		Host        string        `json:"host"`
		Path        string        `json:"path"`
		Method      string        `json:"method"`
		Handler     HandlerFunc   `json:"-"`
		Middlewares []HandlerFunc `json:"-"`

		group *Group
	}
	Router map[string]*node
)

func (r Router) Insert(route *Route) {
	root, ok := r[route.Host]
	if !ok {
		root = &node{}
		r[route.Host] = root
	}
	root.insert(route)
}

func (r Router) Find(host, method, path string, c *context) (*Route, bool) {
	root, ok := r[host]
	if !ok {
		return nil, false
	}
	return root.search(method, path, c)
}

func newRouter() Router {
	return make(map[string]*node)
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

func (r *Route) NotFoundHandler(c Context) error {
	return r.group.engine.NotFoundHandler(c)
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
