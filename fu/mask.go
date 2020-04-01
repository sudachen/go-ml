package fu

import (
	"runtime"
	"sync"
)

type AtomicMask_ struct {
	width      int
	value      uint64
	mu         sync.Mutex
	cond       sync.Cond
	extendable bool
	probe      func(int) bool
}

func AtomicMask(width int) *AtomicMask_ {
	if width > 64 {
		width = 64
	}
	a := &AtomicMask_{width: width}
	a.cond.L = &a.mu
	return a
}

func ExtendableAtomicMask(canextend func(int) bool) *AtomicMask_ {
	a := &AtomicMask_{width: 0}
	a.cond.L = &a.mu
	a.extendable = true
	a.probe = canextend
	return a
}

var numcpu = runtime.NumCPU()

func (a *AtomicMask_) Lock() int {
	n := -1
	a.mu.Lock()
l:
	for n == -1 {
		x := ^uint64(0) >> (64 - a.width)
		if x != 0 && a.value&x == 0 {
			n = 0
			for a.value&(uint64(1)<<n) != 0 {
				n++
			}
			a.value |= (uint64(1) << n)
			break l
		}

		if a.extendable {
			if a.width < 64 && a.probe(a.width) {
				n = a.width
				a.value |= (uint64(1) << n)
				a.width++
				break l
			}
			a.extendable = false
		}

		a.cond.Wait()
	}
	a.mu.Unlock()
	return n
}

func (a *AtomicMask_) Unlock(i int) {
	a.mu.Lock()
	if a.value&(uint64(1)<<i) != 0 {
		a.value = a.value &^ (uint64(1) << i)
		a.mu.Unlock()
		a.cond.Broadcast()
		return
	}
	a.mu.Unlock() // ?
	panic("opps")
}

func (a *AtomicMask_) FinCallForAll(f func(no int)) {
	a.mu.Lock()
	a.extendable = false
	mask := ^uint64(0) >> (64 - a.width)
	for mask != 0 {
		for i := 0; i < a.width; i++ {
			x := uint64(1) << i
			if mask&x != 0 && a.value&x == 0 {
				mask &= ^x
				a.value |= x
				f(i)
			}
		}
		if mask != 0 {
			a.cond.Wait()
		}
	}
	a.mu.Unlock()
}
