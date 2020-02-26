package csv

import (
	"github.com/sudachen/go-ml/internal"
	"reflect"
	"strconv"
	"time"
)

type resolver interface {
	resolve() mapper
}

func (x Column) As(n string) RenamedColumn {
	return RenamedColumn{string(x), n}
}

func (x Column) resolve() mapper {
	return mapper{string(x), string(x), nil, nil}
}

func (x RenamedColumn) resolve() mapper {
	return mapper{x.CsvCol, x.TableCol, nil, nil}
}

func (x String) As(n string) RenamedString {
	return RenamedString{string(x), n}
}

func (x String) resolve() mapper {
	return mapper{string(x), string(x), internal.StringType, nil}
}
func (x RenamedString) resolve() mapper {
	return mapper{x.CsvCol, x.TableCol, internal.StringType, nil}
}

func (x Int) As(n string) RenamedInt {
	return RenamedInt{string(x), n}
}

func converti(s string) (value reflect.Value, na bool, err error) {
	if s == "" {
		return internal.IntZero, true, nil
	}
	v, err := strconv.ParseInt(s, 10, 64)
	value = reflect.ValueOf(int(v))
	return
}

func (x Int) resolve() mapper {
	return mapper{string(x), string(x), internal.IntType, converti}
}

func (x RenamedInt) resolve() mapper {
	return mapper{x.CsvCol, x.TableCol, internal.IntType, converti}
}

func (x Float32) As(n string) RenamedFloat32 {
	return RenamedFloat32{string(x), n}
}

func convert32f(s string) (value reflect.Value, na bool, err error) {
	if s == "" {
		return internal.IntZero, true, nil
	}
	v, err := strconv.ParseFloat(s, 64)
	value = reflect.ValueOf(float32(v))
	return
}

func (x Float32) resolve() mapper {
	return mapper{string(x), string(x), internal.IntType, convert32f}
}

func (x RenamedFloat32) resolve() mapper {
	return mapper{x.CsvCol, x.TableCol, internal.IntType, convert32f}
}

func (x Float64) As(n string) RenamedFloat64 { return RenamedFloat64{string(x), n} }

func convert64f(s string) (value reflect.Value, na bool, err error) {
	if s == "" {
		return internal.IntZero, true, nil
	}
	v, err := strconv.ParseFloat(s, 32)
	value = reflect.ValueOf(v)
	return
}

func (x Float64) resolve() mapper {
	return mapper{string(x), string(x), internal.IntType, convert64f}
}

func (x RenamedFloat64) resolve() mapper {
	return mapper{x.CsvCol, x.TableCol, internal.IntType, convert64f}
}

func (x Time) As(n string) RenamedTime {
	return RenamedTime{string(x), n}
}

func (x Time) Like(layout string) TimeLayout {
	return TimeLayout{string(x), layout}
}

func (x TimeLayout) As(n string) RenamedTimeLayout {
	return RenamedTimeLayout{x.Col, n, x.Layout}
}

func convertts(s string, layout string) (value reflect.Value, na bool, err error) {
	if s == "" {
		return internal.IntZero, true, nil
	}
	v, err := strconv.ParseFloat(s, 32)
	value = reflect.ValueOf(v)
	return
}

func (x Time) resolve() mapper {
	return mapper{string(x), string(x), internal.IntType,
		func(s string) (value reflect.Value, na bool, err error) {
			return convertts(s, time.RFC3339)
		}}
}

func (x RenamedTime) resolve() mapper {
	return mapper{x.CsvCol, x.TableCol, internal.IntType,
		func(s string) (value reflect.Value, na bool, err error) {
			return convertts(s, time.RFC3339)
		}}
}

func (x TimeLayout) resolve() mapper {
	return mapper{x.Col, x.Col, internal.IntType,
		func(s string) (value reflect.Value, na bool, err error) {
			return convertts(s, x.Layout)
		}}
}

func (x RenamedTimeLayout) resolve() mapper {
	return mapper{x.CsvCol, x.TableCol, internal.IntType,
		func(s string) (value reflect.Value, na bool, err error) {
			return convertts(s, x.Layout)
		}}
}
