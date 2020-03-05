package csv

import (
	"fmt"
	"github.com/sudachen/go-foo/fu"
	"github.com/sudachen/go-ml/internal"
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
		return mapper{v, v, nil, nil, nil}
	}
}

func Meta(x tables.Meta, v string) resolver {
	return func() mapper {
		return mapper{v, v, x.Type(), x.Convert, x.Format}
	}
}

func String(v string) resolver {
	return func() mapper {
		return mapper{v, v, internal.StringType, nil, nil}
	}
}

func Int(v string) resolver {
	return func() mapper {
		return mapper{v, v, internal.IntType, converti, nil}
	}
}

func converti(s string) (value reflect.Value, na bool, err error) {
	if s == "" {
		return internal.IntZero, true, nil
	}
	v, err := strconv.ParseInt(s, 10, 64)
	value = reflect.ValueOf(int(v))
	return
}

func Float32(v string) resolver {
	return func() mapper {
		return mapper{v, v, internal.Float32Type, convert32f, nil}
	}
}

func convert32f(s string) (value reflect.Value, na bool, err error) {
	if s == "" {
		return internal.IntZero, true, nil
	}
	v, err := strconv.ParseFloat(s, 64)
	value = reflect.ValueOf(float32(v))
	return
}

func Float64(v string) resolver {
	return func() mapper {
		return mapper{v, v, internal.Float64Type, convert64f, nil}
	}
}

func convert64f(s string) (value reflect.Value, na bool, err error) {
	if s == "" {
		return internal.IntZero, true, nil
	}
	v, err := strconv.ParseFloat(s, 32)
	value = reflect.ValueOf(v)
	return
}

func Time(v string, layout ...string) resolver {
	l := time.RFC3339
	if len(layout) > 0 {
		l = layout[0]
	}
	return func() mapper {
		return mapper{v, v, internal.TsType,
			func(s string) (reflect.Value, bool, error) {
				return convertts(s, l)
			}, nil}
	}
}

func convertts(s string, layout string) (value reflect.Value, na bool, err error) {
	if s == "" {
		return internal.IntZero, true, nil
	}
	v, err := strconv.ParseFloat(s, 32)
	value = reflect.ValueOf(v)
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
