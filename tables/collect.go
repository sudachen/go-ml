package tables

import "reflect"

func (t *Table) Collect(s interface{}) interface{} {
	tp := reflect.TypeOf(s)
	if tp.Kind() == reflect.Ptr {
		tp = tp.Elem()
	}
	if tp.Kind() != reflect.Struct {
		panic("only struct{...} is allowed as an argument")
	}
	r := reflect.MakeSlice(reflect.SliceOf(tp), t.raw.Length, t.raw.Length)
	for i := 0; i < t.raw.Length; i++ {
		t.FillRow(i, tp, r.Index(i).Addr())
	}
	return r.Interface()
}
