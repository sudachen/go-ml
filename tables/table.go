//
// Package tables implements immutable tables abstraction
//
package tables

import (
	"fmt"
	"github.com/sudachen/go-foo/fu"
	"github.com/sudachen/go-foo/lazy"
	"github.com/sudachen/go-ml/mlutil"
	"golang.org/x/xerrors"
	"math/bits"
	"reflect"
)

/*
Table implements column based typed data structure
Every values in a column has the same type.
*/
type Table struct{ raw Raw }

/*
Raw is the row presentation of a table, can be accessed by the Table.Raw method
*/
type Raw struct {
	Names   []string
	Columns []reflect.Value
	Na      []mlutil.Bits
	Length  int
}

/*
IsLazy returns false, because Table is not lazy
it's the tables.AnyData interface implementation
*/
func (*Table) IsLazy() bool { return false }

/*
Table returns self, because Table is a table
it's the tables.AnyData interface implementation
*/
func (t *Table) Table() *Table { return t }

/*
Collect returns self, because Table is a table
it's the tables.AnyData interface implementation
*/
func (t *Table) Collect() (*Table, error) { return t, nil }

/*
Lazy returns new lazy stream sourcing from the table
it's the tables.AnyData interface implementation
*/
func (t *Table) Lazy() Lazy {
	return func() lazy.Stream {
		flag := &lazy.AtomicFlag{Value: 1}
		return func(index uint64) (v reflect.Value, err error) {
			if index == lazy.STOP {
				flag.Clear()
			} else if flag.State() && index < uint64(t.raw.Length) {
				return reflect.ValueOf(t.Index(int(index))), nil
			}
			return reflect.ValueOf(false), nil
		}
	}
}

