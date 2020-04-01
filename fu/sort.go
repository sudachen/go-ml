package fu

import (
	"reflect"
	"sort"
)

func Sort(a interface{}) {
	v := reflect.ValueOf(a)
	t := v.Type().Elem()
	switch t.Kind() {
	case reflect.Int:
		sort.Ints(a.([]int))
	case reflect.String:
		sort.Strings(a.([]string))
	default:
		sort.Slice(a, func(i, j int) bool { return Less(v.Index(i), v.Index(j)) })
	}
}

func Sorted(a interface{}) interface{} {
	b := CopySlice(a)
	Sort(b)
	return b
}
