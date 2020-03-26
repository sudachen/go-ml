package tables

import (
	"fmt"
	"github.com/sudachen/go-foo/fu"
	"github.com/sudachen/go-ml/mlutil"
	"strings"
)

func (t *Table) Head(n int) string {
	if n < t.raw.Length {
		return t.Display(0, n)
	}
	return t.Display(0, t.raw.Length)
}

func (t *Table) Tail(n int) string {
	if n < t.raw.Length {
		return t.Display(t.raw.Length-n, t.raw.Length)
	}
	return t.Display(0, t.raw.Length)
}

func (t *Table) String() string {
	return t.Display(0, t.raw.Length)
}

func (t *Table) Display(from, to int) string {
	n := fu.Mini(to-from, t.raw.Length-from)
	if n < 0 {
		n = 0
	}
	w := make([]int, len(t.raw.Names)+1)
	s := make([][]interface{}, n+1)
	s[0] = append(make([]interface{}, len(w)))
	s[0][0] = ""
	for i, n := range t.raw.Names {
		s[0][i+1] = n
		w[i+1] = len(n)
	}
	for k := 0; k < n; k++ {
		u := make([]interface{}, len(w))
		ln := fmt.Sprint(k + from)
		if w[0] < len(ln) {
			w[0] = len(ln)
		}
		u[0] = ln
		for j := range w[1:] {
			ws := mlutil.Cell{t.raw.Columns[j].Index(k + from)}.String()
			if len(ws) > w[j+1] {
				w[j+1] = len(ws)
			}
			u[j+1] = ws
		}
		s[k+1] = u
	}
	f0 := ""
	f1 := ""
	f2 := ""
	for i, v := range w {
		if i != 0 {
			f0 += " . "
			f1 += " | "
			f2 += "-|-"
		}
		q := fmt.Sprintf("%%-%ds", v)
		f0 += q
		f1 += q
		f2 += strings.Repeat("-", v)
	}
	r := ""
	r += fmt.Sprintf(f0, s[0]...) + "\n"
	r += f2 + "\n"
	for _, u := range s[1:] {
		r += fmt.Sprintf(f1, u...) + "\n"
	}
	return r[:len(r)-1]
}
