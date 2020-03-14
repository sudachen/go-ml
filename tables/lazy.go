package tables

import (
	"github.com/sudachen/go-foo/fu"
	"github.com/sudachen/go-foo/lazy"
	"github.com/sudachen/go-ml/mlutil"
	"reflect"
)

type Lazy lazy.Source
type Sink lazy.Sink

func SourceError(err error) Lazy {
	return func() lazy.Stream {
		return func(_ uint64) (reflect.Value, error) {
			return reflect.Value{}, err
		}
	}
}

func SinkError(err error) Sink {
	return func(_ reflect.Value) error {
		return err
	}
}

func (t *Table) Lazy() Lazy {
	return func() lazy.Stream {
		flag := &lazy.AtomicFlag{Value: 1}
		return func(index uint64) (v reflect.Value, err error) {
			if index == lazy.STOP {
				flag.Clear()
			} else if flag.State() && index < uint64(t.raw.Length) {
				lr := mlutil.Struct{
					Names:   t.raw.Names,
					Columns: make([]reflect.Value, len(t.raw.Names)),
				}
				for i := range t.raw.Columns {
					lr.Columns[i] = t.raw.Columns[i].Index(int(index))
				}
				return reflect.ValueOf(lr), nil
			}
			return reflect.ValueOf(false), nil
		}
	}
}

func (zf Lazy) Map(f interface{}) Lazy {
	return func() lazy.Stream {
		z := zf()
		vf := reflect.ValueOf(f)
		vt := vf.Type()
		or, ir := vt, vt
		if vf.Kind() == reflect.Func {
			ir = vt.In(0)
			or = vt.Out(0)
		} else if vf.Kind() != reflect.Struct {
			panic("only func(struct{...})struct{...} and struct{...} is allowed as an argument of lazy.Map")
		}
		unwrap := mlutil.Unwrapper(ir)
		wrap := mlutil.Wrapper(or)
		return func(index uint64) (v reflect.Value, err error) {
			if v, err = z(index); err != nil || v.Kind() == reflect.Bool {
				return v, err
			}
			x := unwrap(v.Interface().(mlutil.Struct))
			if vf.Kind() == reflect.Func {
				x = vf.Call([]reflect.Value{x})[0]
			}
			return reflect.ValueOf(wrap(x)), nil
		}
	}
}

func (zf Lazy) Transform(f interface{}) Lazy {
	return func() lazy.Stream {
		z := zf()
		vf := reflect.ValueOf(f)
		vt := vf.Type()
		or, ir := vt, vt
		if vf.Kind() == reflect.Func {
			ir = vt.In(0)
			or = vt.Out(0)
		} else if vf.Kind() != reflect.Struct {
			panic("only func(struct{...})struct{...} and struct{...} is allowed as an argument of lazy.Transform")
		}
		unwrap := mlutil.Unwrapper(ir)
		transform := mlutil.Transformer(or)
		return func(index uint64) (v reflect.Value, err error) {
			if v, err = z(index); err != nil || v.Kind() == reflect.Bool {
				return v, err
			}
			x := unwrap(v.Interface().(mlutil.Struct))
			if vf.Kind() == reflect.Func {
				x = vf.Call([]reflect.Value{x})[0]
			}
			return transform(x, v), nil
		}
	}
}

func (zf Lazy) Filter(f interface{}) Lazy {
	return func() lazy.Stream {
		z := zf()
		vf := reflect.ValueOf(f)
		vt := vf.Type()
		unwrap := mlutil.Unwrapper(vt.In(0))
		return func(index uint64) (v reflect.Value, err error) {
			if v, err = z(index); err != nil || v.Kind() == reflect.Bool {
				return v, err
			}
			x := unwrap(v.Interface().(mlutil.Struct))
			if vf.Call([]reflect.Value{x})[0].Bool() {
				return
			}
			return reflect.ValueOf(true), nil
		}
	}
}

