package tests

import (
	"fmt"
	"github.com/sudachen/go-ml/fu"
	"gotest.tools/assert"
	"testing"
)

type FIX8S struct {
	f32 float32
	f8  fu.Fixed8
	s8  string
	s8x string
	err bool
}

var fix8s = []FIX8S{
	{0, fu.AsFixed8(0), "0", "0", false},
	{0, fu.AsFixed8(0), "0.00", "0", false},
	{0, fu.AsFixed8(0), "0.001", "0", false},
	{1, fu.AsFixed8(1), "1", "1", false},
	{1, fu.AsFixed8(1), "1.000", "1", false},
	{-1, fu.AsFixed8(-1), "-1", "-1", false},
	{-0.3, fu.AsFixed8(float32(-0.3)), "-0.3", "-0.3", false},
	{0.1, fu.AsFixed8(float32(0.1)), "0.1", "0.1", false},
	{0.11, fu.AsFixed8(float32(0.11)), "0.11", "0.11", false},
	{0.11, fu.AsFixed8(float32(0.11)), "0.111", "0.11", false},
	{0.11, fu.AsFixed8(float32(0.11)), "0.111", "0.11", false},
	{1.27, fu.AsFixed8(float32(1.27)), "1.271", "1.27", true},
	{0, fu.AsFixed8(float32(0)), "1.28", "0", true},
}

func (x FIX8S) String(v fu.Fixed8) string {
	return fmt.Sprintf("v: %v, f32: %v, f8: %v, s8: %v, s8x: %v", v, x.f32, x.f8, x.s8, x.s8x)
}

func Test_Fixed8strings(t *testing.T) {
	for _, x := range fix8s {
		v, err := fu.Fast8f(x.s8)
		if err != nil {
			assert.Assert(t, x.err)
		} else {
			assert.Assert(t, v == x.f8, x.String(v))
			assert.Assert(t, v.Float32() == x.f32, x.String(v))
			assert.Assert(t, fu.AsFixed8(x.f32) == v, x.String(v))
			assert.Assert(t, v.String() == x.s8x, x.String(v))
		}
	}
}
