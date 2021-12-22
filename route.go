package forest

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

func parseParam(param string) Matcher {
	params := strings.SplitN(param, ":", 2)
	if len(params) == 1 && params[0] == "" {
		panic("path error")
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
	nstr := ""
	pi := 0
	for pi < len(part) {
		if !wild && part[pi] == '/' {
			break
		}
		if part[pi] == '{' {
			i := strings.IndexFunc(part[pi:], func(r rune) bool {
				return (!wild && r == '/') || r == '}'
			})
			if i > 1 && part[pi+i] == '}' {
				if nstr != "" {
					matchers = append(matchers, &npart{kind: skind, cate: nstr})
					nstr = ""
				}
				n := parseParam(part[pi+1 : pi+i])
				if n.Kind() == akind {
					wild = true
				}
				matchers = append(matchers, n)
				pi = pi + i + 1
				continue
			}
		}
		nstr = nstr + part[pi:pi+1]
		pi = pi + 1
	}
	if nstr != "" {
		matchers = append(matchers, &npart{kind: skind, cate: nstr})
	}
	return matchers, wild
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

func (n *node) insert(path string, pi int) {
	ni := 0
	for {
		if pi >= len(path)-1 {
			n.path = path
			return
		}
		ni = strings.Index(path[pi:], "/")
		if ni == -1 {
			ni = len(path) - pi
			break
		}
		if ni > 0 {
			break
		}
		pi = pi + 1
	}

	part := path[pi : pi+ni]
	child := n.matchChild(part)

	if child == nil {
		matchers, wild := parsePart(path[pi:])
		if wild {
			ni = len(path) - pi
			part = path[pi : pi+ni]
		}

		child = &node{part: part, matchers: matchers, children: make([]*node, 0), isWild: wild}
		n.children = append(n.children, child)
		n.children.Sort()
	}
	child.insert(path, pi+ni)
}

func (n *node) travel(list *([]*node)) {
	if n.path != "" {
		*list = append(*list, n)
	}
	for _, child := range n.children {
		child.travel(list)
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
	if ns[i].path == "" || ns[j].path == "" {
		return false
	}
	// anything always last match
	if ns[i].isWild {
		return false
	}
	if ns[j].isWild {
		return true
	}
	// more params match first
	// /path/{var1}-{var2} with match before /path/{var1}
	ilen := len(ns[i].matchers)
	jlen := len(ns[j].matchers)
	if ilen != jlen {
		return ilen > jlen
	}
	for k, v := range ns[i].matchers {
		v1 := ns[j].matchers[k]
		if v.Kind() != v1.Kind() {
			return v.Kind() < v1.Kind()
		}
		if v.Kind() == pkind {
			if v.Param() == "" {
				return false
			}
			if v1.Param() == "" {
				return true
			}
		}
	}
	return true
}
