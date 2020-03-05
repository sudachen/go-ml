package mlutil

import (
	"fmt"
	"github.com/sudachen/go-foo/fu"
	"math"
	"reflect"
	"strconv"
)

func Isna(v reflect.Value) bool {
	if !v.IsValid() {
		return true
	}
	switch v.Kind() {
	case reflect.Float32, reflect.Float64:
		return math.IsNaN(v.Float())
	}
	return false
}

func Nan(tp reflect.Type) reflect.Value {
	switch tp.Kind() {
	case reflect.Float32:
		return reflect.ValueOf(float32(math.NaN()))
	case reflect.Float64:
		return reflect.ValueOf(math.NaN())
	}
	return reflect.Zero(tp)
}

func ConvertSlice(v reflect.Value, na Bits, tp reflect.Type, nocopy ...bool) reflect.Value {
	L := v.Len()
	vt := v.Type().Elem()
	if vt == tp && fu.Fnzb(nocopy...) {
		return v.Slice(0, L)
	}
	r := reflect.MakeSlice(reflect.SliceOf(tp), L, L)
	if vt == tp {
		reflect.Copy(r, v)
	} else {
		for i := 0; i < L; i++ {
			r.Index(i).Set(Convert(v.Index(i), na.Bit(i), tp))
		}
	}
	return r
}

func Convert(v reflect.Value, na bool, tp reflect.Type) reflect.Value {
	if na {
		return Nan(tp)
	}
	if v.Type() == tp {
		return v
	} else if tp.Kind() == reflect.String {
		return reflect.ValueOf(fmt.Sprint(v.Interface()))
	} else if v.Kind() == reflect.String {
		switch tp.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			x, err := strconv.ParseInt(v.String(), 10, 64)
			if err != nil {
				panic(err)
			}
			return reflect.ValueOf(x).Convert(tp)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			x, err := strconv.ParseUint(v.String(), 10, 64)
			if err != nil {
				panic(err)
			}
			return reflect.ValueOf(x).Convert(tp)
		case reflect.Float32, reflect.Float64:
			x, err := strconv.ParseFloat(v.String(), 64)
			if err != nil {
				panic(err)
			}
			return reflect.ValueOf(x).Convert(tp)
		}
	} else if tp.Kind() == reflect.Float32 {
		switch v.Kind() {
		case reflect.Float64:
			return reflect.ValueOf(float32(v.Float()))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return reflect.ValueOf(float32(v.Uint()))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return reflect.ValueOf(float32(v.Int()))
		}
	}
	return v.Convert(tp)
}
