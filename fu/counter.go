package fu

import (
	"math"
	"sync"
	"sync/atomic"
)

/*
WaitCounter implements barrier counter for lazy flow execution synchronization
*/
type WaitCounter struct {
	Value uint64
	cond  sync.Cond
	mu    sync.Mutex
}

/*
Wait waits until counter Integer is not equal to specified index
*/
func (c *WaitCounter) Wait(index uint64) (r bool) {
	r = true
	if atomic.LoadUint64(&c.Value) == index {
		// mostly when executes consequentially
		return
	}
	c.mu.Lock()
	if c.cond.L == nil {
		c.cond.L = &c.mu
	}
	for c.Value != index {
		if c.Value > index {
			if c.Value == math.MaxUint64 {
				r = false
				break
			}
			panic("index continuity broken")
		}
		c.cond.Wait()
	}
	c.mu.Unlock()
	return
}

/*
PostInc increments index and notifies waiting goroutines
*/
func (c *WaitCounter) Inc() (r bool) {
	c.mu.Lock()
	if c.cond.L == nil {
		c.cond.L = &c.mu
	}
	if c.Value < math.MaxUint64 {
		atomic.AddUint64(&c.Value, 1)
		r = true
	}
	c.mu.Unlock()
	c.cond.Broadcast()
	return
}

/*
Stop sets Integer to ~uint64(0) and notifies waiting goroutines. It means also counter will not increment more
*/
func (c *WaitCounter) Stop() {
	c.mu.Lock()
	if c.cond.L == nil {
		c.cond.L = &c.mu
	}
	atomic.StoreUint64(&c.Value, math.MaxUint64)
	c.mu.Unlock()
	c.cond.Broadcast()
}

/*
Stopped returns true if counter is stopped and will not increment more
*/
func (c *WaitCounter) Stopped() bool {
	return atomic.LoadUint64(&c.Value) == math.MaxUint64
}
