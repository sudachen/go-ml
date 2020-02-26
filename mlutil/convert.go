package mlutil

import (
	"fmt"
	"reflect"
)

/*
Convert returns the argument with specified type
*/
func Convert(v reflect.Value, tp reflect.Type) interface{} {
	vt := v.Type()
	if v.Kind() == reflect.Slice {
		if vt.Elem() == tp {
			return v.Interface()
		}
		if tp.Kind() == reflect.String {
			rs := make([]string, v.Len(), v.Len())
			for i := range rs {
				rs[i] = fmt.Sprint(v.Index(i).Interface())
			}
			return rs
		} else if vt.Elem().ConvertibleTo(tp) {
			r := reflect.MakeSlice(reflect.SliceOf(tp), v.Len(), v.Len())
			for i := 0; i < v.Len(); i++ {
				x := v.Index(i).Convert(tp)
				r.Index(i).Set(x)
			}
			return r.Interface()
		}
	} else {
		if vt == tp {
			return v.Interface()
		}
		if tp.Kind() == reflect.String {
			return fmt.Sprint(v.Interface())
		} else if v.Type().ConvertibleTo(tp) {
			return v.Convert(tp).Interface()
		}
	}
	panic("can't convert")
}
