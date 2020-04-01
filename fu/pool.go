package fu

import (
	"io"
	"runtime"
)

type AtomicPool_ struct {
	mask *AtomicMask_
	pool []interface{}
}

func AtomicPool(acquire func(no int) interface{}) *AtomicPool_ {
	a := &AtomicPool_{}
	cpunum := runtime.NumCPU()
	a.mask = ExtendableAtomicMask(
		func(no int) bool {
			// synchronized call
			if no < cpunum {
				if v := acquire(no); v != nil {
					a.pool = append(a.pool, v)
					return true
				}
			}
			return false
		})
	return a
}

func (a *AtomicPool_) Allocate() (interface{}, int) {
	n := a.mask.Lock()
	return a.pool[n], n
}

func (a *AtomicPool_) Release(n int) {
	a.mask.Unlock(n)
}

func (a *AtomicPool_) Close() {
	a.mask.FinCallForAll(func(no int) {
		if ci, ok := a.pool[no].(io.Closer); ok {
			ci.Close()
			a.pool[no] = nil // panic on reuse
		}
	})
}
