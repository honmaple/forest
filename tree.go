package forest

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

const (
	pathChar   = '*'
	paramChar  = ':'
	slashChar  = '/'
	escapeChar = '\\'
)

const (
	skind uint8 = iota // static path
	pkind              // path with params
	rkind              // path with regexp
	akind              // path with anything
)

var ptypes = map[string]uint8{
	"int":   0,
	"float": 0,
	"":      1,
	"str":   2,
	"path":  3,
}

type (
	node struct {
		prefix   string
		kind     uint8
		pname    string
		ptype    string
		regex    *regexp.Regexp
		children [akind + 1]nodes
		routes   map[string]*Route
	}
	nodes []*node
)

func (ns nodes) Sort()         { sort.Sort(ns) }
func (ns nodes) Len() int      { return len(ns) }
func (ns nodes) Swap(i, j int) { ns[i], ns[j] = ns[j], ns[i] }
func (ns nodes) Less(i, j int) bool {
	ni := ns[i]
	nj := ns[j]
	switch ni.kind {
	case skind:
		return ni.prefix[0] < nj.prefix[0]
	case pkind:
		return ptypes[ni.ptype] < ptypes[nj.ptype]
	}
	return false
}

func (ns nodes) findChild(label byte) *node {
	num := len(ns)
	if num == 0 {
		return nil
	}
	idx := 0
	i, j := 0, num-1
	for i <= j {
		idx = i + (j-i)/2
		if label > ns[idx].prefix[0] {
			i = idx + 1
		} else if label < ns[idx].prefix[0] {
			j = idx - 1
		} else {
			i = num
		}
	}
	if ns[idx].prefix[0] != label {
		return nil
	}
	return ns[idx]
}

func commonPrefix(a, b string) int {
	minlen := len(a)
	if len(b) < minlen {
		minlen = len(b)
	}
	i := 0
	for i < minlen && a[i] == b[i] {
		i++
	}
	return i
}

func (s *node) insertStatic(path string, route *Route) *node {
	root := s
	prefix := path

	for {
		sl := len(prefix)
		pl := len(root.prefix)
		if sl == 0 {
			return root
		}
		if pl == 0 {
			root.kind = skind
			root.prefix = prefix
			root.addRoute(route)
			return root
		}
		cl := commonPrefix(root.prefix, prefix)
		// 没有相同前缀
		if cl == 0 {
			if child := root.findChild(skind, "", prefix[0]); child != nil {
				root = child
				continue
			}
			child := newTree(skind, "", "", prefix, route)
			root.addChild(child)
			return child
		}

		// 有相同前缀, 但节点不存在, 需要分裂, 父节点变子节点
		if cl < pl {
			child := newTree(root.kind, root.pname, root.ptype, root.prefix[cl:], nil)
			child.routes = root.routes
			child.children = root.children

			root.kind = skind
			root.prefix = root.prefix[:cl]
			root.children = [akind + 1]nodes{}
			root.routes = make(map[string]*Route)
			root.addChild(child)

			if cl == sl {
				// /user和/us
				root.addRoute(route)
				return root
			}
			// /user和/usad
			child = newTree(skind, "", "", prefix[cl:], route)
			root.addChild(child)
			return child
		}

		if cl < sl {
			// indices
			// 有相同前缀, 并且节点存在 /user和/user123
			prefix = prefix[cl:]
			if child := root.findChild(skind, "", prefix[0]); child != nil {
				root = child
				continue
			}
			child := newTree(skind, "", "", prefix, route)
			root.addChild(child)
			return child
		} else {
			// 相同节点
			root.addRoute(route)
		}
		return root
	}
	return root
}

func (s *node) insertPath(pname string, route *Route) *node {
	child := s.findChild(akind, "", '*')
	if child == nil {
		child = newTree(akind, pname, "", "*", route)
		s.addChild(child)
	}
	return child
}

func (s *node) insertParam(pname, ptype string, route *Route) *node {
	kind := pkind
	if _, ok := ptypes[ptype]; !ok {
		kind = rkind
	}
	child := s.findChild(kind, ptype, ':')
	if child == nil {
		child = newTree(kind, pname, ptype, ":", route)
		s.addChild(child)
	}
	return child
}

func (s *node) insert(route *Route) {
	root := s
	path := route.Path

	l := len(path)
	lstart, start := 0, 0
	for start < l {
		switch path[start] {
		case '{':
			if start > 0 && path[start-1] == '\\' {
				continue
			}

			e := start + 1
			for ; e < l && path[e] != '/' && path[e] != '}'; e++ {
			}
			if e == l || path[e] != '}' {
				panic("forest: route param closing delimiter '}' is missing")
			}
			if start > lstart {
				root = root.insertStatic(path[lstart:start], nil)
			}

			params := strings.SplitN(path[start+1:e], ":", 2)
			if len(params) == 1 && params[0] == "" {
				panic("forest: route param name is missing")
			}
			pname := params[0]
			ptype := "str"
			if len(params) > 1 {
				ptype = params[1]
			}
			if e == l-1 {
				root = root.insertParam(pname, ptype, route)
			} else {
				root = root.insertParam(pname, ptype, nil)
			}
			lstart, start = e+1, e+1
		case ':':
			if start > 0 && path[start-1] == '\\' {
				continue
			}

			e := start + 1
			for ; e < l && path[e] != '/'; e++ {
			}
			if e == start+1 {
				panic("forest: route param name is missing")
			}
			if start > lstart {
				root = root.insertStatic(path[lstart:start], nil)
			}
			if e >= l {
				root = root.insertParam(path[start+1:e], "", route)
			} else {
				root = root.insertParam(path[start+1:e], "", nil)
			}
			lstart, start = e, e
		case '*':
			if start > 0 && path[start-1] == '\\' {
				continue
			}

			e := start + 1
			for ; e < l && path[e] != '/'; e++ {
			}
			if start > lstart {
				root = root.insertStatic(path[lstart:start], nil)
			}
			pname := "*"
			if e > start+1 {
				pname = path[start+1 : e]
			}
			root = root.insertPath(pname, route)

			lstart = len(path)
			start = lstart
		default:
			start++
		}
	}
	if start > lstart {
		root.insertStatic(path[lstart:start], route)
	}
}

