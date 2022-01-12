package forest

import (
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"

	"github.com/honmaple/forest/render"
)

type H map[string]interface{}

type Context interface {
	Request() *http.Request
	Response() *Response

	XML(int, interface{}) error
	JSON(int, interface{}) error
	JSONP(int, string, interface{}) error
	HTML(int, string) error
	String(int, string, ...interface{}) error
	Blob(int, string, []byte) error
	Render(int, render.Renderer) error
	RenderHTML(int, string, interface{}) error
	Redirect(int, string) error

	Param(string) string
	Params() map[string]string
	File(string) error
	Logger() Logger
	Next() error

	Get(string) interface{}
	Set(string, interface{})

	// Bind(interface{})
}

type context struct {
	response *Response
	request  *http.Request
	engine   *Engine
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

func (c *context) Param(key string) string {
	for i, p := range c.route.pnames {
		if p.key == key {
			return c.pvalues[i]
		}
	}
	return ""
}

func (c *context) Params() map[string]string {
	if len(c.route.pnames) == 0 {
		return nil
	}
	params := make(map[string]string)
	for i, p := range c.route.pnames {
		params[p.key] = c.pvalues[i]
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

func (c *context) Render(code int, r render.Renderer) error {
	c.response.WriteHeader(code)
	return r.Render(c.response)
}

func (c *context) RenderHTML(code int, name string, data interface{}) error {
	c.response.WriteHeader(code)
	return c.route.Render(c.response, name, data)
}

func (c *context) Blob(code int, contentType string, data []byte) error {
	c.response.WriteHeader(code)
	return render.Blob(c.response, contentType, data)
}

func (c *context) XML(code int, data interface{}) error {
	c.response.WriteHeader(code)
	return render.XML(c.response, data)
}

func (c *context) JSON(code int, data interface{}) error {
	c.response.WriteHeader(code)
	return render.JSON(c.response, data)
}

func (c *context) JSONP(code int, callback string, data interface{}) error {
	c.response.WriteHeader(code)
	return render.JSONP(c.response, callback, data)
}

func (c *context) String(code int, format string, args ...interface{}) error {
	c.response.WriteHeader(code)
	return render.Text(c.response, sprintf(format, args...))
}

func (c *context) HTML(code int, data string) error {
	c.response.WriteHeader(code)
	return render.HTML(c.response, data)
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
			return
		}
	}
	http.ServeContent(c.Response(), c.Request(), fi.Name(), fi.ModTime(), f)
	return
}

func (c *context) Logger() Logger {
	return c.route.Logger()
}

func (c *context) reset(r *http.Request, w http.ResponseWriter) {
	c.request = r
	c.response.reset(w)
	c.pvalues = c.pvalues[:0]
	c.index = -1
}

func (c *context) Next() (err error) {
	c.index++
	if c.index < len(c.route.Handlers) {
		err = c.route.Handlers[c.index](c)
	}
	if err != nil {
		c.route.ErrorHandler(err, c)
		return nil
	}
	return
}
