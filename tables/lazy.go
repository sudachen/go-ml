package tables

import (
	"github.com/sudachen/go-foo/fu"
	"github.com/sudachen/go-foo/lazy"
	"github.com/sudachen/go-ml/mlutil"
	"reflect"
)

type Lazy lazy.Source
type Sink lazy.Sink

func (Lazy) IsLazy() bool     { return true }
func (zf Lazy) Table() *Table { return zf.LuckyCollect() }
func (zf Lazy) Lazy() Lazy    { return zf }

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

func (zf Lazy) First(n int) Lazy {
	return Lazy(lazy.Source(zf).First(n))
}

func (zf Lazy) Parallel(concurrency ...int) Lazy {
	return Lazy(lazy.Source(zf).Parallel(concurrency...))
}

const iniCollectLength = 13
const maxChankLength = 10000

func (zf Lazy) Collect() (t *Table, err error) {
	length := 0
	columns := []reflect.Value{}
	names := []string{}
	na := []mlutil.Bits{}
	err = zf.Drain(func(v reflect.Value) error {
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

func (zf Lazy) LuckyCollect() *Table {
	t, err := zf.Collect()
	if err != nil {
		panic(fu.Panic(err))
	}
	return t
}

func (zf Lazy) Drain(sink Sink) (err error) {
	return lazy.Source(zf).Drain(sink)
}

func (zf Lazy) LuckySink(sink Sink) {
	if err := zf.Drain(sink); err != nil {
		panic(fu.Panic(err))
	}
}

func (zf Lazy) Count() (int, error) {
	return lazy.Source(zf).Count()
}

func (zf Lazy) LuckyCount() int {
	c, err := zf.Count()
	if err != nil {
		panic(fu.Panic(err))
	}
	return c
}

func (zf Lazy) Rand(seed int, prob float64) Lazy {
	return Lazy(lazy.Source(zf).Rand(seed, prob))
}

func (zf Lazy) RandSkip(seed int, prob float64) Lazy {
	return Lazy(lazy.Source(zf).RandSkip(seed, prob))
}

func (zf Lazy) RandomFlag(c string, seed int, prob float64) Lazy {
	return func() lazy.Stream {
		z := zf()
		nr := fu.NaiveRandom{Value: uint32(seed)}
		wc := lazy.WaitCounter{Value: 0}
		return func(index uint64) (v reflect.Value, err error) {
			v, err = z(index)
			if index == lazy.STOP {
				wc.Stop()
			}
			if wc.Wait(index) {
				if err == nil && v.Kind() != reflect.Bool {
					lr := v.Interface().(mlutil.Struct)
					p := nr.Float()
					val := reflect.ValueOf(p < prob)
					v = reflect.ValueOf(lr.Set(c, val))
				}
				wc.Inc()
			}
			return
		}
	}
}

func (zf Lazy) Round(prec int) Lazy {
	return func() lazy.Stream {
		z := zf()
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

func (zf Lazy) IfFlag(c string) Lazy {
	return func() lazy.Stream {
		z := zf()
		return func(index uint64) (v reflect.Value, err error) {
			v, err = z(index)
			if err != nil || v.Kind() == reflect.Bool {
				return
			}
			lr := v.Interface().(mlutil.Struct)
			if j := fu.IndexOf(c, lr.Names); j >= 0 && lr.Columns[j].Bool() {
				return
			}
			return mlutil.True, nil
		}
	}
}

func (zf Lazy) IfNotFlag(c string) Lazy {
	return func() lazy.Stream {
		z := zf()
		return func(index uint64) (v reflect.Value, err error) {
			v, err = z(index)
			if err != nil || v.Kind() == reflect.Bool {
				return
			}
			lr := v.Interface().(mlutil.Struct)
			if j := fu.IndexOf(c, lr.Names); j < 0 || !lr.Columns[j].Bool() {
				return
			}
			return mlutil.True, nil
		}
	}
}

func (zf Lazy) True(c string) Lazy {
	return func() lazy.Stream {
		z := zf()
		return func(index uint64) (v reflect.Value, err error) {
			v, err = z(index)
			if err != nil || v.Kind() == reflect.Bool {
				return
			}
			lr := v.Interface().(mlutil.Struct)
			return reflect.ValueOf(lr.Set(c, mlutil.True)), nil
		}
	}
}

func (zf Lazy) False(c string) Lazy {
	return func() lazy.Stream {
		z := zf()
		return func(index uint64) (v reflect.Value, err error) {
			v, err = z(index)
			if err != nil || v.Kind() == reflect.Bool {
				return
			}
			lr := v.Interface().(mlutil.Struct)
			return reflect.ValueOf(lr.Set(c, mlutil.False)), nil
		}
	}
}

func (zf Lazy) Chain(zx Lazy) Lazy {
	return Lazy(lazy.Source(zf).Chain(lazy.Source(zx), func(a, b reflect.Value) (eqt bool) {
		if lr, ok := a.Interface().(mlutil.Struct); ok {
			if lrx, ok := b.Interface().(mlutil.Struct); ok {
				if len(lrx.Names) != len(lr.Names) {
					for i, n := range lrx.Names {
						if n != lr.Names[i] || lrx.Columns[i].Type() != lr.Columns[i].Type() {
							return false
						}
					}
					eqt = true
				}
			}
		}
		return
	}))
}