/*
Col returns Column object for the table' column selected by the name

	t := tables.New([]struct{Name string; Age int; Rate float32}{{"Ivanov",42,1.2},{"Petrov",42,1.5}})
	t.Col("Name").String(0) -> "Ivanov"
	t.Col("Name").Len() -> 2
*/
func (t *Table) Col(c string) *Column {
	for i, n := range t.raw.Names {
		if n == c {
			return &Column{t.raw.Columns[i], t.raw.Na[i]}
		}
	}
	panic("there is not column with name " + c)
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

func (t *Table) FilteredLen(f func(int) bool) int {
	if f != nil {
		L := 0
		for i := 0; i < t.raw.Length; i++ {
			if f(i) {
				L++
			}
		}
		return L
	}
	return t.Len()
}

/*
Names returns list of column Names
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
func MakeTable(names []string, columns []reflect.Value, na []mlutil.Bits, length int) *Table {
	return &Table{
		raw: Raw{
			Names:   names,
			Columns: columns,
			Na:      na,
			Length:  length},
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
	case reflect.Slice: // New([]struct{}{{}})
		l := q.Len()
		tp := q.Type().Elem()
		if tp.Kind() != reflect.Struct && !(tp.Kind() == reflect.Ptr && tp.Elem().Kind() == reflect.Struct) {
			panic("slice of structures allowed only")
		}

		if l > 0 && (tp == mlutil.StructType) {
			// New([]mlutil.Struct{{}})
			lrx := q.Interface().([]mlutil.Struct)
			names := lrx[0].Names
			columns := make([]reflect.Value, len(names))
			for i := range columns {
				columns[i] = reflect.MakeSlice(reflect.SliceOf(lrx[0].Columns[i].Type()), l, l)
			}
			na := make([]mlutil.Bits, len(names))
			for i := range names {
				for j := 0; j < l; j++ {
					columns[i].Index(j).Set(lrx[j].Columns[i])
					na[i].Set(j, lrx[j].Na.Bit(i))
				}
			}
			return MakeTable(names, columns, make([]mlutil.Bits, len(names)), l)
		}

		if l > 0 && (tp.Kind() == reflect.Ptr && tp.Elem() == mlutil.StructType) {
			// New([]*mlutil.Struct{{}})
			lrx := q.Interface().([]*mlutil.Struct)
			names := lrx[0].Names
			columns := make([]reflect.Value, len(names))
			for i := range columns {
				columns[i] = reflect.MakeSlice(reflect.SliceOf(lrx[0].Columns[i].Type()), l, l)
			}
			na := make([]mlutil.Bits, len(names))
			for i := range names {
				for j := 0; j < l; j++ {
					columns[i].Index(j).Set(lrx[j].Columns[i])
					na[i].Set(j, lrx[j].Na.Bit(i))
				}
			}
			return MakeTable(names, columns, make([]mlutil.Bits, len(names)), l)
		}

		if tp.Kind() == reflect.Ptr {
			tp = tp.Elem()
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
				x := q.Index(j)
				if x.Kind() == reflect.Ptr {
					x = x.Elem()
				}
				col.Index(j).Set(x.Field(i))
			}
		}

		return MakeTable(names, columns, make([]mlutil.Bits, len(names)), l)

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

		return MakeTable(names, columns, make([]mlutil.Bits, len(names)), length)

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

		return MakeTable(names, columns, make([]mlutil.Bits, len(names)), l)
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
	to, from = fu.Mini(to, t.raw.Length), fu.Mini(from, t.raw.Length)
	rv := make([]reflect.Value, len(t.raw.Columns))
	for i, v := range t.raw.Columns {
		rv[i] = v.Slice(from, to)
	}
	na := make([]mlutil.Bits, len(t.raw.Columns))
	for i, x := range t.raw.Na {
		na[i] = x.Slice(from, to)
	}
	return MakeTable(t.raw.Names, rv, na, to-from)
}

/*
Only takes specified Columns as new table

	t := tables.New([]struct{Name string; Age int; Rate float32}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t2 := t.Only("Age","Rate")
	t2.Names() -> ["Age", "Rate"]
	t2.Row(0) -> {"Age": 32, "Rate": 1.2}
*/
func (t *Table) Only(column ...string) *Table {
	rn := mlutil.Bits{}
	for _, c := range column {
		like := mlutil.Pattern(c)
		for i, x := range t.raw.Names {
			if like(x) {
				rn.Set(i, true)
			}
		}
	}
	names := make([]string, 0, rn.Count())
	for i := range t.raw.Names {
		if rn.Bit(i) {
			names = append(names, t.raw.Names[i])
		}
	}
	rv := make([]reflect.Value, 0, len(names))
	na := make([]mlutil.Bits, 0, len(names))
	for i, v := range t.raw.Columns {
		if rn.Bit(i) {
			rv = append(rv, v.Slice(0, t.raw.Length))
			na = append(na, t.raw.Na[i])
		}
	}
	return MakeTable(names, rv, na, t.raw.Length)
}

func (t *Table) Except(column ...string) *Table {
	rn := mlutil.Bits{}
	for _, c := range column {
		like := mlutil.Pattern(c)
		for i, x := range t.raw.Names {
			if like(x) {
				rn.Set(i, true)
			}
		}
	}
	names := make([]string, 0, len(t.raw.Names)-rn.Count())
	for i := range t.raw.Names {
		if !rn.Bit(i) {
			names = append(names, t.raw.Names[i])
		}
	}
	rv := make([]reflect.Value, 0, len(names))
	na := make([]mlutil.Bits, 0, len(names))
	for i, v := range t.raw.Columns {
		if !rn.Bit(i) {
			rv = append(rv, v.Slice(0, t.raw.Length))
			na = append(na, t.raw.Na[i])
		}
	}
	return MakeTable(names, rv, na, t.raw.Length)
}

func (t *Table) OnlyNames(column ...string) []string {
	rn := mlutil.Bits{}
	for _, c := range column {
		like := mlutil.Pattern(c)
		for i, x := range t.raw.Names {
			if like(x) {
				rn.Set(i, true)
			}
		}
	}
	names := make([]string, 0, rn.Count())
	for i := range t.raw.Names {
		if rn.Bit(i) {
			names = append(names, t.raw.Names[i])
		}
	}
	return names
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
	na := make([]mlutil.Bits, len(names))
	copy(na, t.raw.Na)

	for i, n := range a.raw.Names {
		j := fu.IndexOf(n, names)
		if j < 0 {
			col := reflect.MakeSlice(a.raw.Columns[i].Type() /*[]type*/, t.raw.Length, t.raw.Length+a.raw.Length)
			col = reflect.AppendSlice(col, a.raw.Columns[i])
			names = append(names, n)
			columns = append(columns, col)
			na = append(na, mlutil.FillBits(t.raw.Length).Append(a.raw.Na[i], t.raw.Length))
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
			na[i] = na[i].Append(mlutil.FillBits(a.raw.Length), t.raw.Length)
		}
	}

	return MakeTable(names, columns, na, a.raw.Length+t.raw.Length)
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
Sort sorts rows by specified Columns and returns new sorted table

	t := tables.New([]struct{Name string; Age int; Rate float32}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Row(0) -> {Name: "Ivanov", "Age": 32, "Rate", 1.2}
	q := t.Sort("Name",tables.DESC)
	q.Row(0) -> {Name: "Petrov", "Age": 44, "Rate", 1.5}
*/
func (t *Table) Sort(opt interface{}) *Table {
	return nil
}

/*
 */
func (t *Table) DropNa(names ...string) *Table {
	var dx []int
	if len(names) > 0 {
		dx = make([]int, 0, len(names))
		for _, n := range names {
			i := fu.IndexOf(n, t.raw.Names)
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
	wc := mlutil.Words(t.raw.Length)
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
	na := make([]mlutil.Bits, len(t.raw.Columns))
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
	na := make([]mlutil.Bits, len(t.raw.Columns))
	for n, x := range m {
		j := fu.IndexOf(n, t.raw.Names)
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

func (t *Table) Index(r int) mlutil.Struct {
	names := t.raw.Names
	columns := make([]reflect.Value, len(names))
	na := mlutil.Bits{}
	for i, c := range t.raw.Columns {
		columns[i] = c.Index(r)
		na.Set(i, t.raw.Na[i].Bit(r))
	}
	return mlutil.Struct{names, columns, na}
}

func (t *Table) Last() mlutil.Struct {
	return t.Index(t.Len() - 1)
}

func (t *Table) With(column *Column, name string) *Table {
	if column.Len() != t.raw.Length {
		panic(xerrors.Errorf("column length is not match table length"))
	}
	if fu.IndexOf(name, t.raw.Names) >= 0 {
		panic(xerrors.Errorf("column with name `%v` alreday exists in the table", name))
	}
	names := make([]string, len(t.raw.Names), len(t.raw.Names)+1)
	columns := make([]reflect.Value, len(t.raw.Names), len(t.raw.Names)+1)
	na := make([]mlutil.Bits, len(t.raw.Names), len(t.raw.Names)+1)
	copy(names, t.raw.Names)
	copy(columns, t.raw.Columns)
	copy(na, t.raw.Na)
	names = append(names, name)
	columns = append(columns, column.column)
	na = append(na, column.na)
	return MakeTable(names, columns, na, t.raw.Length)
}

func (t *Table) Round(prec int /*, columns ...string*/) *Table {
	return t.Lazy().Round(prec).LuckyCollect()
}

/*
func Shape(x interface{}, names ...string) *Table {
	v := reflect.ValueOf(x)
	vt := v.Type()
	if v.Kind() != reflect.Slice {
		panic(xerrors.Errorf("only []any allowed as the first argument"))
	}
	width := len(names)
	length := v.Len() / width
	columns := make([]reflect.Value, width)
	na := make([]mlutil.Bits, width)
	for i := range names {
		columns[i] = reflect.MakeSlice(reflect.SliceOf(vt), length, length)
		for j := 0; j < length; j++ {
			y := v.Index(j*width + i)
			na[i].Set(j, mlutil.Isna(y))
			columns[i].Index(j).Set(y)
		}
	}
	return MakeTable(names, columns, na, length)
}

func Shape32f(v []float32, names ...string) *Table {
	width := len(names)
	length := len(v) / width
	columns := make([]reflect.Value, width)
	na := make([]mlutil.Bits, width)
	for i := range names {
		ls := make([]float32, length)
		for j := 0; j < length; j++ {
			y := v[j*width+i]
			na[i].Set(j, math.IsNaN(float64(y)))
			ls[j] = y
		}
		columns[i] = reflect.ValueOf(ls)
	}
	return MakeTable(names, columns, na, length)
}
*/
