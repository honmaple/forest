package forest

import (
	"fmt"
	"regexp"
	"strings"
)

type (
	kind uint8
	node struct {
		kind     kind
		prefix   string
		routes   Routes
		matcher  Matcher
		optional bool
		children [akind + 1]nodes
		hasChild bool
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

func (n *node) insertStatic(path string, route *Route) *node {
	root := n
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

func (n *node) insertParam(rule string, optional bool, route *Route) *node {
	var (
		label byte = ':'
		nkind kind = pkind
	)

	if rule == "path" {
		label = '*'
		nkind = akind
		optional = true
	}

	if optional {
		n.addRoute(route)
	}

	child := n.findParamChild(nkind, rule, optional)
	if child == nil {
		child = newNode(nkind, string(label), route)
		child.matcher = newMatcher(rule)
		child.optional = optional
		n.addChild(child)
	} else {
		child.addRoute(route)
	}
	return child
}

func (n *node) insert(path string, route *Route) {
	root := n

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
			rule := "string"
			if len(params) > 1 {
				rule = params[1]
			}

			route.pnames = append(route.pnames, routeParam{start: start, end: e + 1, name: pname})
			if e == l-1 {
				root = root.insertParam(rule, optional, route)
			} else {
				root = root.insertParam(rule, optional, nil)
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
				root = root.insertParam("", optional, route)
			} else {
				root = root.insertParam("", optional, nil)
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
				root = root.insertParam("path", true, route)
			} else {
				root = root.insertParam("path", true, nil)
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

func (n *node) addRoute(route *Route) {
	if route == nil {
		return
	}
	n.routes = append(n.routes, route)
}

func (n *node) addChild(child *node) {
	if child.kind == skind {
		if len(n.children[skind]) == 0 {
			n.children[skind] = make(nodes, 256)
		}
		n.children[skind][child.prefix[0]] = child
	} else {
		n.children[child.kind] = append(n.children[child.kind], child)
	}
	n.hasChild = true
}

func (n *node) findParamChild(kind kind, rule string, optional bool) *node {
	for _, child := range n.children[kind] {
		if child.optional == optional && (child.matcher == nil || child.matcher.Name() == rule) {
			return child
		}
	}
	return nil
}

func (n *node) findStaticChild(l byte) *node {
	if v := n.children[skind]; len(v) == 0 {
		return nil
	} else {
		return v[l]
	}
}

func (n *node) find(path string, paramIndex int, paramValues []string) (result *node) {
	var (
		root          = n
		index         = 0
		pindex        = paramIndex
		checkOptional = false
	)

LOOP:
	for {
		e, ok := 0, false
		switch root.kind {
		case skind:
			if index > 0 {
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
			index++
		default:
			// if optional, match child first
			if root.optional && !checkOptional {
				checkOptional = true
			} else {
				e, ok = root.matcher.Match(path, index, root.hasChild)
				if !ok {
					break LOOP
				}
				if e == 0 {
					index++
				} else {
					index = e
					paramValues[paramIndex] = path[:e]
				}
			}
			paramIndex++
		}
		if len(path) == e {
			return root
		}
		if root.hasChild {
			search := path[e:]
			if child := root.findStaticChild(search[0]); child != nil {
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
		}
		if root.kind > skind {
			paramIndex--
		}
	}
	// no node found, reset params values
	for ; pindex < paramIndex; pindex++ {
		paramValues[pindex] = ""
	}
	return nil
}

func (n *node) Print(l int) {
	routes := make(map[string]string)
	for _, r := range n.routes {
		routes[r.Method()] = r.Path()
	}
	fmt.Print(strings.Repeat(" ", l))
	fmt.Printf("%s", n.prefix)
	if len(routes) > 0 {
		fmt.Printf(" %+v\n", routes)
	} else {
		fmt.Print(" nil \n")
	}
	for i := range n.children {
		for _, child := range n.children[i] {
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
		// path, index, next, index is last matched index
		Match(string, int, bool) (int, bool)
	}
	ruleMatcher struct {
		rule  string
		regex *regexp.Regexp
	}
)

var (
	matchers = map[string]func(string) Matcher{
		"":       newRuleMatcher,
		"int":    newRuleMatcher,
		"float":  newRuleMatcher,
		"string": newRuleMatcher,
		"path":   newRuleMatcher,
	}
)

func (r *ruleMatcher) Name() string {
	return r.rule
}

func (r *ruleMatcher) Match(path string, index int, next bool) (int, bool) {
	switch r.rule {
	case "":
		return r.match(path, index, next)
	case "int":
		return r.matchInt(path, index, next)
	case "float":
		return r.matchFloat(path, index, next)
	case "string":
		return r.matchString(path, index, next)
	case "path":
		return r.matchPath(path, index, next)
	default:
		return r.matchRegex(path, index, next)
	}
}

func (r *ruleMatcher) match(path string, index int, next bool) (int, bool) {
	if index > 0 {
		return 0, false
	}
	for ; index < len(path) && path[index] != '/'; index++ {
	}
	if index == 0 || (index < len(path) && !next) {
		return 0, false
	}
	return index, true
}

func (r *ruleMatcher) matchInt(path string, index int, next bool) (int, bool) {
	if index > 0 {
		return 0, false
	}
	for ; index < len(path) && isNumeric(path[index]); index++ {
	}
	if index == 0 || (index < len(path) && !next) {
		return 0, false
	}
	return index, true
}

func (r *ruleMatcher) matchFloat(path string, index int, next bool) (int, bool) {
	if index > 0 {
		return 0, false
	}
	dot := false
	for ; index < len(path); index++ {
		if path[index] == '.' {
			if dot {
				break
			}
			dot = true
			continue
		}
		if !isNumeric(path[index]) {
			break
		}
	}
	if index == 0 || (index < len(path) && !next) {
		return 0, false
	}
	return index, true
}

func (r *ruleMatcher) matchRegex(path string, index int, next bool) (int, bool) {
	if index > 0 {
		return 0, false
	}
	if r.regex == nil {
		return 0, false
	}
	for ; index < len(path) && path[index] != '/'; index++ {
	}
	is := r.regex.FindStringIndex(path[:index])
	if len(is) == 0 {
		return 0, false
	}
	return is[1], true
}

// Allow one or more char, /: match /anychar but /
func (r *ruleMatcher) matchString(path string, index int, next bool) (int, bool) {
	if !next {
		if index > 0 {
			return 0, false
		}
		for ; index < len(path) && path[index] != '/'; index++ {
		}
		return index, len(path) == index
	}
	if path[index] == '/' {
		return 0, false
	}
	return index + 1, true
}

// Allow empty path, such as /* match / or /anything
func (r *ruleMatcher) matchPath(path string, index int, next bool) (int, bool) {
	if !next {
		if index > 0 {
			return 0, false
		}
		return len(path), true
	}
	return index + 1, true
}

func newMatcher(rule string) Matcher {
	if mc, ok := matchers[rule]; ok {
		return mc(rule)
	}
	return newRegexMatcher(rule)
}

func newRuleMatcher(rule string) Matcher {
	return &ruleMatcher{rule: rule}
}

func newRegexMatcher(rule string) Matcher {
	r := &ruleMatcher{rule: rule}
	if rule[0] != '^' {
		rule = "^" + rule
	}
	r.regex = regexp.MustCompile(rule)
	return r
}

func RegisterRule(rule string, matcher func(string) Matcher) {
	matchers[rule] = matcher
}
