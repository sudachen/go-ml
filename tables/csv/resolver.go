package csv

import (
	"fmt"
	"github.com/sudachen/go-foo/fu"
	"github.com/sudachen/go-ml/mlutil"
	"github.com/sudachen/go-ml/tables"
	"math"
	"reflect"
	"strconv"
	"time"
)

type resolver func() mapper

func (r resolver) As(n string) resolver {
	return func() mapper {
		m := r()
		m.TableCol = n
		return m
	}
}

func Column(v string) resolver {
	return func() mapper {
		return Mapper(v, v, nil, nil, nil)
	}
}

func (r resolver) Group(v string) resolver {
	return func() mapper {
		g := r()
		z := tables.Xtensor{g.valueType}
		x := g
		x.TableCol = v
		x.group = true
		x.valueType = z.Type()
		x.convert = func(value string, field *reflect.Value, index, width int) (_ bool, err error) {
			err = z.ConvertElm(value, field, index, width)
			return
		}
		return x
	}
}

func Tensor32f(v string) resolver {
	return func() mapper {
		x := tables.Xtensor{mlutil.Float32}
		return Mapper(v, v, x.Type(), x.Convert, x.Format)
	}
}

func Tensor64f(v string) resolver {
	return func() mapper {
		x := tables.Xtensor{mlutil.Float64}
		return Mapper(v, v, x.Type(), x.Convert, x.Format)
	}
}

func Tensor8u(v string) resolver {
	return func() mapper {
		x := tables.Xtensor{mlutil.Byte}
		return Mapper(v, v, x.Type(), x.Convert, x.Format)
	}
}

func Tensor8f(v string) resolver {
	return func() mapper {
		x := tables.Xtensor{mlutil.Fixed8Type}
		return Mapper(v, v, x.Type(), x.Convert, x.Format)
	}
}

func Meta(x tables.Meta, v string) resolver {
	return func() mapper {
		return Mapper(v, v, x.Type(), x.Convert, x.Format)
	}
}

func String(v string) resolver {
	return func() mapper {
		return Mapper(v, v, mlutil.String, nil, nil)
	}
}

func Int(v string) resolver {
	return func() mapper {
		return Mapper(v, v, mlutil.Int, converti, nil)
	}
}

func converti(s string, value *reflect.Value, _, _ int) (na bool, err error) {
	if s == "" {
		*value = mlutil.IntZero
		return true, nil
	}
	v, err := strconv.ParseInt(s, 10, 64)
	*value = reflect.ValueOf(int(v))
	return
}

func Fixed8(v string) resolver {
	return func() mapper {
		return Mapper(v, v, mlutil.Fixed8Type, convert8f, nil)
	}
}

func convert8f(s string, value *reflect.Value, _, _ int) (na bool, err error) {
	if s == "" {
		*value = mlutil.Fixed8Zero
		return true, nil
	}
	f, err := mlutil.Fast8f(s)
	*value = reflect.ValueOf(f)
	return
}

func Float32(v string) resolver {
	return func() mapper {
		return Mapper(v, v, mlutil.Float32, convert32f, nil)
	}
}

func convert32f(s string, value *reflect.Value, _, _ int) (na bool, err error) {
	if s == "" {
		*value = mlutil.Float32Zero
		return true, nil
	}
	f, err := mlutil.Fast32f(s)
	*value = reflect.ValueOf(f)
	return
}

func Float64(v string) resolver {
	return func() mapper {
		return Mapper(v, v, mlutil.Float64, convert64f, nil)
	}
}

func convert64f(s string, value *reflect.Value, _, _ int) (na bool, err error) {
	if s == "" {
		*value = mlutil.Float64Zero
		return true, nil
	}
	v, err := strconv.ParseFloat(s, 32)
	*value = reflect.ValueOf(v)
	return
}

func Time(v string, layout ...string) resolver {
	l := time.RFC3339
	if len(layout) > 0 {
		l = layout[0]
	}
	return func() mapper {
		return Mapper(v, v, mlutil.Ts,
			func(s string, value *reflect.Value, _, _ int) (bool, error) {
				return convertts(s, l, value)
			}, nil)
	}
}

func convertts(s string, layout string, value *reflect.Value) (na bool, err error) {
	if s == "" {
		*value = mlutil.TsZero
		return true, nil
	}
	v, err := strconv.ParseFloat(s, 32)
	*value = reflect.ValueOf(v)
	return
}

func (r resolver) Round(n ...int) resolver {
	return func() mapper {
		m := r()
		xf := m.format
		m.format = func(v reflect.Value, na bool) string {
			if !na {
				if v.Kind() == reflect.Float64 || v.Kind() == reflect.Float32 {
					if len(n) > 0 && n[0] > 0 {
						v = reflect.ValueOf(fu.Round64(v.Float(), n[0]))
					} else {
						v = reflect.ValueOf(math.Round(v.Float()))
					}
				}
			}
			return format(v, na, xf)
		}
		return m
	}
}

func format(v reflect.Value, na bool, xf func(reflect.Value, bool) string) string {
	if xf != nil {
		return xf(v, na)
	}
	if na {
		return ""
	} else {
		return fmt.Sprint(v.Interface())
	}
}
