package fu

import (
	"sync/atomic"
)

/*
AtomicCounter - hm, yes it's atomic counter
*/
type AtomicCounter struct {
	Value uint64
}

/*
PostInc increments counter and returns OLD value
*/
func (c *AtomicCounter) PostInc() uint64 {
	for {
		v := atomic.LoadUint64(&c.Value)
		if atomic.CompareAndSwapUint64(&c.Value, v, v+1) {
			return v
		}
	}
}

/*
Dec decrements counter and returns NEW value
*/
func (c *AtomicCounter) Dec() uint64 {
	for {
		v := atomic.LoadUint64(&c.Value)
		if v == 0 {
			panic("counter underflow")
		}
		if atomic.CompareAndSwapUint64(&c.Value, v, v-1) {
			return v - 1
		}
	}
}

/*
AtomicFlag - hm, yes it's atomic flag
*/
type AtomicFlag struct {
	Value int32
}

/*
Clear switches Integer to 0 atomically
*/
func (c *AtomicFlag) Clear() bool {
	return atomic.CompareAndSwapInt32(&c.Value, 1, 0)
}

/*
Set switches Integer to 1 atomically
*/
func (c *AtomicFlag) Set() bool {
	return atomic.CompareAndSwapInt32(&c.Value, 0, 1)
}

/*
State returns current state
*/
func (c *AtomicFlag) State() bool {
	v := atomic.LoadInt32(&c.Value)
	return bool(v != 0)
}

/*
AtomicSingleIndex - it's an atomic positive integer single set value
*/
type AtomicSingleIndex struct {v uint64}

/*
Get returns index value
*/
func (c *AtomicSingleIndex) Get() (int, bool) {
	v := atomic.LoadUint64(&c.v)
	return int(v-1), v != 0
}

/*
Set sets value if it was not set before
*/
func (c *AtomicSingleIndex) Set(value int) (int,bool) {
	if  atomic.CompareAndSwapUint64(&c.v,0, uint64(value)+1) {
		return value, true
	}
	return int(atomic.LoadUint64(&c.v)-1), false
}
