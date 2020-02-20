//
// Package util implements utility code using in go-ml components
//
package util

import (
	"reflect"
)

func IndexOf(a string, b []string) int {
	for i, v := range b {
		if v == a {
			return i
		}
	}
	return -1
}

func MapInterface(m map[string]reflect.Value) map[string]interface{} {
	r := map[string]interface{}{}
	for k, v := range m {
		r[k] = v.Interface()
	}
	return r
}
