package tests

import (
	"fmt"
	"github.com/sudachen/go-ml/fu"
	"gotest.tools/assert"
	"testing"
)

type STR struct {
	text, result string
	ok           bool
}

type STRr struct {
	ref STR
	r   string
	ok  bool
}

func (str STR) Apply(f func(string) (string, bool)) STRr {
	r, ok := f(str.text)
	return STRr{str, r, ok}
}

func (r STRr) String() string {
	return fmt.Sprintf("x=%v; f(%v) -> %v, %v", r.ref, r.ref.text, r.r, r.ok)
}

func (r STRr) Ok() bool {
	return r.r == r.ref.result && r.ok == r.ref.ok
}

func (str STR) Assert(t *testing.T, f func(string) (string, bool)) {
	r := str.Apply(f)
	assert.Assert(t, r.Ok(), r)
}

func Test_Subst1(t *testing.T) {
	f := fu.Starsub("test*", "F*")
	for _, x := range []STR{
		{"test1", "F1", true},
		{"tes1", "tes1", false},
		{"test2", "F2", true},
		{"test", "F", true},
	} {
		x.Assert(t, f)
	}
}

func Test_Subst3(t *testing.T) {
	f := fu.Starsub("*test", "F*")
	for _, x := range []STR{
		{"123test", "F123", true},
		{"tes1", "tes1", false},
		{"2test", "F2", true},
		{"test", "F", true},
	} {
		x.Assert(t, f)
	}
}

func Test_Subst4(t *testing.T) {
	f := fu.Starsub("test*", "F")
	for _, x := range []STR{
		{"testXX", "F1", true},
		{"testYY", "F2", true},
		{"test", "F", true},
		{"testZZ", "F3", true},
	} {
		x.Assert(t, f)
	}
}

func Test_Subst5(t *testing.T) {
	f := fu.Starsub("test*", "F*i")
	for _, x := range []STR{
		{"testXX", "FXXi", true},
		{"testYY", "FYYi", true},
		{"test", "Fi", true},
		{"tes", "tes", false},
	} {
		x.Assert(t, f)
	}
}
