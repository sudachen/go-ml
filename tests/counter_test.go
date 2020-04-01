package tests

import (
	"github.com/sudachen/go-ml/fu"
	"gotest.tools/assert"
	"gotest.tools/assert/cmp"
	"sync"
	"testing"
)

func Test_Counter1(t *testing.T) {
	assert.Assert(t, cmp.Panics(func() {
		wc := fu.WaitCounter{Value: 10}
		wc.Wait(1)
	}))
	wc := fu.WaitCounter{Value: 0}
	wc.Inc()
	wc.Wait(1)
	wc.Inc()
	assert.Assert(t, wc.Value == 2)
}

func Test_Counter2(t *testing.T) {
	assert.Assert(t, cmp.Panics(func() {
		wc := fu.WaitCounter{Value: 10}
		wc.Wait(1)
	}))

	wc1 := fu.WaitCounter{Value: 0}
	wc2 := fu.WaitCounter{Value: 0}
	wc3 := fu.WaitCounter{Value: 0}

	N := 10
	x := make([]int, N)
	wg := sync.WaitGroup{}

	for i := 0; i < N; i++ {
		wg.Add(1)
		go func(i int) {
			if wc1.Wait(uint64(i)) {
				x[i] = i + 1
				wc1.Inc()
			}
			if wc2.Wait(uint64(i + 1)) {
				x[i] = N - i
				wc2.Inc()
			}
			if wc3.Wait(uint64(i + 1)) {
				x[i] = 0
				wc3.Inc()
			}
			wg.Done()
		}(i)
	}

	wc1.Wait(uint64(N))
	for i := 0; i < N; i++ {
		assert.Assert(t, x[i] == i+1)
	}
	assert.Assert(t, wc2.Inc() == true)
	assert.Assert(t, wc2.Stopped() == false)
	wc2.Wait(uint64(N + 1))
	for i := 0; i < N; i++ {
		assert.Assert(t, x[i] == N-i)
	}
	wc3.Stop()
	assert.Assert(t, wc3.Inc() == false)
	assert.Assert(t, wc3.Stopped() == true)
	wg.Wait()
	for i := 0; i < N; i++ {
		assert.Assert(t, x[i] == N-i)
	}
}
