package forest

import (
	"errors"
	"net/http"
)

type (
	router struct {
		root   *node
		routes map[string]*Route
	}
	Route struct {
		Name        string        `json:"name"`
		Host        string        `json:"host"`
		Path        string        `json:"path"`
		Method      string        `json:"method"`
		Handler     HandlerFunc   `json:"-"`
		Middlewares []HandlerFunc `json:"-"`

		group *Group
	}
	Router interface {
		Host(string, string, ...HandlerFunc) Router
		Group(string, ...HandlerFunc) Router
		Mount(string, *Group)
		Use(...HandlerFunc)
		GET(string, HandlerFunc, ...HandlerFunc) *Route
		HEAD(string, HandlerFunc, ...HandlerFunc) *Route
		POST(string, HandlerFunc, ...HandlerFunc) *Route
		PUT(string, HandlerFunc, ...HandlerFunc) *Route
		PATCH(string, HandlerFunc, ...HandlerFunc) *Route
		DELETE(string, HandlerFunc, ...HandlerFunc) *Route
		OPTIONS(string, HandlerFunc, ...HandlerFunc) *Route
		Add(string, string, HandlerFunc, ...HandlerFunc) *Route
		Any(string, HandlerFunc, ...HandlerFunc) []*Route
	}
)

func (r *router) Add(route *Route) {
	r.root.insert(route.Method, route.Path)

	key := route.Method + "-" + route.Path
	r.routes[key] = route
}

func (r *router) Find(method string, path string, params map[string]string) (*Route, bool) {
	n := r.root.search(path, params)
	if n != nil {
		key := method + "-" + n.path
		return r.routes[key], true
	}
	return nil, false
}

func newrouter() *router {
	return &router{
		root:   &node{},
		routes: make(map[string]*Route),
	}
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
