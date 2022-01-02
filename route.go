package forest

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

func isDefaultParam(m Matcher) bool {
	v := m.Param()
	return v == "" || v == "str"
}

func parseParam(param string) Matcher {
	params := strings.SplitN(param, ":", 2)
	if len(params) == 1 && params[0] == "" {
		panic("parse params error")
	}
	name := params[0]
	cate := ""
	if len(params) > 1 {
		cate = params[1]
	}
	if c, ok := converters[cate]; ok {
		return c(name, cate)
	}
	n := &npart{name: name, cate: cate}
	switch n.cate {
	case "", "int", "str", "float":
		n.kind = pkind
	case "path":
		n.kind = akind
	default:
		n.kind = rkind
		n.regex = regexp.MustCompile(n.cate)
	}
	return n
}

func parsePart(part string) (matchers []Matcher, wild bool) {
	switch part[0] {
	case ':':
		name := ""
		i := strings.Index(part, "/")
		if i > -1 {
			name = part[1:i]
		} else {
			name = part[1:]
		}
		if name == "" {
			panic("params must with name")
		}
		matchers = append(matchers, &npart{kind: pkind, name: name})
		return
	case '*':
		name := ""
		i := strings.Index(part, "/")
		if i > -1 {
			name = part[1:i]
		} else {
			name = part[1:]
		}
		if name == "" {
			name = "*"
		}
		wild = true
		matchers = append(matchers, &npart{kind: akind, name: name})
		return
	}

	for part != "" {
		i := strings.IndexFunc(part, func(r rune) bool {
			return r == '{' || (!wild && r == '/')
		})
		if i == -1 {
			matchers = append(matchers, &npart{kind: skind, cate: part})
			return
		}
		if part[i] == '/' {
			matchers = append(matchers, &npart{kind: skind, cate: part[:i]})
			return
		}

		if i > 0 {
			matchers = append(matchers, &npart{kind: skind, cate: part[:i]})
		}
		e := strings.IndexFunc(part[i:], func(r rune) bool {
			return r == '}' || r == '/'
		})
		if e == -1 {
			matchers = append(matchers, &npart{kind: skind, cate: part})
			return
		}
		if part[i+e] == '/' || e == 0 {
			matchers = append(matchers, &npart{kind: skind, cate: part[:i+e]})
			return
		}
		n := parseParam(part[i+1 : i+e])
		if n.Kind() == akind {
			wild = true
		}
		matchers = append(matchers, n)
		part = part[i+e+1:]
	}
	return
}

type node struct {
	path     string
	part     string
	children nodes
	matchers []Matcher
	isWild   bool
}

func (n *node) match(part string, ni int, wild bool, params map[string]string) bool {
	if part == "" {
		return ni == len(n.matchers)
	}
	if ni == len(n.matchers) {
		return part == "" || (!wild && part[0] == '/')
	}
	np := n.matchers[ni]
	if np.Kind() == akind {
		if ni == len(n.matchers)-1 {
			if params != nil {
				params[np.Name()] = part
			}
			return true
		}
		pi := 1
		for pi < len(part) {
			if n.match(part[pi:], ni+1, true, params) {
				if params != nil {
					params[np.Name()] = part[:pi]
				}
				return true
			}
			pi = pi + 1
		}
		return false
	}
	i, ok := np.Match(part)
	if !ok {
		return false
	}
	if np.Kind() != skind && params != nil {
		params[np.Name()] = part[:i]
	}
	return n.match(part[i:], ni+1, wild, params)
}

// func (n *node) match(part string, ni int, params map[string]string) bool {
//	pi := 0
//	ki := ni
//	for pi < len(part) {
//		if ni >= len(n.matchers) {
//			return (ki == 0 && part[pi] == '/') || pi == len(part)
//		}
//		np := n.matchers[ni]
//		if np.kind == akind {
//			if ni == len(n.matchers)-1 {
//				if params != nil {
//					params[np.name] = part[pi:]
//				}
//				return true
//			}
//			si := pi
//			for pi < len(part) {
//				if n.match(part[pi:], ni+1, params) {
//					if pi == si {
//						return false
//					}
//					if params != nil {
//						params[np.name] = part[si:pi]
//					}
//					return true
//				}
//				pi = pi + 1
//			}
//			return false
//		}
//		i, ok := np.match(part[pi:])
//		if !ok {
//			return false
//		}
//		if np.kind != skind && params != nil {
//			params[np.name] = part[pi : pi+i]
//		}
//		ni = ni + 1
//		pi = pi + i
//	}
//	return pi == len(part) && ni == len(n.matchers)
// }

func (n *node) search(path string, params map[string]string) *node {
	if n.isWild {
		return n
	}
	parts := strings.SplitN(path, "/", 2)
	if len(parts) == 1 {
		return n
	}

	part := parts[1]
	for _, child := range n.children {
		if !child.match(part, 0, false, params) {
			continue
		}
		if result := child.search(part, params); result != nil {
			return result
		}
	}
	return nil
}

func (n *node) insert(method, path string) {
	root := n
	next := path
	for {
		if len(next) > 0 && next[0] == '/' {
			next = next[1:]
		}
		if next == "" {
			root.path = path
			return
		}
		i := strings.Index(next, "/")
		if i == -1 {
			i = len(next)
		}
		part := next[:i]
		child := root.matchChild(part)
		if child == nil {
			matchers, wild := parsePart(next)
			child = &node{
				path:     path,
				part:     part,
				children: make([]*node, 0),
				matchers: matchers,
				isWild:   wild,
			}
			if wild {
				child.part = next[:i]
			}
			root.children = append(root.children, child)
			root.children.Sort()
		}
		root = child
		next = next[i:]
	}
}

func (n *node) matchChild(part string) *node {
	for _, child := range n.children {
		if child.part == part {
			return child
		}
	}
	return nil
}

func (n *node) Print(l int) {
	for _, child := range n.children {
		fmt.Print(strings.Repeat(" ", l))
		fmt.Printf("Path: %s Part: %s Kind[0]: %d\n", child.path, child.part, child.matchers[0].Kind())
		child.Print(l + 2)
	}
}

type nodes []*node

func (ns nodes) Sort() {
	sort.Sort(ns)
}

func (ns nodes) Len() int {
	return len(ns)
}

func (ns nodes) Swap(i, j int) {
	ns[i], ns[j] = ns[j], ns[i]
}

func (ns nodes) Less(i, j int) bool {
	// if ns[i].path == "" || ns[j].path == "" {
	//	return false
	// }
	// fmt.Println(ns[i].matchers[0], ns[j].matchers[0])

	// anything always last match
	if ns[i].isWild {
		return false
	}
	if ns[j].isWild {
		return true
	}

	ilen := len(ns[i].matchers)
	jlen := len(ns[j].matchers)

	// static params always first
	if ilen == 1 && ns[i].matchers[0].Kind() == skind {
		return true
	}
	if jlen == 1 && ns[j].matchers[0].Kind() == skind {
		return false
	}

	// more params match first
	// /path/{var1}-{var2} with match before /path/{var1}
	if ilen != jlen {
		return ilen > jlen
	}
	for k, v := range ns[i].matchers {
		v1 := ns[j].matchers[k]
		if v.Kind() != v1.Kind() {
			return v.Kind() < v1.Kind()
		}
		if v.Kind() == pkind {
			if isDefaultParam(v) {
				return false
			}
			if isDefaultParam(v1) {
				return true
			}
		}
	}
	return true
}
