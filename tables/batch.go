package tables

import (
	"github.com/sudachen/go-foo/lazy"
	"github.com/sudachen/go-ml/base"
	"github.com/sudachen/go-ml/mlutil"
	"reflect"
)

type Batch lazy.Source

func (zf Lazy) Batch(length int) Batch {
	return func() lazy.Stream {
		z := zf()
		wc := lazy.WaitCounter{Value: 0}
		vC := make(chan reflect.Value)
		tC := make(chan reflect.Value, 1)

		go func() {
			for {
				v, ok := <-vC
				if ok {
					lr := v.Interface().(base.Struct)
					t := MakeTable(
						lr.Names,
						make([]reflect.Value, len(lr.Names)),
						make([]mlutil.Bits, len(lr.Names)),
						0)
					for j := range lr.Names {
						q := reflect.MakeSlice(reflect.SliceOf(lr.Columns[j].Type()), 0, length)
						t.raw.Columns[j] = reflect.Append(q, lr.Columns[j])
						t.raw.Na[j].Set(0, lr.Na.Bit(j))
					}
					t.raw.Length++
				l:
					for n := 1; n < length; n++ {
						if v, ok = <-vC; !ok {
							break l
						}
						lr := v.Interface().(base.Struct)
						for j := range lr.Names {
							t.raw.Columns[j] = reflect.Append(t.raw.Columns[j], lr.Columns[j])
							t.raw.Na[j].Set(n, lr.Na.Bit(j))
						}
						t.raw.Length++
					}
					tC <- reflect.ValueOf(t)
				} else {
					close(tC)
					return
				}
			}
		}()

		stopFlag := lazy.AtomicFlag{0}
		stop := func() {
			if stopFlag.Set() {
				close(vC)
				select {
				case <-tC:
				default:
				}
			}
		}

		return func(index uint64) (v reflect.Value, err error) {
			v, err = z(index)
			if index == lazy.STOP || err != nil {
				stop()
				wc.Stop()
			}

			if wc.Wait(index) {
				ok := false
				x := v

				select {
				case v, ok = <-tC:
					if !ok {
						stop()
						wc.Stop()
						return reflect.ValueOf(false), nil
					}
				default:
				}

				if x.Kind() != reflect.Bool {
					vC <- x
				}

				wc.Inc()

				if ok {
					return
				}
				if x.Kind() == reflect.Bool && !x.Bool() {
					stop()
				}
				return reflect.ValueOf(true), nil
			}
			return reflect.ValueOf(false), nil
		}
	}
}

func (z Batch) Parallel(concurrency ...int) Batch {
	return Batch(lazy.Source(z).Parallel(concurrency...))
}

func (z Batch) Drain(sink Sink) (err error) {
	return lazy.Source(z).Drain(sink)
}

func (z Batch) LuckyDrain(sink Sink) {
	if err := z.Drain(sink); err != nil {
		panic(err)
	}
}
