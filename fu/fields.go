package fu

import "reflect"

func FieldsOf(s interface{}) []string {
	v := reflect.TypeOf(s)
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		panic("only struct{} and &struct{} allowed as an argument")
	}
	r := []string{}
	for i := 0; i < v.NumField(); i++ {
		r = append(r, v.Field(i).Name)
	}
	return r
}

func AsMap(s interface{}) map[string]reflect.Value {
	v := reflect.ValueOf(s)
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		panic("only struct{} and &struct{} allowed as an argument")
	}
	vt := v.Type()
	r := map[string]reflect.Value{}
	for i := 0; i < vt.NumField(); i++ {
		r[vt.Field(i).Name] = v.Field(i)
	}
	return r
}
