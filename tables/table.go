//
// Package tables implements immutable tables abstraction
//
package tables

import (
	"fmt"
	"github.com/sudachen/go-fp/fu"
	"github.com/sudachen/go-ml/util"
	"math/bits"
	"reflect"
)

type Raw struct {
	Names   []string
	Columns []reflect.Value
	Na      []util.Bits
	Length  int
}

/*
Table implements column based typed data structure
Every values in a column has the same type.
*/
type Table struct {
	raw  Raw
	cols []*Column
}

/*
Raw returns raw table structure
*/
func (t *Table) Raw() Raw {
	return t.raw
}

/*
Len returns count of table rows
*/
func (t *Table) Len() int {
	return t.raw.Length
}

/*
Names returns list of column names
*/
func (t *Table) Names() []string {
	r := make([]string, len(t.raw.Names), len(t.raw.Names))
	copy(r, t.raw.Names)
	return r
}

/*
Empty creates new empty table
*/
func Empty() *Table {
	t := &Table{}
	return t
}

/*
MakeTable creates ne non-empty table
*/
func MakeTable(names []string, columns []reflect.Value, na []util.Bits, length int) *Table {
	return &Table{
		raw: Raw{
			Names:   names,
			Columns: columns,
			Na:      na,
			Length:  length},
		cols: nil,
	}
}

/*
New creates new Table object

 - from empty list of structs or empty struct
	tables.New([]struct{Name string; Age int; Rate float32}{})
	for empty table.

 - from list of structs
	tables.New([]struct{Name string; Age int; Rate float32}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})

 - from map
	tables.New(map[string]interface{}{"Name":[]string{"Ivanov","Petrov"},"Age":[]int{32,44},"Rate":[]float32{1.2,1.5}})

 - from channel of structs
	type R struct{Name string; Age int; Rate float32}
	c := make(chan R)
	go func(){
		c <- R{"Ivanov",32,1.2}
		c <- R{"Petrov",44,1.5}
		close(c)
	}()
	tables.New(c)
*/
func New(o interface{}) *Table {

	q := reflect.ValueOf(o)

	switch q.Kind() {
	/*case reflect.Ptr: // New(&struct{}{})
			q = q.Elem()
			fallthrough
	  	case reflect.Struct: // New(struct{}{})
			q = reflect.MakeSlice(reflect.SliceOf(q.Type()), 0, 0)
			fallthrough*/
	case reflect.Slice: // New([]struct{}{{}})
		l := q.Len()
		tp := q.Type().Elem()
		if tp.Kind() != reflect.Struct {
			panic("slice of structures allowed only")
		}
		nl := tp.NumField()
		names := make([]string, 0, nl)
		columns := make([]reflect.Value, 0, nl)
		for i := 0; i < nl; i++ {
			fv := tp.Field(i)
			names = append(names, fv.Name)
			col := reflect.MakeSlice(reflect.SliceOf(fv.Type), l, l)
			columns = append(columns, col)
			for j := 0; j < l; j++ {
				col.Index(j).Set(q.Index(j).Field(i))
			}
		}

		return MakeTable(names, columns, make([]util.Bits, len(names)), l)

	case reflect.Chan: // New(chan struct{})
		tp := q.Type().Elem()
		nl := tp.NumField()
		names := make([]string, nl)
		columns := make([]reflect.Value, nl)
		scase := []reflect.SelectCase{{Dir: reflect.SelectRecv, Chan: q}}

		for i := 0; i < nl; i++ {
			fv := tp.Field(i)
			names[i] = fv.Name
			columns[i] = reflect.MakeSlice(reflect.SliceOf(fv.Type), 0, 1)
		}

		length := 0
		for {
			_, v, ok := reflect.Select(scase)
			if !ok {
				break
			}
			for i := 0; i < nl; i++ {
				columns[i] = reflect.Append(columns[i], v.Field(i))
			}
			length++
		}

		return MakeTable(names, columns, make([]util.Bits, len(names)), length)

	case reflect.Map: // New(map[string]interface{}{})
		m := o.(map[string]interface{})
		names := fu.SortedKeysOf(m).([]string)
		columns := make([]reflect.Value, len(names), len(names))
		l := reflect.ValueOf(m[names[0]]).Len()

		for i, n := range names {
			vals := reflect.ValueOf(m[n])
			if vals.Len() != l {
				panic("bad count of elements in column " + n)
			}
			columns[i] = reflect.MakeSlice(vals.Type() /*[]type*/, l, l)
			reflect.Copy(columns[i], vals)
		}

		return MakeTable(names, columns, make([]util.Bits, len(names)), l)
	}

	panic("bad argument type")
}

