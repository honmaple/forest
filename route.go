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
		Name     string
		desc     string
		host     string
		path     string
		method   string
		group    *Group
		pnames   []routeParam
		handlers []HandlerFunc
	}
	Routes []*Route
)

func (rs Routes) find(method string) *Route {
	for _, r := range rs {
		if r.Method() == method {
			return r
		}
	}
	return nil
}

func (r *Route) Named(name string, desc ...string) *Route {
	prefix := ""
	if r.group.Name != "" {
		prefix = r.group.Name + "."
	}
	r.Name = prefix + name
	if len(desc) > 0 {
		r.desc = desc[0]
	}
	return r
}

func (r *Route) Desc() string {
	return r.desc
}

func (r *Route) Host() string {
	return r.host
}

func (r *Route) Path() string {
	return r.path
}

func (r *Route) Method() string {
	return r.method
}

func (r *Route) Handlers() []HandlerFunc {
	return r.handlers
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

func (r *Route) ErrorHandle(err error, c Context) {
	group := r.group
	for group != nil {
		if group.ErrorHandler != nil {
			group.ErrorHandler(err, c)
			return
		}
		group = group.parent
	}
}

// func (r *Route) NotFoundHandle(c Context) error {
//	return r.group.forest.router.notFoundRoute.Handle(c)
// }

// func (r *Route) Handle(c Context) error {
//	return r.Last()(c)
// }

// func (r *Route) Last() HandlerFunc {
//	if len(r.handlers) == 0 {
//		return nil
//	}
//	return r.handlers[len(r.handlers)-1]
// }

func (r *Route) URL(args ...interface{}) string {
	if len(args) == 0 {
		return r.path
	}
	uri := new(bytes.Buffer)
	path := r.path
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
	return fmt.Sprintf("[DEBUG] %-6s %s%-30s --> %s (%d handlers)\n", r.Method(), r.Host(), r.Path(), r.Name, len(r.Handlers()))
}
