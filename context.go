package forest

import (
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"sync"

	"github.com/honmaple/forest/binder"
	"github.com/honmaple/forest/render"
)

type H map[string]interface{}

type Context interface {
	Forest() *Forest
	Logger() Logger
	Route() *Route
	Request() *http.Request
	Response() *Response

	Next() error
	NextWith(Context) error

	Get(string) interface{}
	Set(string, interface{})

	Param(string) string
	Params() map[string]string

	FormParam(string, ...string) string
	FormParams() (url.Values, error)

	QueryParam(string, ...string) string
	QueryParams() url.Values

	Cookie(string, ...*http.Cookie) (*http.Cookie, error)
	Cookies() []*http.Cookie
	SetCookie(*http.Cookie)

	Bind(interface{}) error
	BindWith(interface{}, binder.Binder) error
	BindParams(interface{}) error
	BindHeader(interface{}) error

	XML(int, interface{}) error
	JSON(int, interface{}) error
	JSONP(int, string, interface{}) error
	HTML(int, string) error
	Bytes(int, []byte) error
	String(int, string, ...interface{}) error
	Blob(int, string, []byte) error
	Render(int, string, interface{}) error
	RenderWith(int, render.Renderer) error

	File(string) error
	FileFromFS(string, http.FileSystem) error

	URL(string, ...interface{}) string
	Status(int) error
	Redirect(int, string, ...interface{}) error
}

type context struct {
	response  *Response
	request   *http.Request
	params    *contextParams
	storeLock sync.RWMutex
	store     map[string]interface{}
	query     url.Values
	route     *Route
	index     int
}

func (c *context) Forest() *Forest {
	return c.route.Forest()
}

func (c *context) Route() *Route {
	return c.route
}

func (c *context) Request() *http.Request {
	return c.request
}

func (c *context) Response() *Response {
	return c.response
}

func (c *context) Set(key string, value interface{}) {
	c.storeLock.Lock()
	defer c.storeLock.Unlock()

	if c.store == nil {
		c.store = make(map[string]interface{})
	}
	c.store[key] = value
}

func (c *context) Get(key string) interface{} {
	c.storeLock.RLock()
	defer c.storeLock.RUnlock()

	return c.store[key]
}

func (c *context) Param(name string) string {
	for i, p := range c.route.pnames {
		if i < len(c.params.pvalues) && p.name == name {
			return c.params.pvalues[i]
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
		if i < len(c.params.pvalues) {
			params[p.name] = c.params.pvalues[i]
		}
	}
	return params
}

func (c *context) FormParam(key string, defaults ...string) string {
	v := c.request.FormValue(key)
	if v == "" && len(defaults) > 0 {
		return defaults[0]
	}
	return v
}

func (c *context) FormParams() (url.Values, error) {
	if err := binder.ParseForm(c.request, 0); err != nil {
		return nil, err
	}
	return c.request.Form, nil
}

func (c *context) QueryParam(key string, defaults ...string) string {
	v := c.QueryParams().Get(key)
	if v == "" && len(defaults) > 0 {
		return defaults[0]
	}
	return v
}

func (c *context) QueryParams() url.Values {
	if c.query == nil {
		c.query = c.request.URL.Query()
	}
	return c.query
}

func (c *context) Cookie(name string, defaults ...*http.Cookie) (*http.Cookie, error) {
	v, err := c.request.Cookie(name)
	if err != nil {
		return nil, err
	}
	if v == nil && len(defaults) > 0 {
		return defaults[0], nil
	}
	return v, nil
}

func (c *context) Cookies() []*http.Cookie {
	return c.request.Cookies()
}

func (c *context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.response, cookie)
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

func (c *context) Bytes(code int, data []byte) error {
	return render.Bytes(c.response, code, data)
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

func (c *context) URL(name string, args ...interface{}) string {
	return c.route.Forest().URL(name, args...)
}

func (c *context) Redirect(code int, url string, args ...interface{}) error {
	if code < 300 || code > 308 {
		return nil
	}
	if !strings.HasPrefix(url, "/") && !strings.Contains(url, "://") {
		url = c.route.Forest().URL(url, args...)
	}
	c.response.Header().Set("Location", url)
	c.response.WriteHeader(code)
	return nil
}

func (c *context) File(file string) (err error) {
	return c.FileFromFS(filepath.Base(file), http.Dir(filepath.Dir(file)))
}

func (c *context) FileFromFS(file string, fs http.FileSystem) error {
	const indexPage = "index.html"
	if file == "" {
		file = indexPage
	}

	file = filepath.Clean(file)
	f, err := fs.Open(file)
	if err != nil {
		return NotFoundHandler(c)
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return NotFoundHandler(c)
	}
	url := c.Request().URL.Path
	if fi.IsDir() {
		if url == "" || url[len(url)-1] != '/' {
			return c.Redirect(http.StatusMovedPermanently, url+"/")
		}
		index, err := fs.Open(filepath.Join(file, indexPage))
		if err != nil {
			return NotFoundHandler(c)
		}
		defer index.Close()

		indexfi, err := f.Stat()
		if err != nil {
			return NotFoundHandler(c)
		}
		f, fi = index, indexfi
	}
	// index.html is Dir
	if fi.IsDir() {
		return NotFoundHandler(c)
	}
	http.ServeContent(c.Response(), c.Request(), fi.Name(), fi.ModTime(), f)
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
	if c.index < len(c.route.handlers) {
		err = c.route.handlers[c.index](ctx)
	}
	if err != nil {
		c.route.ErrorHandle(err, ctx)
		return nil
	}
	return
}

func (c *context) reset(r *http.Request, w http.ResponseWriter) {
	c.response.reset(w)
	c.request = r
	c.store = nil
	c.index = -1
	c.params.reset(0)
}

type contextParams struct {
	pindex  int
	pvalues []string
}

func (m *contextParams) reset(pindex int) {
	m.pindex = pindex
	for ; pindex < len(m.pvalues); pindex++ {
		m.pvalues[pindex] = ""
	}
}

func NewContext(r *http.Request, w http.ResponseWriter) Context {
	c := &context{response: NewResponse(w)}
	c.reset(r, w)
	return c
}
