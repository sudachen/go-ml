package mlutil

import (
	"fmt"
	"github.com/sudachen/go-foo/lazy"
	"strings"
)

/*
	123textE + *textE -> *TEXT => 123TEXT
	123textE + 123text* -> TEXT* => TEXTE
	123textE + *text* -> *TEXT* => 123TEXTE
	123textE + 123*E -> 55*5 => 55text5

*/

func makesubst(subst string) func(string) string {
	counter := lazy.AtomicCounter{1}
	j := strings.Index(subst, "*")
	if j < 0 {
		return func(v string) string {
			if len(v) == 0 {
				return subst
			}
			return subst + fmt.Sprint(counter.PostInc())
		}
	}
	if j == 0 {
		return func(v string) string { return v + subst[1:] }
	}
	if j == len(subst)-1 {
		return func(v string) string { return subst[:j] + v }
	}
	return func(v string) string {
		return subst[:j] + v + subst[j+1:]
	}
}

func Starsub(pattern, subst string) func(string) (string, bool) {
	j := strings.Index(pattern, "*")
	if j < 0 {
		return func(v string) (string, bool) {
			if v == pattern {
				return subst, true
			}
			return pattern, false
		}
	}
	if j == 0 {
		right := pattern[1:]
		if right[len(right)-1] == '*' {
			center := right[:len(right)-1]
			if subst[0] != '*' && subst[len(subst)-1] != '*' {
				panic("substitution must be like *blablabla*")
			}
			subst = subst[1 : len(subst)-1]
			return func(v string) (string, bool) {
				if k := strings.Index(v, center); k > 0 {
					return v[:k] + subst + v[k+len(center):], true
				}
				return v, false
			}
		}
		f := makesubst(subst)
		return func(v string) (string, bool) {
			if strings.HasSuffix(v, right) {
				return f(v[:len(v)-len(right)]), true
			}
			return v, false
		}
	}
	if j == len(pattern)-1 {
		left := pattern[:j]
		f := makesubst(subst)
		return func(v string) (string, bool) {
			if strings.HasPrefix(v, left) {
				return f(v[len(left):]), true
			}
			return v, false
		}
	}
	// 	123textE + 123*E -> 55*5 => 55text5
	left := pattern[:j]
	right := pattern[j+1:]
	f := makesubst(subst)
	return func(v string) (string, bool) {
		if len(v) > len(left)+len(right) && strings.HasPrefix(v, left) && strings.HasSuffix(v, right) {
			return f(v[len(left) : len(v)-len(right)]), true
		}
		return v, false
	}
}

func Pattern(pattern string) func(string) bool {
	l := len(pattern)
	j := strings.Index(pattern, "*")
	if j < 0 {
		return func(name string) bool { return name == pattern }
	}
	if j == 0 {
		p := pattern[1:]
		if p[len(p)-1] == '*' {
			p := p[:len(p)-1]
			return func(v string) bool {
				k := strings.Index(v, p)
				return k > 0 && k < len(v)-1
			}
		}
		return func(name string) bool {
			return strings.HasSuffix(name, p)
		}
	}
	if j == l-1 {
		p := pattern[:j]
		return func(name string) bool {
			return strings.HasPrefix(name, p)
		}
	}
	left := pattern[:j]
	right := pattern[j+1:]
	return func(name string) bool {
		return len(name) >= l &&
			strings.HasPrefix(name, left) &&
			strings.HasSuffix(name, right)
	}
}
