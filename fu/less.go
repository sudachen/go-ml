package fu

import (
	"reflect"
)

/*
Less compares two values and returns true if the left value is less than right one otherwise it returns false
*/
func Less(a, b reflect.Value) bool {
	if a.Kind() != b.Kind() {
		panic("values must have the same type")
	}
	if m := a.MethodByName("Less"); m.IsValid() {
		t := m.Type()
		if t.NumOut() == 1 && t.NumIn() == 1 && t.Out(0).Kind() == reflect.Bool && t.In(0) == a.Type() {
			return m.Call([]reflect.Value{b})[0].Bool()
		}
	}
	switch a.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return a.Int() < b.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return a.Uint() < b.Uint()
	case reflect.Float32, reflect.Float64:
		return a.Float() < b.Float()
	case reflect.String:
		return a.String() < b.String()
	case reflect.Ptr, reflect.Interface:
		if a.IsNil() && !b.IsNil() {
			return true
		} else if b.IsNil() {
			return false
		}
		return Less(a.Elem(), b.Elem())
	case reflect.Struct:
		N := a.NumField()
		for i := 0; i < N; i++ {
			if Less(a.Field(i), b.Field(i)) {
				return true
			} else if Less(b.Field(i), a.Field(i)) {
				return false
			}
		}
		return false
	case reflect.Array:
		if a.Len() != b.Len() {
			panic("values must have the same type")
		}
		N := a.Len()
		for i := 0; i < N; i++ {
			if Less(a.Index(i), b.Index(i)) {
				return true
			} else if Less(b.Index(i), a.Index(i)) {
				return false
			}
		}
		return false
	default:
		panic("only int,float,string,struct,array are allowed to compare")
	}
}
