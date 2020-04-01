package fu

import (
	"math"
	"os"
	"sync/atomic"
	"time"
)

type NaiveRandom struct {
	Value uint32
}

func (nr *NaiveRandom) Reseed() {
	atomic.StoreUint32(&nr.Value, uint32(time.Now().UnixNano()+int64(os.Getpid())))
}

func (nr *NaiveRandom) Uint() uint {
	var r uint32
	for {
		r = atomic.LoadUint32(&nr.Value)
		rx := r
		r = r*1664525 + 1013904223
		if atomic.CompareAndSwapUint32(&nr.Value, rx, r) {
			break
		}
	}
	return uint(r)
}

func (nr *NaiveRandom) Float() float64 {
	return float64(nr.Uint()) / math.MaxUint32
}
