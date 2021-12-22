package forest

import (
	"regexp"
	"strings"
	"unicode"
)

type (
	Matcher interface {
		Kind() uint8
		Match(string) (int, bool)
		Name() string
		Param() string
	}
	Converter func(string, string) Matcher
)

var converters map[string]Converter

type npart struct {
	kind  uint8
	name  string
	cate  string
	regex *regexp.Regexp
}

const (
	skind uint8 = iota // static path
	pkind              // path with params
	rkind              // path with regexp
	akind = 10         // path with anything
)

func (n *npart) Name() string {
	return n.name
}

func (n *npart) Param() string {
	return n.cate
}

func (n *npart) Kind() uint8 {
	return n.kind
}

func (n *npart) Match(part string) (int, bool) {
	switch n.kind {
	case skind:
		if len(part) < len(n.cate) || part[:len(n.cate)] != n.cate {
			return 0, false
		}
		return len(n.cate), true
	case rkind:
		i := strings.Index(part, "/")
		if i == -1 {
			i = len(part)
		}
		is := n.regex.FindStringIndex(part[:i])
		if len(is) == 0 || is[0] > 0 {
			return 0, false
		}
		return is[1], true
	case pkind:
		switch n.cate {
		case "int":
			i := strings.IndexFunc(part, func(r rune) bool {
				return !unicode.IsDigit(r)
			})
			if i == 0 {
				return 0, false
			}
			if i == -1 {
				i = len(part)
			}
			return i, true
		case "float":
			dot := false
			i := strings.IndexFunc(part, func(r rune) bool {
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
				return 0, false
			}
			if i == -1 {
				i = len(part)
			}
			return i, true
		case "", "str":
			i := strings.IndexFunc(part, func(r rune) bool {
				return r == '/'
			})
			if i == 0 {
				return 0, false
			}
			if i == -1 {
				i = len(part)
			}
			return i, true
		}
	case akind:
		return len(part), true
	}
	return 0, false
}
