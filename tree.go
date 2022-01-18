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

var (
	ptypes = map[string]uint8{
		"int":   0,
		"float": 0,
		"":      1,
		"str":   2,
		"path":  3,
	}
	matchers = map[string]func(string, string) Matcher{}
)

type (
	node struct {
		prefix   string
		kind     uint8
		matcher  Matcher
		children [akind + 1]nodes
		routes   Routes
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
		return ptypes[ni.matcher.Name()] < ptypes[nj.matcher.Name()]
	}
	return false
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
		// no common prefix
		if cl == 0 {
			if child := root.findStaticChild(prefix[0]); child != nil {
				root = child
				continue
			}
			child := newNode(skind, prefix, route)
			root.addChild(child)
			return child
		}

		// has common prefix, but node is not exists, parent node need split
		if cl < pl {
			child := newNode(root.kind, root.prefix[cl:], nil)
			child.routes = root.routes
			child.matcher = root.matcher
			child.children = root.children

			root.kind = skind
			root.prefix = root.prefix[:cl]
			root.children = [akind + 1]nodes{}
			// root.routes = make(map[string]*Route)
			root.routes = root.routes[:0]
			root.addChild(child)

			if cl == sl {
				// /user和/us
				root.addRoute(route)
				return root
			}
			// /user和/usad
			child = newNode(skind, prefix[cl:], route)
			root.addChild(child)
			return child
		}

		if cl < sl {
			// has common prefix, and node is exists /user和/user123
			prefix = prefix[cl:]
			if child := root.findStaticChild(prefix[0]); child != nil {
				root = child
				continue
			}
			child := newNode(skind, prefix, route)
			root.addChild(child)
			return child
		} else {
			// same node
			root.addRoute(route)
		}
		return root
	}
	return root
}

func (s *node) insertParam(pname, ptype string, route *Route) *node {
	var (
		kind    uint8 = pkind
		label   byte  = ':'
		matcher Matcher
	)
	if ptype == "path" {
		kind = akind
		label = '*'
	} else if IsRegexURLParam(ptype) {
		kind = rkind
	}

	mc, ok := matchers[ptype]
	if ok {
		matcher = mc(pname, ptype)
	} else {
		matcher = paramMatcher(pname, ptype)
	}
	child := s.findParamChild(kind, ptype, label)
	if child == nil {
		child = newNode(kind, string(label), route)
		child.matcher = matcher
		s.addChild(child)
	}
	return child
}

