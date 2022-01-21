package forest

import (
	"net/http"
	"net/url"
	"sync"

	"github.com/honmaple/forest/binder"
	"github.com/honmaple/forest/render"
	"os"
	"path/filepath"
)

type H map[string]interface{}

type Context interface {
	Logger() Logger
	Request() *http.Request
	Response() *Response

	Next() error
	NextWith(Context) error

	Get(string) interface{}
	Set(string, interface{})

	Param(string) string
	Params() map[string]string

	Bind(interface{}) error
	BindWith(interface{}, binder.Binder) error
	BindParams(interface{}) error
	BindHeader(interface{}) error

	XML(int, interface{}) error
	JSON(int, interface{}) error
	JSONP(int, string, interface{}) error
	HTML(int, string) error
	String(int, string, ...interface{}) error
	Blob(int, string, []byte) error
	Render(int, string, interface{}) error
	RenderWith(int, render.Renderer) error
	File(string) error

	Status(int) error
	Redirect(int, string) error
}

type context struct {
	response *Response
	request  *http.Request
	pvalues  []string
	store    sync.Map
	query    url.Values
	route    *Route
	index    int
}

func (c *context) Request() *http.Request {
	return c.request
}

func (c *context) Response() *Response {
	return c.response
}

func (c *context) Set(key string, value interface{}) {
	c.store.Store(key, value)
}

func (c *context) Get(key string) interface{} {
	if v, ok := c.store.Load(key); ok {
		return v
	}
	return nil
}

func (c *context) Param(name string) string {
	for i, p := range c.route.pnames {
		if i < len(c.pvalues) && p.name == name {
			return c.pvalues[i]
		}
	}
	return ""
}

func (c *context) Params() map[string]string {
	if c.route == nil || len(c.route.pnames) == 0 {
		return nil
	}
	params := make(map[string]string)
	for i, p := range c.route.pnames {
		if i < len(c.pvalues) {
			params[p.name] = c.pvalues[i]
		}
	}
	return params
}

func (c *context) FormValue(key string) string {
	return c.request.FormValue(key)
}

func (c *context) FormParams() (url.Values, error) {
	return nil, nil
}

func (c *context) QueryParam(key string) string {
	return c.QueryParams().Get(key)
}

func (c *context) QueryParams() url.Values {
	if c.query == nil {
		c.query = c.request.URL.Query()
	}
	return c.query
}

func (c *context) Bind(data interface{}) error {
	return binder.Bind(c.request, data)
}

func (c *context) BindWith(data interface{}, b binder.Binder) error {
	return b.Bind(c.request, data)
}

func (c *context) BindParams(data interface{}) error {
	return binder.Params.Bind(c.Params(), data)
}

func (c *context) BindHeader(data interface{}) error {
	return c.BindWith(data, binder.Header)
}

func (c *context) Render(code int, name string, data interface{}) error {
	c.response.WriteHeader(code)
	return c.route.Render(c.response, name, data)
}

func (c *context) RenderWith(code int, r render.Renderer) error {
	c.response.WriteHeader(code)
	return r.Render(c.response)
}

func (c *context) Blob(code int, contentType string, data []byte) error {
	return render.Blob(c.response, code, contentType, data)
}

func (c *context) XML(code int, data interface{}) error {
	return render.XML(c.response, code, data)
}

func (c *context) JSON(code int, data interface{}) error {
	return render.JSON(c.response, code, data)
}

func (c *context) JSONP(code int, callback string, data interface{}) error {
	return render.JSONP(c.response, code, callback, data)
}

func (c *context) String(code int, format string, args ...interface{}) error {
	return render.Text(c.response, code, sprintf(format, args...))
}

func (c *context) HTML(code int, data string) error {
	return render.HTML(c.response, code, data)
}

func (c *context) Status(code int) error {
	c.response.WriteHeader(code)
	return nil
}

func (c *context) Redirect(code int, url string) error {
	if code < 300 || code > 308 {
		return nil
	}
	c.response.Header().Set("Location", url)
	c.response.WriteHeader(code)
	return nil
}

func (c *context) File(file string) (err error) {
	f, err := os.Open(file)
	if err != nil {
		return c.route.NotFoundHandler(c)
	}
	defer f.Close()

	fi, _ := f.Stat()
	if fi.IsDir() {
		file = filepath.Join(file, "index.html")
		f, err = os.Open(file)
		if err != nil {
			return c.route.NotFoundHandler(c)
		}
		defer f.Close()
		if fi, err = f.Stat(); err != nil {
			return c.route.NotFoundHandler(c)
		}
	}
	http.ServeContent(c.Response(), c.Request(), fi.Name(), fi.ModTime(), f)
	return
}

func (c *context) FileFromFS(file string, fs http.FileSystem) error {
	defer func(old string) {
		c.request.URL.Path = old
	}(c.request.URL.Path)

	c.request.URL.Path = file

	http.FileServer(fs).ServeHTTP(c.response, c.request)
	return nil
}

func (c *context) Logger() Logger {
	return c.route.Logger()
}

func (c *context) Next() error {
	return c.NextWith(c)
}

func (c *context) NextWith(ctx Context) (err error) {
	c.index++
	if c.index < len(c.route.Handlers) {
		err = c.route.Handlers[c.index](ctx)
	}
	if err != nil {
		c.route.ErrorHandler(err, ctx)
		return nil
	}
	return
}

func (c *context) reset(r *http.Request, w http.ResponseWriter) {
	c.request = r
	c.response.reset(w)
	c.pvalues = c.pvalues[:0]
	c.index = -1
}

func NewContext(r *http.Request, w http.ResponseWriter) Context {
	c := &context{response: NewResponse(w)}
	c.reset(r, w)
	return c
}
