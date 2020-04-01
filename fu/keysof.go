package fu

import (
	"reflect"
)

/*
KeysOf returns list of keys of a map
*/
func KeysOf(m interface{}) interface{} {
	v := reflect.ValueOf(m)
	if v.Kind() != reflect.Map {
		panic("parameter is not a map")
	}
	k := v.MapKeys()
	keys := reflect.MakeSlice(reflect.SliceOf(v.Type().Key()), len(k), len(k))
	for i, s := range k {
		keys.Index(i).Set(s)
	}
	return keys.Interface()
}

/*
SortedKeysOf returns sorted list of keys of a map
*/
func SortedKeysOf(m interface{}) interface{} {
	keys := KeysOf(m)
	Sort(keys)
	return keys
}
