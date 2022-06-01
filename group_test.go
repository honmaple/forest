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
	assert.Equal(t, router, group.forest)

	group2 := group.Group("/2")
	group2.Use(h, h)

	assert.Len(t, group2.middlewares, 4)
	assert.Equal(t, "/1/2", group2.prefix)
	assert.Equal(t, router, group2.forest)
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
	g.GET("/404", m4, h)
	g.GET("/405", m5, h)

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
	h := func(Context) error { return nil }
	group.GET("/1", h).Name = "handler1"
	group.GET("/2", h).Name = "handler2"
	group.GET("/3/:var", h).Name = "handler3"
	group.GET("/4/:var1/1/:var2", h).Name = "handler4"

	assert.Equal(t, router.URL("handler1"), "/group/1")
	assert.Equal(t, router.URL("handler2"), "/group/2")
	assert.Equal(t, router.URL("handler3", "var1"), "/group/3/var1")
	assert.Equal(t, router.URL("handler4", "var1"), "/group/4/var1/1/:var2")
	assert.Equal(t, router.URL("handler4", "var1", "var2"), "/group/4/var1/1/var2")

	v1 := router.Group("/v1").Named("v1")
	r1 := v1.GET("/1", h).Named("r1")
	r2 := v1.GET("/2", h).Named("r2")
	assert.Equal(t, router.Route("v1.r1"), r1)
	assert.Equal(t, router.Route("v1.r2"), r2)

	v2 := v1.Group("/v2").Named("v2")
	r3 := v2.GET("/3", h).Named("r3")
	assert.Equal(t, router.Route("v1.v2.r3"), r3)

	v3 := router.Group("/v3")
	v4 := v3.Group("/v4").Named("v4")
	r4 := v4.GET("/4", h).Named("r4")
	assert.Equal(t, router.Route("v4.r4"), r4)
	r4.Name = "r4.1"
	assert.Equal(t, router.Route("r4.1"), r4)
}