/*
Slice takes a row slice from table

	t := tables.New([]struct{Name string; Age int; Rate float32}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Slice(0).Row(0) -> {"Ivanov",32,1.2}
	t.Slice(1).Row(0) -> {"Petrov",44,1.5}
	t.Slice(0,2).Len() -> 2
	t.Slice(1,2).Len() -> 1
*/
func (t *Table) Slice(slice ...int) *Table {
	from, to := 0, t.raw.Length
	if len(slice) > 0 {
		from = slice[0]
		to = from + 1
	}
	if len(slice) > 1 {
		to = slice[1]
	}
	rv := make([]reflect.Value, len(t.raw.Columns))
	for i, v := range t.raw.Columns {
		rv[i] = v.Slice(from, to)
	}
	na := make([]util.Bits, len(t.raw.Columns))
	for i, x := range t.raw.Na {
		na[i] = x.Slice(from, to)
	}
	return MakeTable(t.raw.Names, rv, na, to-from)
}

/*
Only takes specified columns as new table

	t := tables.New([]struct{Name string; Age int; Rate float32}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t2 := t.Only("Age","Rate")
	t2.Names() -> ["Age", "Rate"]
	t2.Row(0) -> {"Age": 32, "Rate": 1.2}
*/
func (t *Table) Only(column ...string) *Table {
	rn := make([]string, len(column))
	copy(rn, column)
	rv := make([]reflect.Value, len(column))
	na := make([]util.Bits, len(column))
	for i, v := range t.raw.Columns {
		for j, n := range rn {
			if n == t.raw.Names[i] {
				rv[j] = v.Slice(0, t.raw.Length)
				na[j] = t.raw.Na[i]
			}
		}
	}
	return MakeTable(rn, rv, na, t.raw.Length)
}

/*
Append adds data to table

	t := tables.Empty()

  - from list of structs
	t = t.Append([]struct{Name string; Age int; Rate: float32}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
  - from map of values
	t = t.Append(map[string]interface{}{"Name":[]string{"Ivanov","Petrov"},"Age":[]int{32,44},"Rate":[]float32{1.2,1.5}})

  - from channel
	type R struct{Name string; Age int; Rate float32}
	c := make(chan R)
	go func(){
		c <- R{"Ivanov",32,1.2}
		c <- R{"Petrov",44,1.5}
		close(c)
	}()
	t.Append(c)

Or inserts empty column
  - by empty list of structs
	t = t.Append([]struct{Info string}{})
  - by map of values
	t = t.Append(map[string]interface{}{"Info":[]string{})

*/
func (t *Table) Append(o interface{}) *Table {
	return t.Concat(New(o))
}

/*
Concat concats two tables into new one

	t1 := tables.New(struct{Name string; Age int; Rate float32}{"Ivanov",32,1.2})
	t2 := tables.New(struct{Name string; Age int; Rate float32}{"Petrov",44})
	q := t1.Concat(t2)
	q.Row(0) -> {"Ivanov",32,1.2}
	q.Row(1) -> {"Petrov",44,0}
*/
func (t *Table) Concat(a *Table) *Table {
	names := t.Names()
	columns := make([]reflect.Value, len(names), len(names))
	copy(columns, t.raw.Columns)
	na := make([]util.Bits, len(names))
	copy(na, t.raw.Na)

	for i, n := range a.raw.Names {
		j := util.IndexOf(n, names)
		if j < 0 {
			col := reflect.MakeSlice(a.raw.Columns[i].Type() /*[]type*/, t.raw.Length, t.raw.Length+a.raw.Length)
			col = reflect.AppendSlice(col, a.raw.Columns[i])
			names = append(names, n)
			columns = append(columns, col)
			na = append(na, util.FillBits(t.raw.Length).Append(a.raw.Na[i], t.raw.Length))
		} else {
			columns[j] = reflect.AppendSlice(columns[j], a.raw.Columns[i])
			na[j] = na[j].Append(a.raw.Na[i], t.raw.Length)
		}
	}

	for i, col := range columns {
		if col.Len() < a.raw.Length+t.raw.Length {
			columns[i] = reflect.AppendSlice(
				col,
				reflect.MakeSlice(col.Type(), a.raw.Length, a.raw.Length))
			na[i] = na[i].Append(util.FillBits(a.raw.Length), t.raw.Length)
		}
	}

	return MakeTable(names, columns, na, a.raw.Length+t.raw.Length)
}

