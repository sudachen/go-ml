package lazy

import (
	"github.com/sudachen/go-ml/fu"
	"github.com/sudachen/go-zorros/zorros"
	"math"
	"reflect"
	"runtime"
	"sync/atomic"
	"unsafe"
)

const STOP = math.MaxUint64

type Stream func(index uint64) (reflect.Value, error)
type Source func() Stream
type Sink func(reflect.Value) error
type Parallel int

var falseValue = reflect.ValueOf(false)
var trueValue = reflect.ValueOf(true)

func (zf Source) Map(f interface{}) Source {
	return func() Stream {
		z := zf()
		return func(index uint64) (v reflect.Value, err error) {
			if v, err = z(index); err != nil || v.Kind() == reflect.Bool {
				return v, err
			}
			fv := reflect.ValueOf(f)
			return fv.Call([]reflect.Value{v})[0], nil
		}
	}
}

func (zf Source) Filter(f interface{}) Source {
	return func() Stream {
		z := zf()
		return func(index uint64) (v reflect.Value, err error) {
			if v, err = z(index); err != nil || v.Kind() == reflect.Bool {
				return v, err
			}
			fv := reflect.ValueOf(f)
			if fv.Call([]reflect.Value{v})[0].Bool() {
				return
			}
			return trueValue, nil
		}
	}
}

func (zf Source) Parallel(concurrency ...int) Source {
	return func() Stream {
		z := zf()
		ccrn := fu.Fnzi(fu.Fnzi(concurrency...), runtime.NumCPU())
		type C struct {
			reflect.Value
			error
		}
		index := fu.AtomicCounter{0}
		wc := fu.WaitCounter{Value: 0}
		c := make(chan C)
		stop := make(chan struct{})
		alive := fu.AtomicCounter{uint64(ccrn)}
		for i := 0; i < ccrn; i++ {
			go func() {
			loop:
				for !wc.Stopped() {
					n := index.PostInc() // returns old value
					v, err := z(n)
					if n < STOP && wc.Wait(n) {
						select {
						case c <- C{v, err}:
						case <-stop:
							wc.Stop()
							break loop
						}
						wc.Inc()
					}
				}
				if alive.Dec() == 0 { // return new value
					close(c)
				}
			}()
		}
		return func(index uint64) (reflect.Value, error) {
			if index == STOP {
				close(stop)
				return z(STOP)
			}
			if x, ok := <-c; ok {
				return x.Value, x.error
			}
			return falseValue, nil
		}
	}
}

func (zf Source) First(n int) Source {
	return func() Stream {
		z := zf()
		count := 0
		wc := fu.WaitCounter{Value: 0}
		return func(index uint64) (v reflect.Value, err error) {
			v, err = z(index)
			if index != STOP && wc.Wait(index) {
				if count < n && err == nil {
					if v.Kind() != reflect.Bool {
						count++
					}
					wc.Inc()
					return
				}
				wc.Stop()
			}
			return falseValue, nil
		}
	}
}

func (zf Source) Drain(sink func(reflect.Value) error) (err error) {
	z := zf()
	var v reflect.Value
	var i uint64
	for {
		if v, err = z(i); err != nil {
			break
		}
		i++
		if v.Kind() != reflect.Bool {
			if err = sink(v); err != nil {
				break
			}
		} else if !v.Bool() {
			break
		}
	}
	z(STOP)
	e := sink(reflect.ValueOf(err == nil))
	return fu.Fnze(err, e)
}

func Chan(c interface{}, stop ...chan struct{}) Source {
	return func() Stream {
		scase := []reflect.SelectCase{{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(c)}}
		wc := fu.WaitCounter{Value: 0}
		return func(index uint64) (v reflect.Value, err error) {
			if index == STOP {
				wc.Stop()
				for _, s := range stop {
					close(s)
				}
			}
			if wc.Wait(index) {
				_, r, ok := reflect.Select(scase)
				if wc.Inc() && ok {
					return r, nil
				}
			}
			return falseValue, nil
		}
	}
}

