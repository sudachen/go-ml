package fu

import (
	"reflect"
)

/*
CopySlice returns copy of given slice
*/
func CopySlice(a interface{}) interface{} {
	v := reflect.ValueOf(a)
	r := reflect.MakeSlice(v.Type(), v.Len(), v.Len())
	reflect.Copy(r, v)
	return r.Interface()
}