/*
Transform transforms table by rows and returns new transformed table

	t := tables.New([]struct{Name string; Age int; Rate float32}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	q := t.Transform(func(r struct{Name string}, i int) struct{Info string}{
				return struct{Info string}{fmt.Sprintf("rec %d for %s", i, r.Name)}
			})
	q.Row(0) -> {Name: "Ivanov", "Age": 32, "Rate", 1.2, "Info": "rec 0 for Ivanov"}
*/
func (t *Table) Transform(f interface{}) *Table {
	l := len(t.raw.Names)
	t2 := t.Map(f)
	names := make([]string, l)
	copy(names, t.raw.Names)
	columns := make([]reflect.Value, l)
	copy(columns, t.raw.Columns)
	na := make([]util.Bits, l)
	copy(na, t.raw.Na)
	for i, n := range t2.raw.Names {
		if j := util.IndexOf(n, names); j >= 0 {
			columns[j] = t2.raw.Columns[i]
			na[j] = util.Bits{}
		} else {
			names = append(names, n)
			columns = append(columns, t2.raw.Columns[i])
			na = append(na, util.Bits{})
		}
	}
	return MakeTable(names, columns, na, t.raw.Length)
}

/*
List executes function for every row

	t := tables.New([]struct{Name string; Age int; Rate float32}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.List(func(r struct{Rate float}, i int){
				fmt.Println(i, r.Rate)
			})
*/
func (t *Table) List(f interface{}) {
	q := reflect.ValueOf(f)
	tp := q.Type().In(0)
	for i := 0; i < t.raw.Length; i++ {
		iv := reflect.ValueOf(i)
		q.Call([]reflect.Value{t.GetRow(i, tp), iv})
	}
}

/*
Filter call predicate for every row and returns new table with passed only rows

	t := tables.New([]struct{Name string; Age int; Rate float32}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	q := t.Filter(func(r struct{Age int}) bool{
				return r.Age > 40
			})
	q.Row(0) -> {Name: "Petrov", "Age": 44, "Rate", 1.5}
*/
func (t *Table) Filter(f interface{}) *Table {
	q := reflect.ValueOf(f)
	tp := q.Type().In(0)
	idxs := make([]int, 0, t.raw.Length)
	for i := 0; i < t.raw.Length; i++ {
		iv := reflect.ValueOf(i)
		rv := q.Call([]reflect.Value{t.GetRow(i, tp), iv})
		if rv[0].Bool() {
			idxs = append(idxs, i)
		}
	}
	return &Table{}
}

/*
Sort sorts rows by specified columns and returns new sorted table

	t := tables.New([]struct{Name string; Age int; Rate float32}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Row(0) -> {Name: "Ivanov", "Age": 32, "Rate", 1.2}
	q := t.Sort("Name",tables.DESC)
	q.Row(0) -> {Name: "Petrov", "Age": 44, "Rate", 1.5}
*/
func (t *Table) Sort(opt interface{}) *Table {
	return nil
}

/*
Reduce groups several rows into one by specified columns or all if no one specified and returns new reduced table

	t := tables.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Ivanov",33,1.3},{"Petrov",44,1.5}})
	t.Len() -> 3
	q := t.Reduce(func(a struct{Age int}, r *struct{Age int}, i int){
				r.Age = func(a,b int)int{ if a.Age >= r.Age {return a.Age} return r.Age }(a,b)
				return
			}, "Name")
	q.Len() -> 2
	// "Name" is grouping field so it's retained, all other fields not presented in result will skipped
	q.Row(0) -> {"Name":"Ivanov", "Age": 33}
	q.Row(1) -> {"Name":"Petrov", "Age": 44}
*/
func (t *Table) Reduce(f interface{}, groupby ...string) *Table {
	return nil
}

/*
Map applies transformation to every row and returns new table containing only transformation results

	t := tables.New([]struct{Name string; Age int; Rate float32}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	q := t.Map(func(r struct{Name string}, i int) struct{Info string}){
				return struct{Info string}{fmt.Sprintf("rec %d for %s", i, r.Name)}
			})
	q.Row(0) -> {"Info": "rec 0 for Ivanov"}
*/
func (t *Table) Map(f interface{}) *Table {
	l := 0
	names := make([]string, l)
	columns := make([]reflect.Value, l)
	na := make([]util.Bits, l)

	return MakeTable(names, columns, na, t.raw.Length)
}

