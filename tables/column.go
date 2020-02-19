package tables

import (
	"github.com/sudachen/go-ml/internal"
	"github.com/sudachen/go-ml/util"
	"reflect"
)

type Column struct {
	column reflect.Value
	na     util.Bits
}

/*
Col returns Column wrapper over slice
*/
func Col(a interface{}) *Column {
	v := reflect.ValueOf(a)
	if v.Kind() != reflect.Slice {
		panic("anly slice is allowed as an argument")
	}
	return &Column{v, util.Bits{}}
}

/*
Col returns Column object for the table' column selected by the name

	t := tables.New([]struct{Name string; Age int; Rate float32}{{"Ivanov",42,1.2},{"Petrov",42,1.5}})
	t.Col("Name").String(0) -> "Ivanov"
	t.Col("Name").Len() -> 2
*/
func (t *Table) Col(column string) *Column {
	for i, n := range t.raw.Names {
		if n == column {
			if t.cols == nil {
				t.cols = make([]*Column, len(t.raw.Names))
			}
			if t.cols[i] == nil {
				c := &Column{t.raw.Columns[i], t.raw.Na[i]}
				t.cols[i] = c
			}
			return t.cols[i]
		}
	}
	panic("there is not column with name " + column)
}

/*
String returns column' value converted to string

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Name").String(0) -> "Ivanov"
	t.Col("Name").Index(0).String() -> "Ivanov"
*/
func (c *Column) String(row int) string {
	return c.Index(row).String()
}

func (c *Column) Na(i int) bool {
	return c.na.Bit(i)
}

/*
Strings extracts column' values as []string

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Name").Strings() -> {"Ivanov","Petrow"}
*/
func (c *Column) Strings() []string {
	return c.ExtractAs(internal.StringType).([]string)
}

/*
Int returns column' value converted to int

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Age").Int(0) -> 32
	t.Col("Age").Index(0).Int() -> 32
*/
func (c *Column) Int(row int) int {
	return c.Index(row).Int()
}

/*
Int8 returns column' value converted to int8

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Age").Int8(0) -> 32
	t.Col("Age").Index().Int8() -> 32
*/
func (c *Column) Int8(row int) int8 {
	return c.Index(row).Int8()
}

/*
Int16 returns column' value converted to int16

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Age").Int16(0) -> 32
	t.Col("Age").Index().Int16() -> 32
*/
func (c *Column) Int16(row int) int16 {
	return c.Index(row).Int16()
}

/*
Int32 returns column' value converted to int32

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Age").Int32(0) -> 32
	t.Col("Age").Index(0).Int32() -> 32
*/
func (c *Column) Int32(row int) int32 {
	return c.Index(row).Int32()
}

/*
Int64 returns column' value converted to int64

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Age").Int64(0) -> 32
	t.Col("Age").Index(0).Int64() -> 32
*/
func (c *Column) Int64(row int) int64 {
	return c.Index(row).Int64()
}

/*
Uint returns column' value converted to uint

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Age").Uint(0) -> 32
*/
func (c *Column) Uint(row int) uint {
	v := c.column.Index(row)
	return util.Convert(v, internal.UintType).(uint)
}

/*
Uint8 returns column' value converted to uint8

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Age").Uint8(0) -> 32
*/
func (c *Column) Uint8(row int) uint8 {
	v := c.column.Index(row)
	return util.Convert(v, internal.Uint8Type).(uint8)
}

/*
Uint16 returns column' value converted to uint16

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Age").Uint16(0) -> 32
*/
func (c *Column) Uint16(row int) uint16 {
	v := c.column.Index(row)
	return util.Convert(v, internal.Uint16Type).(uint16)
}

/*
Uint32 returns column' value converted to uint32

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Age").Uint32(0) -> 32
*/
func (c *Column) Uint32(row int) uint32 {
	v := c.column.Index(row)
	return util.Convert(v, internal.Uint32Type).(uint32)
}

/*
Uint64 returns column' value converted to uint64

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Age").Uint64(0) -> 32
	t.Col("Age").Index(0).Uint64() -> 32
*/
func (c *Column) Uint64(row int) uint64 {
	return c.Index(row).Uint64()
}

/*
Ints extracts column' values as []int

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Age").Ints() -> {32,44}
*/
func (c *Column) Ints() []int {
	return c.ExtractAs(internal.IntType).([]int)
}

/*
Ints8 extracts column' values as []int8

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Age").Ints8() -> {32,44}
*/
func (c *Column) Ints8() []int8 {
	return c.ExtractAs(internal.Int8Type).([]int8)
}

/*
Ints16 extracts column' values as []int16

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Age").Ints16() -> {32,44}
*/
func (c *Column) Ints16() []int16 {
	return c.ExtractAs(internal.Int16Type).([]int16)
}

/*
Ints32 extracts column' values as []int32

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Age").Ints32() -> {32,44}
*/
func (c *Column) Ints32() []int32 {
	return c.ExtractAs(internal.Int32Type).([]int32)
}

/*
Ints64 extracts column' values as []int64

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Age").Ints64() -> {32,44}
*/
func (c *Column) Ints64() []int64 {
	return c.ExtractAs(internal.Int64Type).([]int64)
}

/*
Uints extracts column' values as []uint

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Age").Uints() -> {32,44}
*/
func (c *Column) Uints() []uint {
	return c.ExtractAs(internal.UintType).([]uint)
}