func (s *node) addRoute(route *Route) {
	if route == nil {
		return
	}
	if s.routes == nil {
		s.routes = make(map[string]*Route)
	}
	if _, ok := s.routes[route.Method]; ok {
		// panic("route has been exists")
	} else {
		s.routes[route.Method] = route
	}
}

func (s *node) addChild(child *node) {
	s.children[child.kind] = append(s.children[child.kind], child)
	s.children[child.kind].Sort()
}

func (s *node) findChild(kind uint8, ptype string, l byte) *node {
	for _, child := range s.children[kind] {
		if child.prefix[0] == l && child.ptype == ptype {
			return child
		}
	}
	return nil
}

func (s *node) isLeaf() bool {
	for i := range s.children {
		if len(s.children[i]) > 0 {
			return false
		}
	}
	return true
}

func (s *node) matchChild(path string, c *context) *node {
	if path == "" {
		return s
	}
	for kind, cs := range s.children {
		if len(cs) == 0 {
			continue
		}

		label := path[0]
		switch uint8(kind) {
		case skind:
			child := cs.findChild(label)
			if child == nil {
				continue
			}
			if t := child.match(path, c); t != nil {
				return t
			}
		default:
			for _, child := range cs {
				if t := child.match(path, c); t != nil {
					return t
				}
			}
		}
	}
	return nil
}

func (s *node) match(path string, c *context) *node {
	switch s.kind {
	case skind:
		if !strings.HasPrefix(path, s.prefix) {
			return nil
		}
		return s.matchChild(path[len(s.prefix):], c)
	case pkind:
		i := 0
		isLeaf := s.isLeaf()
		switch s.ptype {
		case "":
			i = strings.IndexByte(path, '/')
			if i == 0 {
				return nil
			}
			if i == -1 {
				c.params[s.pname] = path
				return s
			}
		case "str":
			for i = 0; i < len(path); i++ {
				if path[i] == '/' {
					break
				}
				if isLeaf {
					continue
				}
				if t := s.matchChild(path[i:], c); t != nil {
					c.params[s.pname] = path[:i]
					return t
				}
			}
			if i == 0 {
				return nil
			}
			if i == len(path) {
				c.params[s.pname] = path[:i]
				return s
			}
		case "int":
			i = strings.IndexFunc(path, func(r rune) bool {
				return !unicode.IsDigit(r)
			})
			if i == 0 {
				return nil
			}
			if i == -1 {
				c.params[s.pname] = path
				return s
			}
		case "float":
			dot := false
			i = strings.IndexFunc(path, func(r rune) bool {
				if r == '.' {
					if dot {
						return true
					}
					dot = true
					return false
				}
				return !unicode.IsDigit(r)
			})
			if i == 0 {
				return nil
			}
			if i == -1 {
				c.params[s.pname] = path
				return s
			}
		case "path":
			if isLeaf {
				c.params[s.pname] = path
				return s
			}
			for i = 1; i < len(path); i++ {
				if t := s.matchChild(path[i:], c); t != nil {
					c.params[s.pname] = path[:i]
					return t
				}
			}
			if i == len(path) {
				c.params[s.pname] = path
				return s
			}
		}
		if isLeaf || i == len(path) {
			return nil
		}
		if t := s.matchChild(path[i:], c); t != nil {
			c.params[s.pname] = path[:i]
			return t
		}
		return nil
	case rkind:
		if s.regex == nil {
			return nil
		}
		e := strings.IndexByte(path, '/')
		if e == -1 {
			e = len(path)
		}
		is := s.regex.FindStringIndex(path[:e])
		if len(is) == 0 || is[0] > 0 || is[1] < e {
			return nil
		}
		i := is[1]
		if t := s.matchChild(path[i:], c); t != nil {
			c.params[s.pname] = path[:i]
			return t
		}
		return nil
	case akind:
		c.params[s.pname] = path
		return s
	}
	return nil
}

func (s *node) search(method, path string, c *context) (*Route, bool) {
	root := s.match(path, c)
	if root == nil || root.routes == nil {
		return nil, false
	}
	return root.routes[method], len(root.routes) > 0
}

func (s *node) Print(l int) {
	routes := make(map[string]string)
	for _, v := range s.routes {
		routes[v.Method] = v.Path
	}
	fmt.Print(strings.Repeat(" ", l))
	fmt.Printf("%s", s.prefix)
	if len(routes) > 0 {
		fmt.Printf(" %+v\n", routes)
	} else {
		fmt.Print(" nil \n")
	}
	for i := range s.children {
		for _, child := range s.children[i] {
			child.Print(l + 2)
		}
	}
}

func newTree(kind uint8, pname, ptype, prefix string, route *Route) *node {
	t := &node{
		kind:   kind,
		pname:  pname,
		ptype:  ptype,
		prefix: prefix,
	}
	t.addRoute(route)
	if kind == rkind && ptype != "" {
		t.regex = regexp.MustCompile(ptype)
	}
	return t
}
