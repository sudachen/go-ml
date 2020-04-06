package tables

import (
	"github.com/sudachen/go-ml/fu"
	"github.com/sudachen/go-ml/fu/lazy"
	"reflect"
)

/*
FeaturesMapper interface is a features transformation abstraction
*/
type FeaturesMapper interface {
	// MapFeature returns new table with all original columns except features
	// adding one new column with prediction/calculation
	MapFeatures(*Table) (*Table, error)
	// Cloase releases all bounded resources
	Close() error
}

type LambdaMapper func(table *Table) (*Table, error)

func (LambdaMapper) Close() error                           { return nil }
func (l LambdaMapper) MapFeatures(t *Table) (*Table, error) { return l(t) }

/*
Batch is batching abstraction to process lazy streams
*/
type Batch struct {
	int
	lazy.Source
}

/*
Batch transforms lazy stream to a batching flow
*/
func (zf Lazy) Batch(length int) Batch {
	return Batch{length, func() lazy.Stream {
		z := zf()
		wc := fu.WaitCounter{Value: 0}
		columns := []reflect.Value{}
		na := []fu.Bits{}
		names := []string{}
		ac := fu.AtomicCounter{Value: 0}

		return func(index uint64) (v reflect.Value, err error) {
			v, err = z(index)
			if index == lazy.STOP || err != nil {
				wc.Stop()
				return
			}

			x := fu.True
			if wc.Wait(index) {
				if v.Kind() == reflect.Bool {
					if !v.Bool() {
						n := int(ac.Value % uint64(length))
						if ac.Value != 0 {
							if n == 0 {
								n = length
							}
							v = reflect.ValueOf(MakeTable(names, columns, na, n))
						}
						wc.Stop()
					}
					wc.Inc()
					return v, nil
				}

				lr := v.Interface().(fu.Struct)
				ndx := ac.PostInc()
				n := int(ndx % uint64(length))

				if n == 0 {
					if ndx != 0 {
						x = reflect.ValueOf(MakeTable(names, columns, na, length))
					}
					names = lr.Names
					width := len(names)
					columns = make([]reflect.Value, width)
					for i := range columns {
						columns[i] = reflect.MakeSlice(reflect.SliceOf(lr.Columns[i].Type()), 0, length)
					}
					na = make([]fu.Bits, width)
				}

				for i := range lr.Names {
					columns[i] = reflect.Append(columns[i], lr.Columns[i])
					na[i].Set(n, lr.Na.Bit(i))
				}

				wc.Inc()
				return x, nil
			}
			return fu.False, nil
		}
	}}
}

/*
Flat transforms batching to the normal lazy stream
*/
func (zf Batch) Flat() Lazy {
	return func() lazy.Stream {
		z := zf.Source()
		wc := fu.WaitCounter{Value: 0}
		ac := fu.AtomicCounter{Value: 0}
		t := (*Table)(nil)
		row := 0
		return func(index uint64) (v reflect.Value, err error) {
			v = fu.False
			if index == lazy.STOP {
				wc.Stop()
				return
			}
			if wc.Wait(index) {
				if t == nil {
					v, err = z(ac.PostInc())
					if err != nil || (v.Kind() == reflect.Bool && !v.Bool()) {
						wc.Stop()
						return
					}
					if v.Kind() != reflect.Bool {
						t = v.Interface().(*Table)
						row = 0
					}
				}
				if t != nil {
					v = reflect.ValueOf(t.Index(row))
					row++
					if row >= t.Len() {
						t = nil
					}
				}
				wc.Inc()
				return v, nil
			}
			return fu.False, nil
		}
	}
}

/*
Transform transforms streamed data by batches
*/
func (zf Batch) Transform(tf func(int) (FeaturesMapper, error)) Batch {
	return Batch{zf.int, func() lazy.Stream {
		f := fu.AtomicFlag{Value: 0}
		tx, err := tf(zf.int)
		if err != nil {
			return lazy.Error(err)
		}
		z := zf.Source()
		return func(index uint64) (v reflect.Value, err error) {
			v, err = z(index)
			if index == lazy.STOP || err != nil {
				f.Set()
				tx.Close()
				return
			}
			if !f.State() {
				if v.Kind() != reflect.Bool {
					lr := v.Interface().(*Table)
					t, err := tx.MapFeatures(lr)
					if err != nil {
						f.Set()
						return fu.False, err
					}
					return reflect.ValueOf(t), nil
				}
				if v.Bool() {
					return fu.True, nil
				}
				f.Set()
			}
			return fu.False, nil
		}
	}}
}
