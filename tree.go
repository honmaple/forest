package forest

import (
	"fmt"
	"regexp"
	"strings"
)

type (
	kind uint8
	node struct {
		kind          kind
		prefix        string
		routes        Routes
		matcher       Matcher
		children      [akind + 1]nodes
		hasChild      bool
		hasParamChild bool
	}
	nodes []*node
)

const (
	skind kind = iota // static path
	pkind             // path with params
	akind             // path with anything
)

func isNumeric(a byte) bool {
	return a >= '0' && a <= '9'
}

func isOptional(path string) bool {
	pl := len(path)
	return pl > 1 && path[pl-1] == '?' && path[pl-2] != '\\'
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
			child.hasChild = root.hasChild
			child.hasParamChild = root.hasParamChild

			root.kind = skind
			root.prefix = root.prefix[:cl]
			root.children = [akind + 1]nodes{}
			root.routes = Routes{}
			root.addChild(child)

			if cl == sl {
				// /user then /us
				root.addRoute(route)
				return root
			}
			// /user then /usad
			child = newNode(skind, prefix[cl:], route)
			root.addChild(child)
			return child
		}

		if cl < sl {
			// has common prefix, and node is exists /user then /user123
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
		nkind kind = pkind
		label byte = ':'
	)

	if ptype == "path" {
		nkind = akind
		label = '*'
	}

	child := s.findParamChild(nkind, ptype, label)
	if child == nil {
		child = newNode(nkind, string(label), route)
		if mc, ok := matchers[ptype]; ok {
			child.matcher = mc(pname, ptype)
		} else {
			child.matcher = regexMatcher(pname, ptype)
		}
		s.addChild(child)
	} else {
		child.addRoute(route)
	}
	return child
}

func (s *node) insert(route *Route) {
	root := s
	path := route.Path()

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
				start++
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
			// /path/{var:int} match /path/100 but /path/
			// /path/{var?:int} match /path/100 and /path/
			pname := params[0]
			optional := isOptional(pname)
			if optional {
				pname = pname[:len(pname)-1]
			}
			if len(pname) == 0 {
				panic("forest: route param name is missing")
			}
			ptype := "string"
			if len(params) > 1 {
				ptype = params[1]
			}
			route.pnames = append(route.pnames, routeParam{start: start, end: e + 1, name: pname})

			if e == l-1 {
				if optional || ptype == "path" {
					root.addRoute(route)
				}
				root = root.insertParam(pname, ptype, route)
			} else {
				root = root.insertParam(pname, ptype, nil)
			}
			lstart, start = e+1, e+1
		case ':':
			if start > 0 && path[start-1] == '\\' {
				start++
				continue
			}

			e := start + 1
			for ; e < l && path[e] != '/'; e++ {
			}
			// /path/:var match /path/anyword but /path/
			// /path/:var? match /path/anyword and /path/
			pname := path[start+1 : e]
			optional := isOptional(pname)
			if optional {
				pname = pname[:len(pname)-1]
			}
			if len(pname) == 0 {
				panic("forest: route param name is missing")
			}
			if start > lstart {
				root = root.insertStatic(path[lstart:start], nil)
			}
			route.pnames = append(route.pnames, routeParam{start: start, end: e, name: pname})
			if e >= l {
				if optional {
					root.addRoute(route)
				}
				root = root.insertParam(pname, "", route)
			} else {
				root = root.insertParam(pname, "", nil)
			}
			lstart, start = e, e
		case '*':
			if start > 0 && path[start-1] == '\\' {
				start++
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
			route.pnames = append(route.pnames, routeParam{start: start, end: e, name: pname})
			if e >= l {
				root.addRoute(route)
				root = root.insertParam(pname, "path", route)
			} else {
				root = root.insertParam(pname, "path", nil)
			}
			lstart, start = e, e
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
	s.routes = append(s.routes, route)
	// if r := s.routes.find(route.Method()); r != nil {
	//	panic("forest: path '" + route.Path() + "' conflicts with existing route '" + r.Path() + "'")
	// } else {
	//	s.routes = append(s.routes, route)
	// }
}

func (s *node) addChild(child *node) {
	if child.kind == skind {
		if len(s.children[skind]) == 0 {
			s.children[skind] = make(nodes, 256)
		}
		s.children[skind][child.prefix[0]] = child
	} else {
		s.children[child.kind] = append(s.children[child.kind], child)
		s.hasParamChild = true
	}
	s.hasChild = true
}

func (s *node) findParamChild(kind kind, ptype string, l byte) *node {
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
	if v := s.children[skind]; len(v) == 0 {
		return nil
	} else {
		return v[l]
	}
}

func (s *node) find(path string, paramIndex int, paramValues []string) (result *node) {
	var (
		root   = s
		count  = 0
		pindex = paramIndex
	)
LOOP:
	for {
		e, ok := 0, false
		switch root.kind {
		case skind:
			if count > 0 {
				break LOOP
			}
			pl := len(root.prefix)
			// don't use strings.HasPrefix, it is slower
			if pl <= len(path) {
				for ; e < pl && root.prefix[e] == path[e]; e++ {
				}
			}
			if e != pl {
				return
			}
		default:
			e, ok = root.matcher.Match(path, count, root.hasChild)
			if !ok {
				break LOOP
			}
			paramValues[paramIndex] = path[:e]
			paramIndex++
		}
		if len(path) == e {
			result = root
			return
		}

		if !root.hasChild {
			break
		}

		search := path[e:]
		if child := root.findStaticChild(search[0]); child != nil {
			// avoid recursion when no param children
			if !root.hasParamChild && count == 0 {
				root = child
				path = search
				count = 0
				continue
			}
			if result = child.find(search, paramIndex, paramValues); result != nil {
				return
			}
		}
		for i := 1; i < len(root.children); i++ {
			for _, child := range root.children[i] {
				if result = child.find(search, paramIndex, paramValues); result != nil {
					return
				}
			}
		}
		count++
		paramIndex--
	}
	// no node found, reset params values
	for ; pindex < paramIndex; pindex++ {
		paramValues[pindex] = ""
	}
	return nil
}

func (s *node) Print(l int) {
	routes := make(map[string]string)
	for _, r := range s.routes {
		routes[r.Method()] = r.Path()
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
			if child != nil {
				child.Print(l + 2)
			}
		}
	}
}

