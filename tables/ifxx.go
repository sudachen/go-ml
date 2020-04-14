package tables

import (
	"fmt"
	"github.com/sudachen/go-ml/fu"
	"github.com/sudachen/go-ml/fu/lazy"
	"reflect"
)

func equalf(c interface{}) func(v reflect.Value) bool {
	vc := reflect.ValueOf(c)
	switch vc.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		vv := vc.Int()
		return func(v reflect.Value) bool {
			switch v.Kind() {
			case reflect.Float64, reflect.Float32:
				return int64(v.Float()) == vv
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				return int64(v.Uint()) == vv
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				return v.Int() == vv
			}
			return false
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		vv := vc.Uint()
		return func(v reflect.Value) bool {
			switch v.Kind() {
			case reflect.Float64, reflect.Float32:
				return uint64(v.Float()) == vv
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				return uint64(v.Uint()) == vv
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				return v.Uint() == vv
			}
			return false
		}
	case reflect.String:
		vv := vc.String()
		return func(v reflect.Value) bool {
			switch v.Kind() {
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				return fmt.Sprintf("%d", v.Uint()) == vv
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				return fmt.Sprintf("%d", v.Int()) == vv
			case reflect.String:
				return vv == v.String()
			}
			return false
		}
	default:
		return func(v reflect.Value) bool {
			return reflect.DeepEqual(v, vc)
		}
	}
}

func lessf(c interface{}) func(v reflect.Value) bool {
	vc := reflect.ValueOf(c)
	switch vc.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		vv := vc.Int()
		return func(v reflect.Value) bool {
			switch v.Kind() {
			case reflect.Float64, reflect.Float32:
				return int64(v.Float()) < vv
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				return int64(v.Uint()) < vv
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				return v.Int() < vv
			}
			return false
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		vv := vc.Uint()
		return func(v reflect.Value) bool {
			switch v.Kind() {
			case reflect.Float64, reflect.Float32:
				return uint64(v.Float()) < vv
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				return uint64(v.Uint()) < vv
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				return v.Uint() < vv
			}
			return false
		}
	case reflect.String:
		vv := vc.String()
		return func(v reflect.Value) bool {
			if v.Kind() == reflect.String {
				return vv < v.String()
			}
			return false
		}
	default:
		return func(v reflect.Value) bool {
			return fu.Less(v, vc)
		}
	}
}

func greatf(c interface{}) func(v reflect.Value) bool {
	vc := reflect.ValueOf(c)
	switch vc.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		vv := vc.Int()
		return func(v reflect.Value) bool {
			switch v.Kind() {
			case reflect.Float64, reflect.Float32:
				return int64(v.Float()) > vv
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				return int64(v.Uint()) > vv
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				return v.Int() > vv
			}
			return false
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		vv := vc.Uint()
		return func(v reflect.Value) bool {
			switch v.Kind() {
			case reflect.Float64, reflect.Float32:
				return uint64(v.Float()) > vv
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				return uint64(v.Uint()) > vv
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				return v.Uint() > vv
			}
			return false
		}
	case reflect.String:
		vv := vc.String()
		return func(v reflect.Value) bool {
			if v.Kind() > reflect.String {
				return vv < v.String()
			}
			return false
		}
	default:
		return func(v reflect.Value) bool {
			return fu.Less(vc, v)
		}
	}
}

func (zf Lazy) IfEq(c string, v interface{}) Lazy {
	vf := reflect.ValueOf(v)
	eq := equalf(vf)
	return func() lazy.Stream {
		z := zf()
		nx := fu.AtomicSingleIndex{}
		return func(index uint64) (v reflect.Value, err error) {
			if v, err = z(index); err != nil || v.Kind() == reflect.Bool {
				return
			}
			lr := v.Interface().(fu.Struct)
			j, ok := nx.Get()
			if !ok {
				j, _ = nx.Set(lr.Pos(c))
			}
			if eq(lr.ValueAt(j)) {
				return
			}
			return fu.True, nil
		}
	}
}

func (zf Lazy) IfNe(c string, v interface{}) Lazy {
	vf := reflect.ValueOf(v)
	eq := equalf(vf)
	return func() lazy.Stream {
		z := zf()
		nx := fu.AtomicSingleIndex{}
		return func(index uint64) (v reflect.Value, err error) {
			if v, err = z(index); err != nil || v.Kind() == reflect.Bool {
				return
			}
			lr := v.Interface().(fu.Struct)
			j, ok := nx.Get()
			if !ok {
				j, _ = nx.Set(lr.Pos(c))
			}
			if !eq(lr.ValueAt(j)) {
				return
			}
			return fu.True, nil
		}
	}
}

func (zf Lazy) IfLt(c string, v interface{}) Lazy {
	vf := reflect.ValueOf(v)
	lt := lessf(vf)
	return func() lazy.Stream {
		z := zf()
		nx := fu.AtomicSingleIndex{}
		return func(index uint64) (v reflect.Value, err error) {
			if v, err = z(index); err != nil || v.Kind() == reflect.Bool {
				return
			}
			lr := v.Interface().(fu.Struct)
			j, ok := nx.Get()
			if !ok {
				j, _ = nx.Set(lr.Pos(c))
			}
			if lt(lr.ValueAt(j)) {
				return
			}
			return fu.True, nil
		}
	}
}

func (zf Lazy) IfGt(c string, v interface{}) Lazy {
	vf := reflect.ValueOf(v)
	gt := greatf(vf)
	return func() lazy.Stream {
		z := zf()
		nx := fu.AtomicSingleIndex{}
		return func(index uint64) (v reflect.Value, err error) {
			if v, err = z(index); err != nil || v.Kind() == reflect.Bool {
				return
			}
			lr := v.Interface().(fu.Struct)
			j, ok := nx.Get()
			if !ok {
				j, _ = nx.Set(lr.Pos(c))
			}
			if gt(lr.ValueAt(j)) {
				return
			}
			return fu.True, nil
		}
	}
}