func (z Lazy) First(n int) Lazy {
	return Lazy(lazy.Source(z).First(n))
}

func (z Lazy) Parallel(concurrency ...int) Lazy {
	return Lazy(lazy.Source(z).Parallel(concurrency...))
}

const iniCollectLength = 13
const maxChankLength = 10000

func (z Lazy) Collect() (t *Table, err error) {
	length := 0
	columns := []reflect.Value{}
	names := []string{}
	na := []mlutil.Bits{}
	err = z.Drain(func(v reflect.Value) error {
		if v.Kind() != reflect.Bool {
			lr := v.Interface().(mlutil.Struct)
			if length == 0 {
				names = lr.Names
				columns = make([]reflect.Value, len(names))
				na = make([]mlutil.Bits, len(names))
				for i, x := range lr.Columns {
					columns[i] = reflect.MakeSlice(reflect.SliceOf(x.Type()), 0, iniCollectLength)
				}
			}
			for i, x := range lr.Columns {
				columns[i] = reflect.Append(columns[i], x)
				na[i].Set(length, lr.Na.Bit(i))
			}
			length++
		}
		return nil
	})
	if err != nil {
		return
	}
	return MakeTable(names, columns, na, length), nil
}

func (z Lazy) LuckyCollect() *Table {
	t, err := z.Collect()
	if err != nil {
		panic(err)
	}
	return t
}

func (z Lazy) Drain(sink Sink) (err error) {
	return lazy.Source(z).Drain(sink)
}

func (z Lazy) LuckySink(sink Sink) {
	if err := z.Drain(sink); err != nil {
		panic(err)
	}
}

func (z Lazy) Count() (int, error) {
	return lazy.Source(z).Count()
}

func (z Lazy) LuckyCount() int {
	c, err := z.Count()
	if err != nil {
		panic(err)
	}
	return c
}

func (z Lazy) Rand(seed int, prob float64) Lazy {
	return Lazy(lazy.Source(z).Rand(seed, prob))
}

func (z Lazy) RandSkip(seed int, prob float64) Lazy {
	return Lazy(lazy.Source(z).RandSkip(seed, prob))
}

func (zf Lazy) RandomFlag(column string, seed int, prob float64) Lazy {
	z := zf()
	return func() lazy.Stream {
		nr := fu.NaiveRandom{Value: uint32(seed)}
		wc := lazy.WaitCounter{Value: 0}
		cj := -1
		return func(index uint64) (v reflect.Value, err error) {
			v, err = z(index)
			if index == lazy.STOP {
				wc.Stop()
			}
			if wc.Wait(index) {
				if err == nil && v.Kind() != reflect.Bool {
					lr := v.Interface().(mlutil.Struct)
					if cj < 0 {
						cj = fu.IndexOf(column, lr.Names)
					}
					p := nr.Float()
					val := reflect.ValueOf(p < prob)
					lr = lr.Copy(cj + 1)
					if cj < 0 {
						lr.Names = append(lr.Names, column)
						lr.Columns = append(lr.Columns, val)
					} else {
						lr.Columns[cj] = val
						lr.Na.Set(cj, false)
					}
					v = reflect.ValueOf(lr)
				}
				wc.Inc()
			}
			return
		}
	}
}

func (zf Lazy) Round(prec int) Lazy {
	z := zf()
	return func() lazy.Stream {
		return func(index uint64) (v reflect.Value, err error) {
			v, err = z(index)
			if err != nil || v.Kind() == reflect.Bool {
				return
			}
			lrx := v.Interface().(mlutil.Struct)
			lr := lrx.Copy(0)
			for i, c := range lr.Columns {
				switch c.Kind() {
				case reflect.Float32:
					lr.Columns[i] = reflect.ValueOf(fu.Round32(float32(c.Float()), prec))
				case reflect.Float64:
					lr.Columns[i] = reflect.ValueOf(fu.Round64(c.Float(), prec))
				}
			}
			return reflect.ValueOf(lr), nil
		}
	}
}