func List(list interface{}) Source {
	return func() Stream {
		v := reflect.ValueOf(list)
		l := uint64(v.Len())
		flag := fu.AtomicFlag{Value: 1}
		return func(index uint64) (reflect.Value, error) {
			if index < l && flag.State() {
				return v.Index(int(index)), nil
			}
			return falseValue, nil
		}
	}
}

const iniCollectLength = 13

func (zf Source) Collect() (r interface{}, err error) {
	length := 0
	values := reflect.ValueOf((interface{})(nil))
	err = zf.Drain(func(v reflect.Value) error {
		if length == 0 {
			values = reflect.MakeSlice(reflect.SliceOf(v.Type()), 0, iniCollectLength)
		}
		if v.Kind() != reflect.Bool {
			values = reflect.Append(values, v)
			length++
		}
		return nil
	})
	if err != nil {
		return
	}
	return values.Interface(), nil
}

func (zf Source) LuckyCollect() interface{} {
	t, err := zf.Collect()
	if err != nil {
		panic(err)
	}
	return t
}

func (zf Source) Count() (count int, err error) {
	err = zf.Drain(func(v reflect.Value) error {
		if v.Kind() != reflect.Bool {
			count++
		}
		return nil
	})
	return
}

func (zf Source) LuckyCount() int {
	c, err := zf.Count()
	if err != nil {
		panic(err)
	}
	return c
}

func (zf Source) RandFilter(seed int, prob float64, t bool) Source {
	z := zf()
	return func() Stream {
		nr := fu.NaiveRandom{Value: uint32(seed)}
		wc := fu.WaitCounter{Value: 0}
		return func(index uint64) (v reflect.Value, err error) {
			v, err = z(index)
			if index == STOP {
				wc.Stop()
			}
			if wc.Wait(index) {
				if v.Kind() != reflect.Bool {
					p := nr.Float()
					if (t && p <= prob) || (!t && p > prob) {
						v = trueValue // skip
					}
				}
				wc.Inc()
			}
			return
		}
	}
}

func (zf Source) RandSkip(seed int, prob float64) Source {
	return zf.RandFilter(seed, prob, true)
}

func (zf Source) Rand(seed int, prob float64) Source {
	return zf.RandFilter(seed, prob, false)
}

func Error(err error, z ...Stream) Stream {
	return func(index uint64) (reflect.Value, error) {
		if index == STOP && len(z) > 0 {
			z[0](STOP)
		}
		return falseValue, err
	}
}

func Wrap(e interface{}) Stream {
	if stream, ok := e.(Stream); ok {
		return stream
	} else {
		return Error(e.(error))
	}
}

func (zf Source) Chain(zx Source, eqt ...func(a, b reflect.Value) bool) Source {
	return func() Stream {
		z0 := zf()
		z1 := zx()
		b := uint64(0)
		ptr := unsafe.Pointer(nil)
		return func(index uint64) (v reflect.Value, err error) {
			q := atomic.LoadUint64(&b)
			if index == STOP {
				_, err = z0(index)
				_, err1 := z1(index)
				if q > 0 || err == nil {
					err = err1
				}
				return falseValue, err
			}
			if q == 0 || index < q {
				v, err = z0(index)
				if err == nil {
					if v.Kind() == reflect.Bool {
						if !v.Bool() { // end first stream
							atomic.CompareAndSwapUint64(&b, q, index)
							return trueValue, nil
						}
					}
					if q == 0 && atomic.LoadPointer(&ptr) == nil {
						vx := v
						atomic.CompareAndSwapPointer(&ptr, nil, unsafe.Pointer(&vx))
					}
				}
			} else {
				v, err = z1(index - q)
				if err == nil && v.Kind() != reflect.Bool {
					if p := atomic.LoadPointer(&ptr); p != nil {
						vx := (*reflect.Value)(p)
						if v.Type() != vx.Type() {
							return falseValue, zorros.Errorf("chained stream is not compatible")
						}
						for _, f := range eqt {
							if !f(v, *vx) {
								return falseValue, zorros.Errorf("chained stream has non equal value type")
							}
						}
						atomic.CompareAndSwapPointer(&ptr, p, nil)
					}
				}
			}
			return
		}
	}
}