func (s *node) insert(route *Route) {
	root := s
	path := route.Path

	if path == "" {
		path = "/"
	}
	if !strings.HasPrefix(path, "/") {
		panic("forest: route path must startswith '/'")
	}

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
			// route.pnames = append(route.pnames, pname)
			route.pnames = append(route.pnames, routeParam{pname, start, e + 1})
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
			pname := path[start+1 : e]
			// route.pnames = append(route.pnames, pname)
			route.pnames = append(route.pnames, routeParam{pname, start, e})
			if e >= l {
				root = root.insertParam(pname, "", route)
			} else {
				root = root.insertParam(pname, "", nil)
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
			route.pnames = append(route.pnames, routeParam{pname, start, e})
			root = root.insertParam(pname, "path", route)

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
	v := s.routes.find(route.Method)
	if v != nil {
		// panic("forest: path '" + route.Path + "' conflicts with existing route '" + v.Path + "'")
	} else {
		s.routes = append(s.routes, route)
	}
}

func (s *node) addChild(child *node) {
	s.children[child.kind] = append(s.children[child.kind], child)
	s.children[child.kind].Sort()
}

func (s *node) findParamChild(kind uint8, ptype string, l byte) *node {
	for _, child := range s.children[kind] {
		if child.prefix[0] == l {
			if child.matcher == nil || child.matcher.Name() == ptype {
				return child
			}
		}
	}
	return nil
}

func (s *node) findStaticChild(l byte) *node {
	children := s.children[skind]
	num := len(children)
	if num == 0 {
		return nil
	}
	idx := 0
	i, j := 0, num-1
	for i <= j {
		idx = i + (j-i)/2
		if l > children[idx].prefix[0] {
			i = idx + 1
		} else if l < children[idx].prefix[0] {
			j = idx - 1
		} else {
			i = num
		}
	}
	if children[idx].prefix[0] != l {
		return nil
	}
	return children[idx]
}

func (s *node) matchChild(path string, c *context) *node {
	if path == "" {
		return s
	}
	for kind := range s.children {
		label := path[0]
		switch uint8(kind) {
		case skind:
			child := s.findStaticChild(label)
			if child == nil {
				continue
			}
			if t := child.match(path, c); t != nil {
				return t
			}
		default:
			for _, child := range s.children[kind] {
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
		pl := len(s.prefix)
		if len(path) == pl {
			return s
		}
		return s.matchChild(path[pl:], c)
	case pkind, rkind, akind:
		i := 0
		for i < len(path) {
			e, loop := s.matcher.Match(path[i:])
			if e == -1 {
				return nil
			}
			c.pvalues = append(c.pvalues, path[:i+e])
			if e == len(path[i:]) {
				return s
			}
			if child := s.matchChild(path[i+e:], c); child != nil {
				return child
			}
			c.pvalues = c.pvalues[:len(c.pvalues)-1]
			if !loop {
				break
			}
			i = i + e
		}
		return nil
	default:
		return nil
	}
}

func (s *node) search(method, path string, c *context) (*Route, bool) {
	root := s.match(path, c)
	if root == nil || root.routes == nil {
		return nil, false
	}
	return root.routes.find(method), len(root.routes) > 0
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

func newNode(kind uint8, prefix string, route *Route) *node {
	n := &node{
		kind:   kind,
		prefix: prefix,
	}
	n.addRoute(route)
	return n
}

type (
	Matcher interface {
		Name() string
		Match(string) (int, bool)
	}
	pMatcher struct {
		ptype string
		regex *regexp.Regexp
	}
)

func (p *pMatcher) Name() string {
	return p.ptype
}

func (p *pMatcher) Match(path string) (int, bool) {
	switch p.ptype {
	case "":
		return p.match(path)
	case "str":
		return p.matchStr(path)
	case "int":
		return p.matchInt(path)
	case "float":
		return p.matchFloat(path)
	case "path":
		return p.matchPath(path)
	default:
		return p.matchRegex(path)
	}
}

func (p *pMatcher) match(path string) (int, bool) {
	i := strings.IndexByte(path, '/')
	if i == 0 {
		return -1, false
	}
	if i == -1 {
		return len(path), false
	}
	return i, false
}

func (p *pMatcher) matchInt(path string) (int, bool) {
	i := strings.IndexFunc(path, func(r rune) bool {
		return !unicode.IsDigit(r)
	})
	if i == 0 {
		return -1, false
	}
	if i == -1 {
		return len(path), false
	}
	return i, false
}

func (p *pMatcher) matchFloat(path string) (int, bool) {
	dot := false
	i := strings.IndexFunc(path, func(r rune) bool {
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
		return -1, false
	}
	if i == -1 {
		return len(path), false
	}
	return i, false
}

func (p *pMatcher) matchStr(path string) (int, bool) {
	if path[0] == '/' {
		return -1, false
	}
	return 1, true
}

func (p *pMatcher) matchPath(path string) (int, bool) {
	return 1, true
}

func (p *pMatcher) matchRegex(path string) (int, bool) {
	if p.regex == nil {
		return -1, false
	}
	e := strings.IndexByte(path, '/')
	if e == -1 {
		e = len(path)
	}
	is := p.regex.FindStringIndex(path[:e])
	if len(is) == 0 {
		return -1, false
	}
	return is[1], false
}

func paramMatcher(pname, ptype string) Matcher {
	p := &pMatcher{ptype: ptype}
	if IsRegexURLParam(ptype) {
		if ptype[0] != '^' {
			ptype = "^" + ptype
		}
		p.regex = regexp.MustCompile(ptype)
	}
	return p
}

func IsRegexURLParam(ptype string) bool {
	if _, ok := ptypes[ptype]; ok {
		return false
	}
	if _, ok := matchers[ptype]; ok {
		return false
	}
	return true
}

func RegisterURLParam(ptype string, matcher func(string, string) Matcher) {
	matchers[ptype] = matcher
}