/*
Uints8 extracts column' values as []uint8

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Age").Uints8() -> {32,44}
*/
func (c *Column) Uints8() []uint8 {
	return c.ExtractAs(internal.Uint8Type).([]uint8)
}

/*
Uints16 extracts column' values as []uint16

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Age").Uints16() -> {32,44}
*/
func (c *Column) Uints16() []uint16 {
	return c.ExtractAs(internal.Uint16Type).([]uint16)
}

/*
Uints32 extracts column' values as []uint32

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Age").Uints32() -> {32,44}
*/
func (c *Column) Uints32() []uint32 {
	return c.ExtractAs(internal.Uint32Type).([]uint32)
}

/*
Uints64 extracts column' values as []uint64

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Age").Uints64() -> {32,44}
*/
func (c *Column) Uints64() []uint64 {
	return c.ExtractAs(internal.Uint64Type).([]uint64)
}

/*
Float returns column' value converted to float32

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Rate").Float(0) -> 1.2
*/
func (c *Column) Float(row int) float32 {
	return c.Index(row).Float()
}

/*
Float64 returns column' value converted to float64

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Rate").Float64(0) -> 1.2
*/
func (c *Column) Float64(row int) float64 {
	return c.Index(row).Float64()
}

/*
Floats extracts column' values as []float32

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Rate").Floats() -> {1.2,1.5}
*/
func (c *Column) Floats() []float32 {
	return c.ExtractAs(internal.FloatType).([]float32)
}

/*
Floats64 extracts column' values as []float64

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Rate").Floats64() -> {1.2,1.5}
*/
func (c *Column) Floats64() []float64 {
	return c.ExtractAs(internal.Float64Type).([]float64)
}

/*
Interface returns column' value as is

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Rate").Interface(0).(float32) -> 1.2
	t.Col("Rate").Index(0).Interface().(float32) -> 1.2
*/
func (c *Column) Interface(row int) interface{} {
	return c.Index(row).Interface()
}

/*
ExtractAs extracts values as array with specified type

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Age").ExtractAs(reflect.TypeOf("")).([]string)[0] -> "32"
	t.Col("Rate").ExtractAs(reflect.TypeOf(int(0))).([]int)[0] -> 1
*/
func (c *Column) ExtractAs(tp reflect.Type) interface{} {
	if c.column.Type().Elem() == tp {
		l := c.column.Len()
		r := reflect.MakeSlice(c.column.Type(), l, l)
		reflect.Copy(r, c.column)
		return r.Interface()
	} else {
		return util.Convert(c.column, tp)
	}
}

/*
Inspect returns raw array of column's values

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Name").Inspect().([]string)[0] -> "Ivanov"
	t.Col("Age").Inspect().([]int)[0] -> 32
	t.Col("Rate").Inspect().([]float32)[0] -> 1.2
*/
func (c *Column) Inspect() interface{} {
	return c.column.Interface()
}

/*
Type returns (reflect) type of column' values
*/
func (c *Column) Type() reflect.Type {
	return c.column.Type().Elem()
}

/*
Len returns length of column

	t := tables.New([]struct{Name string}{{"Ivanov"}})
	c1 := t.Col("Name")
	t.Append([]struct{Name string}{{"Petrov"}})
	c2 := t.Col("Name")
	c1.Len() -> 1
	c2.Len() -> 2
*/
func (c *Column) Len() int {
	return c.column.Len()
}

/*
Unique returns column with only unique values

	t := tables.New([]struct{Name string}{{"Ivanov"}})
	u1 := t.Col("Name").Unique()
	t = t.Append([]struct{Name string}{{"Petrov"},{"Petrov"}})
	u2 := t.Col("Name").Unique()
	u1.Unique().Inspect() -> {}
	u2.Unique().Len() -> 2
*/
func (c *Column) Unique() *Column {
	v := reflect.ValueOf(true)
	m := reflect.MakeMap(reflect.MapOf(c.column.Type().Elem(), v.Type()))
	r := reflect.MakeSlice(c.column.Type(), 0, 0)
	for i := 0; i < c.column.Len(); i++ {
		x := c.column.Index(i)
		q := m.MapIndex(x)
		if !q.IsValid() {
			r = reflect.Append(r, x)
			m.SetMapIndex(x, v)
		}
	}
	return &Column{r, util.Bits{}}
}

/*
Index returns cell with value at specified index

	t := tables.New([]struct{Age int}{{"33"}})
	c := t.Col("Age").Index(0)
	c.String() -> "33"
	c.Float() -> 33.0
	c.Int() -> 33
*/
func (c *Column) Index(i int) Cell {
	return Cell{c.column.Index(i)}
}

/*
Max returns cell with max column' maximal value

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Age").Max().Int() -> 44
	t.Col("Rate").Max().Float() -> 1.5
*/
func (c *Column) Max() Cell {
	return Cell{util.Max(c.column)}
}

/*
Min returns cell with column' minimal value

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Age").Min().Int() -> 32
	t.Col("Rate").Min().Float() -> 1.2
*/
func (c *Column) Min() Cell {
	return Cell{util.Min(c.column)}
}

/*
MaxIndex returns index of first column' maximal value

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Age").MaxIndex() -> 1
*/
func (c *Column) MaxIndex() int {
	return util.MaxIndex(c.column)
}

/*
MinIndex returns index of first column' minimal value

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Age").MinIndex() -> 0
*/
func (c *Column) MinIndex() int {
	return util.MinIndex(c.column)
}

/*
Raw returns column internals
*/
func (c *Column) Raw() (reflect.Value, util.Bits) {
	return c.column, c.na
}
