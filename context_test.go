package forest

import (
	"net/http"
	"testing"

	"github.com/honmaple/forest/render"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
)

var (
	testJSONString = "{\"path\":\"/json\",\"value\":1}"
	testString     = "test ok"
)

func testJSON() H {
	return H{"path": "/json", "value": 1}
}

func TestContextRenderString(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := NewContext(req, rec)
	err := c.String(http.StatusOK, testString)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, render.MIMETextPlainCharsetUTF8, rec.Header().Get("Content-Type"))
		assert.Equal(t, testString, rec.Body.String())
	}
}

func TestContextRenderJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := NewContext(req, rec)
	err := c.JSON(http.StatusCreated, testJSON())
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Equal(t, render.MIMEApplicationJSONCharsetUTF8, rec.Header().Get("Content-Type"))
		assert.Equal(t, testJSONString+"\n", rec.Body.String())
	}
}
