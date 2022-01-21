package forest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/honmaple/forest/render"
	"github.com/stretchr/testify/assert"
)

type testStruct struct {
	Path  string `json:"path"`
	Value int    `json:"value"`
}

var (
	testString     = "test ok"
	testJSONString = "{\"path\":\"/json\",\"value\":1}"
)

func testJSON() H {
	return H{"path": "/json", "value": 1}
}

func TestContextRenderString(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := NewContext(req, rec)
	err := c.String(http.StatusOK, testString)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, render.ContentTypeTextCharsetUTF8, rec.Header().Get("Content-Type"))
	assert.Equal(t, testString, rec.Body.String())
}

func TestContextRenderJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := NewContext(req, rec)
	err := c.JSON(http.StatusCreated, testJSON())
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, render.ContentTypeJSONCharsetUTF8, rec.Header().Get("Content-Type"))
	assert.Equal(t, testJSONString+"\n", rec.Body.String())
}

func TestContextBind(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(testJSONString))
	c := NewContext(req, nil)
	u := new(testStruct)

	req.Header.Add(render.ContentType, render.ContentTypeJSON)
	err := c.Bind(u)
	assert.NoError(t, err)
	assert.Equal(t, &testStruct{"/json", 1}, u)

	testHeaders := map[string]string{
		render.ContentType: render.ContentTypeJSON,
		"Test-Header":      "header",
	}
	for k, v := range testHeaders {
		req.Header.Add(k, v)
	}
	headers := make(map[string]string)
	err = c.BindHeader(headers)
	assert.Equal(t, testHeaders, headers)
}

func TestContextBindParam(t *testing.T) {
	router := New()
	router.GET("/:var1/:var2/{var3:int}", func(c Context) error {
		return c.JSON(200, c.Params())
	})
	req := httptest.NewRequest(http.MethodGet, "/1/2/3", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, 200, rec.Code)
	dst := map[string]string{}
	err := json.NewDecoder(rec.Body).Decode(&dst)
	assert.NoError(t, err)
	assert.Equal(t, dst["var1"], "1")
	assert.Equal(t, dst["var2"], "2")
	assert.Equal(t, dst["var3"], "3")
}
