package forest

import (
	"net/http"
	"testing"

	testify "github.com/stretchr/testify/assert"
)

type testPath struct {
	path, route string
	params      map[string]string
}

func TestTree(t *testing.T) {
	root := &node{}
	urls := []string{
		"/path",
		"/path/test",
		"/path/{var1}",
		"/path/{var1:int}",
		"/path/{var1:int}-{var2:int}",
		"/path/{var1}/{var2:int}",
		"/path/{var1:path}-1/{var2:int}",
		"/path/{var1:int}/{var2:int}/{var3:path}",
		"/path/{var1:path}/{var2:int}/test/{var3:path}",
		"/path1/:var1",
		"/path1/:var1/test/pre:var2",
		"/path2/*",
		"/path2/*/test",
		"/path3/*var1",
		"/path3/*var1/test",
		"/path3/pre*var1",
	}
	for _, url := range urls {
		root.insert(&Route{Method: http.MethodGet, Path: url})
	}
	// root.Print(0)

	assert := testify.New(t)
	paths := []testPath{
		{"/path", "/path", nil},
		{"/path/test", "/path/test", nil},
		{
			"/path/1", "/path/{var1:int}",
			map[string]string{"var1": "1"},
		},
		{
			"/path/s", "/path/{var1}",
			map[string]string{"var1": "s"},
		},
		{
			"/path/1-3", "/path/{var1:int}-{var2:int}",
			map[string]string{"var1": "1", "var2": "3"},
		},
		{
			"/path/s/1", "/path/{var1}/{var2:int}",
			map[string]string{"var1": "s", "var2": "1"},
		},
		{
			"/path/11c/s/1-1/2", "/path/{var1:path}-1/{var2:int}",
			map[string]string{"var1": "11c/s/1", "var2": "2"},
		},
		{
			"/path/1/5/s/1", "/path/{var1:int}/{var2:int}/{var3:path}",
			map[string]string{"var1": "1", "var2": "5", "var3": "s/1"},
		},
		{
			"/path/s/s/5/test/1/c", "/path/{var1:path}/{var2:int}/test/{var3:path}",
			map[string]string{"var1": "s/s", "var2": "5", "var3": "1/c"},
		},
		{
			"/path1/test", "/path1/:var1",
			map[string]string{"var1": "test"},
		},
		{
			"/path1/test/test/pre1", "/path1/:var1/test/pre:var2",
			map[string]string{"var1": "test", "var2": "1"},
		},
		{
			"/path2/s/1", "/path2/*",
			map[string]string{"*": "s/1"},
		},
		{
			"/path3/s/1/4/c", "/path3/*var1",
			map[string]string{"var1": "s/1/4/c"},
		},
		{
			"/path3/pre/1/4/c", "/path3/pre*var1",
			map[string]string{"var1": "/1/4/c"},
		},
	}
	for _, p := range paths {
		ctx := &context{}
		if v, found := root.search(http.MethodGet, p.path, ctx); v != nil && found {
			ctx.route = v
			assert.Equal(p.route, v.Path, p.path)
		} else {
			assert.Equal(p.route, "nil", p.path)
		}
		params := ctx.Params()
		if len(params) > 0 {
			assert.Equal(p.params, params, p.path)
		} else {
			assert.Nil(p.params)
		}
	}
}