func newNode(kind kind, prefix string, route *Route) *node {
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
		Match(string, int, bool) (int, bool)
	}
	pMatcher struct {
		ptype string
		regex *regexp.Regexp
	}
)

var (
	matchers = map[string]func(string, string) Matcher{
		"":       paramMatcher,
		"int":    paramMatcher,
		"float":  paramMatcher,
		"string": paramMatcher,
		"path":   paramMatcher,
	}
)

func (p *pMatcher) Name() string {
	return p.ptype
}

func (p *pMatcher) Match(path string, count int, next bool) (int, bool) {
	switch p.ptype {
	case "":
		return p.match(path, count, next)
	case "int":
		return p.matchInt(path, count, next)
	case "float":
		return p.matchFloat(path, count, next)
	case "string":
		return p.matchString(path, count, next)
	case "path":
		return p.matchPath(path, count, next)
	default:
		return p.matchRegex(path, count, next)
	}
}

func (p *pMatcher) match(path string, count int, next bool) (int, bool) {
	if count > 0 {
		return 0, false
	}
	end := 0
	for ; end < len(path) && path[end] != '/'; end++ {
	}
	if end == 0 || (end < len(path) && !next) {
		return 0, false
	}
	return end, true
}

func (p *pMatcher) matchInt(path string, count int, next bool) (int, bool) {
	if count > 0 {
		return 0, false
	}
	end := 0
	for ; end < len(path) && isNumeric(path[end]); end++ {
	}
	if end == 0 || (end < len(path) && !next) {
		return 0, false
	}
	return end, true
}

func (p *pMatcher) matchFloat(path string, count int, next bool) (int, bool) {
	if count > 0 {
		return 0, false
	}
	end := 0
	dot := false
	for ; end < len(path); end++ {
		if path[end] == '.' {
			if dot {
				break
			}
			dot = true
			continue
		}
		if !isNumeric(path[end]) {
			break
		}
	}
	if end == 0 || (end < len(path) && !next) {
		return 0, false
	}
	return end, true
}

// Allow one or more char, /: match /anychar but /
func (p *pMatcher) matchString(path string, count int, next bool) (int, bool) {
	if !next {
		end := 0
		for ; end < len(path) && path[end] != '/'; end++ {
		}
		return end, len(path) == end
	}
	if path[count] == '/' {
		return 0, false
	}
	return count + 1, true
}

func (p *pMatcher) matchRegex(path string, count int, next bool) (int, bool) {
	if count > 0 {
		return 0, false
	}
	if p.regex == nil {
		return 0, false
	}
	end := 0
	for ; end < len(path) && path[end] != '/'; end++ {
	}
	is := p.regex.FindStringIndex(path[:end])
	if len(is) == 0 {
		return 0, false
	}
	return is[1], true
}

// Allow empty path, such as /* match / or /anything
func (p *pMatcher) matchPath(path string, count int, next bool) (int, bool) {
	if !next {
		return len(path), true
	}
	return count, true
}

func paramMatcher(pname, ptype string) Matcher {
	return &pMatcher{ptype: ptype}
}

func regexMatcher(pname, ptype string) Matcher {
	p := &pMatcher{ptype: ptype}
	if ptype[0] != '^' {
		ptype = "^" + ptype
	}
	p.regex = regexp.MustCompile(ptype)
	return p
}

func RegisterURLParam(ptype string, matcher func(string, string) Matcher) {
	matchers[ptype] = matcher
}
