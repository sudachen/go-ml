package tables

import (
	"github.com/sudachen/go-ml/mlutil"
	"reflect"
)

/*
Row returns table row as a map of reflect.Value
	t := tables.New([]struct{Name string; Age int; Rate float32}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Row(0)["Name"].String() -> "Ivanov"
*/
func (t *Table) Row(row int) map[string]reflect.Value {
	r := map[string]reflect.Value{}
	for i, n := range t.raw.Names {
		// prevent to change value in slice via returned reflect.Value
		r[n] = reflect.ValueOf(t.raw.Columns[i].Index(row).Interface())
	}
	return r
}

/*
FillRow fills row as a struct
*/
func (t *Table) FillRow(i int, tp reflect.Type, p reflect.Value) {
	v := p.Elem()
	fl := tp.NumField()
	for fi := 0; fi < fl; fi++ {
		n := tp.Field(fi).Name
		j := mlutil.IndexOf(n, t.raw.Names)
		if j < 0 {
			panic("table does not have field " + n)
		}
		v.Field(fi).Set(t.raw.Columns[j].Index(i))
	}
}

/*
GetRow gets row as a struct wrapped by reflect.Value
*/
func (t *Table) GetRow(i int, tp reflect.Type) reflect.Value {
	v := reflect.New(tp)
	t.FillRow(i, tp, v)
	return v.Elem()
}

/*
Fetch fills struct with table' row data

	t := tables.New([]struct{Name string; Age int; Rate float32}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	r := struct{Name string; Age int}{}
	t.Fetch(0,&r)
	r.Name -> "Ivanov"
	r.Age -> 32
*/
func (t *Table) Fetch(i int, r interface{}) {
	q := reflect.ValueOf(r)
	if q.Kind() != reflect.Ptr || q.Elem().Kind() != reflect.Struct {
		panic("only &struct{...} is allowed as an argumet")
	}
	t.FillRow(i, q.Elem().Type(), q)
}
