package forest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testRequest(method, path string, ser http.Handler) (int, string) {
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	ser.ServeHTTP(rec, req)
	return rec.Code, rec.Body.String()
}

func TestGroup(t *testing.T) {
	router := New()
	h := func(c Context) error { return nil }
	group := router.Group("/1", h)
	group.Use(h)

	assert.Len(t, group.middlewares, 2)
	assert.Equal(t, "/1", group.prefix)
	assert.Equal(t, router, group.engine)

	group2 := group.Group("/2")
	group2.Use(h, h)

	assert.Len(t, group2.middlewares, 4)
	assert.Equal(t, "/1/2", group2.prefix)
	assert.Equal(t, router, group2.engine)
}

func TestGroupMount(t *testing.T) {
	r := New()
	g := r.Group("/group")

	group := NewGroup()
	group.GET("/2", func(c Context) error { return c.Status(200) })
	group.GET("/{var}", func(c Context) error { return c.String(201, c.Param("var")) })

	g.Mount("/1", group)

	c, _ := testRequest(http.MethodGet, "/group/1/2", r)
	assert.Equal(t, 200, c)

	c, b := testRequest(http.MethodGet, "/group/1/ss", r)
	assert.Equal(t, 201, c)
	assert.Equal(t, "ss", b)

	r.MountGroup("/3", group)
	c, b = testRequest(http.MethodGet, "/3/ss", r)
	assert.Equal(t, 201, c)
	assert.Equal(t, "ss", b)
}

func TestGroupMiddleware(t *testing.T) {
	r := New()
	g := r.Group("/group")
	h := func(c Context) error { return c.Status(200) }
	m1 := func(c Context) error {
		return c.Next()
	}
	m2 := func(c Context) error {
		return c.Next()
	}
	m3 := func(c Context) error {
		return c.Next()
	}
	m4 := func(c Context) error {
		return NewError(404)
	}
	m5 := func(c Context) error {
		return NewError(405)
	}
	g.Use(m1, m2, m3)
	g.GET("/200", h)
	g.GET("/404", h, m4)
	g.GET("/405", h, m5)

	c, _ := testRequest(http.MethodGet, "/group/200", r)
	assert.Equal(t, 200, c)
	c, _ = testRequest(http.MethodGet, "/group/404", r)
	assert.Equal(t, 404, c)
	c, _ = testRequest(http.MethodGet, "/group/405", r)
	assert.Equal(t, 405, c)
}

func TestGroupBadMethod(t *testing.T) {
	router := New()
	h := func(Context) error { return nil }
	assert.Panics(t, func() {
		router.Add(" GET", "/", h)
	})
	assert.Panics(t, func() {
		router.Add("GET ", "/", h)
	})
	assert.Panics(t, func() {
		router.Add("", "/", h)
	})
	assert.Panics(t, func() {
		router.Add("PO ST", "/", h)
	})
	assert.Panics(t, func() {
		router.Add("1GET", "/", h)
	})
	assert.Panics(t, func() {
		router.Add("PATCh", "/", h)
	})
}

func TestGroupCustomHandler(t *testing.T) {
	router := New()

	group := router.Group("/group")
	group.GET("/1", func(Context) error { return nil })

	c, b := testRequest(http.MethodGet, "/group/2", router)
	assert.Equal(t, 404, c)
	assert.Equal(t, ErrNotFound.Error(), b)

	router.NotFound(func(c Context) error {
		return c.String(404, "404 1")
	})
	c, b = testRequest(http.MethodGet, "/group/2", router)
	assert.Equal(t, 404, c)
	assert.Equal(t, "404 1", b)

	c, b = testRequest(http.MethodPost, "/group/1", router)
	assert.Equal(t, 405, c)
	assert.Equal(t, ErrMethodNotAllowed.Error(), b)

	router.MethodNotAllowed(func(c Context) error {
		return c.String(405, "405 2")
	})
	c, b = testRequest(http.MethodPost, "/group/1", router)
	assert.Equal(t, 405, c)
	assert.Equal(t, "405 2", b)

	group.GET("/2", func(Context) error { return NewError(404, ":404:") })
	c, b = testRequest(http.MethodGet, "/group/2", router)
	assert.Equal(t, 404, c)
	assert.Equal(t, "code=404, message=:404:", b)
}

func TestGroupCustomErrorHandler(t *testing.T) {
	router := New()

	httpErr := NewError(500)
	router.ErrorHandler = func(err error, c Context) {
		assert.Equal(t, err, httpErr)
		c.String(200, "200")
	}

	group := router.Group("/group")
	group.GET("/1", func(Context) error { return httpErr })

	c, b := testRequest(http.MethodGet, "/group/1", router)
	assert.Equal(t, 200, c)
	assert.Equal(t, "200", b)

	group.ErrorHandler = func(err error, c Context) {
		assert.Equal(t, err, httpErr)
		c.String(500, "500")
	}
	c, b = testRequest(http.MethodGet, "/group/1", router)
	assert.Equal(t, 500, c)
	assert.Equal(t, "500", b)
}

func TestGroupRouteName(t *testing.T) {
	router := New()
	group := router.Group("/group")
	group.Name = "group"
	group.GET("/1", func(Context) error { return nil }).Name = "handler1"
	group.GET("/2", func(Context) error { return nil }).Name = "handler2"
	group.GET("/3/:var", func(Context) error { return nil }).Name = "handler3"
	group.GET("/4/:var1/1/:var2", func(Context) error { return nil }).Name = "handler4"

	assert.Equal(t, router.URL("handler1"), "/group/1")
	assert.Equal(t, router.URL("handler2"), "/group/2")
	assert.Equal(t, router.URL("handler3", "var1"), "/group/3/var1")
	assert.Equal(t, router.URL("handler4", "var1"), "/group/4/var1/1/:var2")
	assert.Equal(t, router.URL("handler4", "var1", "var2"), "/group/4/var1/1/var2")
}
