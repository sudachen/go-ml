package tests

import (
	"fmt"
	"github.com/sudachen/go-ml/lazy"
	"gotest.tools/assert"
	"reflect"
	"testing"
)

type Color struct {
	Color string
	Index int
}

var colors = []Color{
	{"White", 0},
	{"Yellow", 1},
	{"Blue", 2},
	{"Red", 3},
	{"Green", 4},
	{"Black", 5},
	{"Brown", 6},
	{"Azure", 7},
	{"Ivory", 8},
	{"Teal", 9},
	{"Silver", 10},
	{"Purple", 11},
	{"Navy blue", 12},
	{"Pea green", 13},
	{"Gray", 14},
	{"Orange", 15},
	{"Maroon", 16},
	{"Charcoal", 17},
	{"Aquamarine", 18},
	{"Coral", 19},
	{"Fuchsia", 20},
	{"Wheat", 21},
	{"Lime", 22},
	{"Crimson", 23},
	{"Khaki", 24},
	{"Hot pink", 25},
	{"Magenta", 26},
	{"Olden", 27},
	{"Plum", 28},
	{"Olive", 29},
	{"Cyan", 30},
}

func Test_NewFromChan(t *testing.T) {
	c := make(chan Color)
	go func() {
		for _, x := range colors {
			c <- x
		}
		close(c)
	}()
	rs := lazy.Chan(c).LuckyCollect().([]Color)
	assert.DeepEqual(t, rs, colors)
}

func Test_Collect(t *testing.T) {
	rs := lazy.List(colors).LuckyCollect().([]Color)
	assert.DeepEqual(t, rs, colors)
}

func Test_ConqCollect(t *testing.T) {
	z := lazy.List(colors)
	rs := z.Parallel(8).LuckyCollect().([]Color)
	assert.DeepEqual(t, rs, colors)
	rs = z.Parallel().LuckyCollect().([]Color)
	assert.DeepEqual(t, rs, colors)
}

func Test_Filter(t *testing.T) {
	z := lazy.List(colors).
		Filter(func(c Color) bool { return c.Index%2 == 0 }).
		Parallel()
	rs := z.LuckyCollect().([]Color)
	for _, c := range rs {
		assert.Assert(t, c.Index%2 == 0)
	}
	for _, c := range colors {
		if c.Index%2 == 0 {
			assert.Assert(t, rs[c.Index/2].Index == c.Index)
		}
	}
	// again
	rs = z.LuckyCollect().([]Color)
	for _, c := range rs {
		assert.Assert(t, c.Index%2 == 0)
	}
	for _, c := range colors {
		if c.Index%2 == 0 {
			assert.Assert(t, rs[c.Index/2].Index == c.Index)
		}
	}
}

func Test_Map1(t *testing.T) {
	rs := lazy.List([]int{0, 1, 2, 3, 4}).
		Map(func(r int) string { return fmt.Sprint(r) }).
		Parallel().
		LuckyCollect().([]string)
	assert.Assert(t, len(rs) == 5)
	for i, r := range rs {
		assert.Assert(t, r == fmt.Sprint(i))
	}
}

func Test_Map2(t *testing.T) {
	rs := lazy.List(colors).
		Map(func(r Color) string { return r.Color }).
		Parallel(6).
		LuckyCollect().([]string)
	assert.Assert(t, len(rs) == len(colors))
	for i, r := range rs {
		assert.Assert(t, r == colors[i].Color)
	}
}

func Test_Map3(t *testing.T) {
	type R struct{ c string }
	z := lazy.List(colors)
	rs := z.Map(func(r Color) R { return R{r.Color} }).LuckyCollect().([]R)
	assert.Assert(t, len(rs) == len(colors))
	for i, r := range rs {
		assert.Assert(t, r.c == colors[i].Color)
	}
}

func Test_Sink(t *testing.T) {
	type R struct{ c string }
	rs := []R{}
	z := lazy.List(colors)
	z.Map(func(r Color) R { return R{r.Color} }).Drain(func(value reflect.Value) error {
		if value.Kind() != reflect.Bool {
			rs = append(rs, value.Interface().(R))
		}
		return nil
	})
	assert.Assert(t, len(rs) == len(colors))
	for i, r := range rs {
		assert.Assert(t, r.c == colors[i].Color)
	}
}

func Test_SinkClose(t *testing.T) {
	type R struct{ c string }
	closed := false
	rs := []R{}
	z := lazy.List(colors)
	z.Map(func(r Color) R { return R{r.Color} }).Drain(func(value reflect.Value) error {
		if closed {
			panic("already closed")
		}
		if value.Kind() != reflect.Bool {
			rs = append(rs, value.Interface().(R))
		} else {
			closed = true
		}
		return nil
	})
	assert.Assert(t, len(rs) == len(colors))
	for i, r := range rs {
		assert.Assert(t, r.c == colors[i].Color)
	}
	assert.Assert(t, closed)
}

func Test_SinkErrClose(t *testing.T) {
	type R struct{ c string }
	closed := false
	z := lazy.List(colors)
	z.Map(func(r Color) R { return R{r.Color} }).Drain(func(value reflect.Value) error {
		if closed {
			panic("already closed")
		}
		if value.Kind() != reflect.Bool {
			// nothing
		} else {
			closed = true
		}
		return fmt.Errorf("error")
	})
	assert.Assert(t, closed)
}

func Test_SinkErrClose2(t *testing.T) {
	type R struct{ c string }
	closed1 := false
	closed2 := false
	z := lazy.Source(func() lazy.Stream {
		return func(index uint64) (reflect.Value, error) {
			if index == lazy.STOP {
				closed1 = true
			}
			return reflect.Value{}, fmt.Errorf("error")
		}
	})
	z.Map(func(r Color) R { return R{r.Color} }).Drain(func(value reflect.Value) error {
		if closed2 {
			panic("already closed")
		}
		if value.Kind() != reflect.Bool {
			// nothing
		} else {
			closed2 = true
		}
		return fmt.Errorf("error")
	})
	assert.Assert(t, closed1)
	assert.Assert(t, closed2)
}
