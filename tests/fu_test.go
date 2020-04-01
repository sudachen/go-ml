package tests

import (
	"github.com/sudachen/go-ml/fu"
	"gotest.tools/assert"
	"gotest.tools/assert/cmp"
	"reflect"
	"testing"
)

func Test_Less1(t *testing.T) {
	a := map[int]interface{}{0: 0}
	assert.Assert(t, cmp.Panics(func() {
		fu.Less(reflect.ValueOf(a), reflect.ValueOf(a))
	}))
	assert.Assert(t, cmp.Panics(func() {
		fu.Less(reflect.ValueOf(1), reflect.ValueOf(""))
	}))
	assert.Assert(t, fu.Less(reflect.ValueOf([2]int{0, 1}), reflect.ValueOf([2]int{0, 2})))
	assert.Assert(t, cmp.Panics(func() {
		fu.Less(reflect.ValueOf([2]int{0, 1}), reflect.ValueOf([1]int{0}))
	}))
}

func Test_MinMax(t *testing.T) {
	assert.Assert(t, fu.MinIndex(reflect.ValueOf([]int{1, 2, 3, 4, 5})) == 0)
	assert.Assert(t, fu.MaxIndex(reflect.ValueOf([]int{1, 2, 3, 4, 5})) == 4)
	assert.Assert(t, fu.Min(1, 2, 3, 4, 5).(int) == 1)
	assert.Assert(t, fu.Max(1, 2, 3, 4, 5).(int) == 5)
	assert.Assert(t, fu.MinValue(reflect.ValueOf([]int{1, 2, 3, 4, 5})).Int() == 1)
	assert.Assert(t, fu.MaxValue(reflect.ValueOf([]int{1, 2, 3, 4, 5})).Int() == 5)
}
