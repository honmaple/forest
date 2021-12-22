package forest

import (
	"testing"

	testify "github.com/stretchr/testify/assert"
)

type testPath struct {
	path, route string
	params      map[string]string
}

func TestRoute(t *testing.T) {
	assert := testify.New(t)

	root := node{}
	urls := []string{
		"/path",
		"/path/test",
		"/path/{var1}",
		"/path/{var1:str}",
		"/path/{var1:int}",
		"/path/{var1:int}-{var2:int}",
		"/path/{var1}/{var2:int}",
		"/path/{var1:path}/{var2:int}",
		"/path/{var1:int}/{var2:int}/{var3:path}",
		"/path/{var1:path}/{var2:int}/test/{var3:path}",
	}
	for _, url := range urls {
		root.insert(url, 0)
	}
	root.Print(0)

	paths := []testPath{
		{"/path", "/path", nil},
		{"/path1/test", "nil", nil},
		{
			"/path/1", "/path/{var1:int}",
			map[string]string{"var1": "1"},
		},
		{
			"/path/s", "/path/{var1:str}",
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
			"/path/11c/s/1/2", "/path/{var1:path}/{var2:int}",
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
	}
	for _, p := range paths {
		params := make(map[string]string)
		if v := root.search(p.path, params); v != nil {
			assert.Equal(p.route, v.path, p.path)
		} else {
			assert.Equal(p.route, "nil", p.path)
		}
		if len(params) > 0 {
			assert.Equal(p.params, params, p.path)
		} else {
			assert.Nil(p.params)
		}
	}
}
