package tables

import (
	"github.com/sudachen/go-ml/fu"
	"reflect"
)

type Column struct {
	column reflect.Value
	na     fu.Bits
}

/*
Col returns Column wrapper over slice
*/
func Col(a interface{}) *Column {
	v := reflect.ValueOf(a)
	if v.Kind() != reflect.Slice {
		panic("anly slice is allowed as an argument")
	}
	return &Column{v, fu.Bits{}}
}

/*
Text returns column' value converted to string

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Name").Text(0) -> "Ivanov"
	t.Col("Name").Index(0).String() -> "Ivanov"
*/
func (c *Column) Text(row int) string {
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
	return c.ExtractAs(fu.String).([]string)
}

/*
TzeInt returns column' value converted to int

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Age").TzeInt(0) -> 32
	t.Col("Age").Index(0).TzeInt() -> 32
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
	return c.Index(row).Uint()
}

/*
Uint8 returns column' value converted to uint8

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Age").Uint8(0) -> 32
*/
func (c *Column) Uint8(row int) uint8 {
	return c.Index(row).Uint8()
}

/*
Uint16 returns column' value converted to uint16

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Age").Uint16(0) -> 32
*/
func (c *Column) Uint16(row int) uint16 {
	return c.Index(row).Uint16()
}

/*
Uint32 returns column' value converted to uint32

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Age").Uint32(0) -> 32
*/
func (c *Column) Uint32(row int) uint32 {
	return c.Index(row).Uint32()
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
	return c.ExtractAs(fu.Int).([]int)
}

/*
Ints8 extracts column' values as []int8

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Age").Ints8() -> {32,44}
*/
func (c *Column) Ints8() []int8 {
	return c.ExtractAs(fu.Int8).([]int8)
}

/*
Ints16 extracts column' values as []int16

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Age").Ints16() -> {32,44}
*/
func (c *Column) Ints16() []int16 {
	return c.ExtractAs(fu.Int16).([]int16)
}

/*
Ints32 extracts column' values as []int32

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Age").Ints32() -> {32,44}
*/
func (c *Column) Ints32() []int32 {
	return c.ExtractAs(fu.Int32).([]int32)
}

/*
Ints64 extracts column' values as []int64

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Age").Ints64() -> {32,44}
*/
func (c *Column) Ints64() []int64 {
	return c.ExtractAs(fu.Int64).([]int64)
}

/*
Uints extracts column' values as []uint

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Age").Uints() -> {32,44}
*/
func (c *Column) Uints() []uint {
	return c.ExtractAs(fu.Uint).([]uint)
}

/*
Uints8 extracts column' values as []uint8

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Age").Uints8() -> {32,44}
*/
func (c *Column) Uints8() []uint8 {
	return c.ExtractAs(fu.Uint8).([]uint8)
}

/*
Uints16 extracts column' values as []uint16

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Age").Uints16() -> {32,44}
*/
func (c *Column) Uints16() []uint16 {
	return c.ExtractAs(fu.Uint16).([]uint16)
}

/*
Uints32 extracts column' values as []uint32

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Age").Uints32() -> {32,44}
*/
func (c *Column) Uints32() []uint32 {
	return c.ExtractAs(fu.Uint32).([]uint32)
}

/*
Uints64 extracts column' values as []uint64

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Age").Uints64() -> {32,44}
*/
func (c *Column) Uints64() []uint64 {
	return c.ExtractAs(fu.Uint64).([]uint64)
}

/*
Float32 returns column' value converted to float32

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Rate").Float32(0) -> 1.2
*/
func (c *Column) Real(row int) float32 {
	return c.Index(row).Real()
}

/*
Float returns column' value converted to float64

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Rate").Float(0) -> 1.2
*/
func (c *Column) Float(row int) float64 {
	return c.Index(row).Float()
}

/*
Reals extracts column' values as []float32

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Rate").Reals() -> {1.2,1.5}
*/
func (c *Column) Reals() []float32 {
	return c.ExtractAs(fu.Float32).([]float32)
}

/*
Floats extracts column' values as []float64

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Rate").Floats() -> {1.2,1.5}
*/
func (c *Column) Floats() []float64 {
	return c.ExtractAs(fu.Float64).([]float64)
}


/*
Bool returns column' value converted to bool
*/
func (c *Column) Bool(row int) bool {
	return c.Index(row).Bool()
}

/*
Bools extracts column' values as []bool
*/
func (c *Column) Bools() []bool {
	return c.ExtractAs(fu.Bool).([]bool)
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
func (c *Column) ExtractAs(tp reflect.Type, nocopy ...bool) interface{} {
	if c.column.Type().Elem() == tp {
		l := c.column.Len()
		if fu.Fnzb(nocopy...) {
			return c.column.Interface()
		}
		r := reflect.MakeSlice(c.column.Type(), l, l)
		reflect.Copy(r, c.column)
		return r.Interface()
	} else {
		return fu.ConvertSlice(c.column, c.na, tp, nocopy...).Interface()
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
	return &Column{r, fu.Bits{}}
}

/*
Index returns cell with value at specified index

	t := tables.New([]struct{Age int}{{"33"}})
	c := t.Col("Age").Index(0)
	c.String() -> "33"
	c.Float32() -> 33.0
	c.TzeInt() -> 33
*/
func (c *Column) Index(i int) fu.Cell {
	return fu.Cell{c.column.Index(i)}
}

func (c *Column) Value(i int) reflect.Value {
	return c.column.Index(i)
}

/*
Max returns cell with max column' maximal value

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Age").Max().TzeInt() -> 44
	t.Col("Rate").Max().Float32() -> 1.5
*/
func (c *Column) Max() fu.Cell {
	return fu.Cell{fu.MaxValue(c.column)}
}

/*
Min returns cell with column' minimal value

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Age").Min().TzeInt() -> 32
	t.Col("Rate").Min().Float32() -> 1.2
*/
func (c *Column) Min() fu.Cell {
	return fu.Cell{fu.MinValue(c.column)}
}

/*
MaxIndex returns index of first column' maximal value

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Age").MaxIndex() -> 1
*/
func (c *Column) MaxIndex() int {
	return fu.MaxIndex(c.column)
}

/*
MinIndex returns index of first column' minimal value

	t := table.New([]struct{Name string; Age int; Rate float}{{"Ivanov",32,1.2},{"Petrov",44,1.5}})
	t.Col("Age").MinIndex() -> 0
*/
func (c *Column) MinIndex() int {
	return fu.MinIndex(c.column)
}

/*
Raw returns column internals
*/
func (c *Column) Raw() (reflect.Value, fu.Bits) {
	return c.column, c.na
}

func (c *Column) IsFloat() bool {
	t := c.Type()
	return t.Kind() == reflect.Float32 || t.Kind() == reflect.Float64
}
