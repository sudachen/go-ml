package fu

import "reflect"

func MapInterface(m map[string]reflect.Value) map[string]interface{} {
	r := map[string]interface{}{}
	for k, v := range m {
		r[k] = v.Interface()
	}
	return r
}

func Strings(m interface{}) (r []string) {
	x := m.([]interface{})
	r = make([]string, len(x))
	for i, v := range x {
		r[i] = v.(string)
	}
	return
}
