package forest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
)

type H map[string]interface{}

type Context interface {
	Request() *http.Request
	Response() *Response

	HTML(int, string) error
	JSON(int, interface{}) error
	String(int, string, ...interface{}) error
	Render(int, string, interface{}) error
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
	params   map[string]string
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
	value, _ := c.params[key]
	return value
}

func (c *context) Params() map[string]string {
	params := make(map[string]string)
	for k, v := range c.params {
		params[k] = v
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

func (c *context) SetHeader(key string, value string) {
	c.response.Header().Set(key, value)
}

func (c *context) String(code int, format string, values ...interface{}) error {
	c.SetHeader("Content-Type", "text/plain")
	c.response.WriteHeader(code)
	c.response.Write([]byte(fmt.Sprintf(format, values...)))
	return nil
}

func (c *context) JSON(code int, data interface{}) error {
	c.SetHeader("Content-Type", "application/json")
	c.response.WriteHeader(code)
	encoder := json.NewEncoder(c.response)
	if err := encoder.Encode(data); err != nil {
		http.Error(c.response, err.Error(), 500)
	}
	return nil
}

func (c *context) HTML(code int, html string) error {
	c.SetHeader("Content-Type", "text/html")
	c.response.WriteHeader(code)
	c.response.Write([]byte(html))
	return nil
}

func (c *context) Render(code int, name string, data interface{}) error {
	return c.route.Render(c.response, name, data)
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
	c.params = make(map[string]string)
	c.index = -1
}

func (c *context) Next() error {
	c.index++
	if c.index < len(c.route.Middlewares) {
		c.route.Middlewares[c.index](c)
		c.index++
	}
	if c.index == len(c.route.Middlewares) {
		return c.route.Handler(c)
	}
	return nil
}