/*
Sink sends all rows to the channel

	t := tables.New([]struct{Name string; Age int; Rate float32}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	c := make(chan struct{Name string})
	go t.Sink(c)
	for x := range c {
		fmr.Println(x.Name)
	}
*/
func (t *Table) Sink(c interface{}) {
}

/*
SinkMap transforms table rows and sends results to the channel

	type R struct{Info string}
	c := make(chan R)
	t.SinkMap(c, func(a struct{Name string; Age int})R{ return R{fmt.Sprint("%s: %d",a.Name,a.Age)} })
	for x := range c {
		fmt.Println(x.Info)
	}
*/
func (t *Table) SinkMap(c interface{}, f interface{}) {
}

/*
 */
func (t *Table) DropNa(names ...string) *Table {
	var dx []int
	if len(names) > 0 {
		dx = make([]int, 0, len(names))
		for _, n := range names {
			i := util.IndexOf(n, t.raw.Names)
			if i < 0 {
				panic("does not have field " + n)
			}
			dx = append(dx, i)
		}
	} else {
		dx = make([]int, len(t.raw.Names))
		for i := range t.raw.Names {
			dx[i] = i
		}
	}
	rc := t.raw.Length
	wc := util.Words(t.raw.Length)
	for j := 0; j < wc; j++ {
		w := uint(0)
		for _, i := range dx {
			w |= t.raw.Na[i].Word(j)
		}
		rc -= bits.OnesCount(w)
	}
	if rc == t.raw.Length {
		return t
	}
	na := make([]util.Bits, len(t.raw.Columns))
	columns := make([]reflect.Value, len(t.raw.Columns))
	for i := range t.raw.Columns {
		columns[i] = reflect.MakeSlice(t.raw.Columns[i].Type(), rc, rc)
	}
	k := 0
	for j := 0; j < t.raw.Length; j++ {
		b := false
		for _, i := range dx {
			b = b || t.raw.Na[i].Bit(j)
		}
		if b {
			continue
		}
		for i := range t.raw.Columns {
			columns[i].Index(k).Set(t.raw.Columns[i].Index(j))
			na[i].Set(k, t.raw.Na[i].Bit(j))
		}
		k++
	}
	return MakeTable(t.raw.Names, columns, na, rc)
}

func (t *Table) FillNa(r interface{}) *Table {
	var m map[string]interface{}
	v := reflect.ValueOf(r)
	if v.Kind() != reflect.Struct && v.Kind() != reflect.Map {
		panic("only struct{...} or map[string]interface{} are allowed as an argument")
	}

	if v.Kind() == reflect.Struct {
		m = map[string]interface{}{}
		n := v.NumField()
		ft := v.Type()
		for i := 0; i < n; i++ {
			f := ft.Field(i)
			m[f.Name] = v.Field(i).Interface()
		}
	} else {
		m = r.(map[string]interface{})
	}

	columns := make([]reflect.Value, len(t.raw.Columns))
	na := make([]util.Bits, len(t.raw.Columns))
	for n, x := range m {
		j := util.IndexOf(n, t.raw.Names)
		if j < 0 {
			panic(" table does not have column " + n)
		}
		if t.raw.Na[j].Len() == 0 {
			columns[j] = t.raw.Columns[j]
		} else {
			y := reflect.ValueOf(x)
			vt := t.raw.Columns[j].Type().Elem()
			if vt != y.Type() {
				if vt.Kind() == reflect.String {
					y = reflect.ValueOf(fmt.Sprint(x))
				} else {
					y = y.Convert(vt)
				}
			}
			columns[j] = reflect.MakeSlice(t.raw.Columns[j].Type(), t.raw.Length, t.raw.Length)
			reflect.Copy(columns[j], t.raw.Columns[j])
			for k := 0; k < t.raw.Length; k++ {
				if t.raw.Na[j].Bit(k) {
					columns[j].Index(k).Set(y)
				}
			}
		}
	}
	for i := range columns {
		if !columns[i].IsValid() {
			columns[i] = t.raw.Columns[i]
			na[i] = t.raw.Na[i]
		}
	}

	return MakeTable(t.raw.Names, columns, na, t.raw.Length)
}
