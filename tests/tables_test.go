package tests

import (
	"fmt"
	"github.com/sudachen/go-fp/lazy"
	"github.com/sudachen/go-ml/tables"
	"github.com/sudachen/go-ml/util"
	"gotest.tools/assert"
	"gotest.tools/assert/cmp"
	"math/rand"
	"reflect"
	"testing"
	"time"
)

func Test_New0(t *testing.T) {
	q := tables.New([]struct {
		Name string
		Age  int
		Rate float32
	}{})
	assert.DeepEqual(t, q.Names(), []string{"Name", "Age", "Rate"})
	assert.Assert(t, q.Len() == 0)
}

func Test_New1(t *testing.T) {
	q := tables.New([]struct {
		Name string
		Age  int
		Rate float32
	}{{"Ivanov", 32, 1.2}})
	assert.DeepEqual(t, q.Names(), []string{"Name", "Age", "Rate"})
	assert.Assert(t, q.Len() == 1)
	assert.DeepEqual(t, util.MapInterface(q.Row(0)),
		map[string]interface{}{
			"Name": "Ivanov",
			"Age":  32,
			"Rate": float32(1.2),
		})
}

func Test_New2(t *testing.T) {
	q := tables.New([]struct {
		Name string
		Age  int
		Rate float32
	}{
		{"Ivanov", 32, 1.2},
		{"Petrov", 44, 1.5}})
	assert.DeepEqual(t, q.Names(), []string{"Name", "Age", "Rate"})
	assert.Assert(t, q.Len() == 2)
	assert.DeepEqual(t, util.MapInterface(q.Row(0)),
		map[string]interface{}{
			"Name": "Ivanov",
			"Age":  32,
			"Rate": float32(1.2),
		})
	assert.DeepEqual(t, util.MapInterface(q.Row(1)),
		map[string]interface{}{
			"Name": "Petrov",
			"Age":  44,
			"Rate": float32(1.5),
		})
}

func Test_New3(t *testing.T) {
	q := tables.New(map[string]interface{}{
		"Name": []string{"Ivanov", "Petrov"},
		"Age":  []int{32, 44},
		"Rate": []float32{1.2, 1.5}})
	assert.DeepEqual(t, q.Names(), []string{"Age", "Name", "Rate"})
	assert.Assert(t, q.Len() == 2)
	assert.DeepEqual(t, util.MapInterface(q.Row(0)),
		map[string]interface{}{
			"Name": "Ivanov",
			"Age":  32,
			"Rate": float32(1.2),
		})
	assert.DeepEqual(t, util.MapInterface(q.Row(1)),
		map[string]interface{}{
			"Name": "Petrov",
			"Age":  44,
			"Rate": float32(1.5),
		})
}

func Test_New4(t *testing.T) {
	type R struct {
		Name string
		Age  int
		Rate float32
	}
	c := make(chan R)
	go func() {
		c <- R{"Ivanov", 32, 1.2}
		c <- R{"Petrov", 44, 1.5}
		close(c)
	}()
	q := tables.New(c)
	assert.DeepEqual(t, q.Names(), []string{"Name", "Age", "Rate"})
	assert.Assert(t, q.Len() == 2)
	assert.DeepEqual(t, util.MapInterface(q.Row(0)),
		map[string]interface{}{
			"Name": "Ivanov",
			"Age":  32,
			"Rate": float32(1.2),
		})
	assert.DeepEqual(t, util.MapInterface(q.Row(1)),
		map[string]interface{}{
			"Name": "Petrov",
			"Age":  44,
			"Rate": float32(1.5),
		})
}

func Test_Row1(t *testing.T) {
	q := TrTable()
	r := TR{}
	for i, v := range trList {
		q.Fetch(i, &r)
		assert.DeepEqual(t, r, v)
	}
}

func Test_Row2(t *testing.T) {
	q := TrTable()
	r := struct{ A int }{}
	assert.Assert(t, cmp.Panics(func() {
		q.Fetch(0, &r)
	}))
	x := map[int]interface{}{}
	assert.Assert(t, cmp.Panics(func() {
		q.Fetch(0, &x)
	}))
}

func Test_Append0(t *testing.T) {
	q := tables.Empty()
	assert.Assert(t, cmp.Panics(func() { q.Append([]int{0}) }))
	assert.Assert(t, cmp.Panics(func() { q.Append(0) }))
	assert.Assert(t, cmp.Panics(func() {
		q.Append(map[string]interface{}{
			"Name": []string{"a", "b"},
			"Age":  []int{0},
		})
	}))
	q2 := q.Append([]struct{ Name string }{})
	assert.Assert(t, q.Len() == q2.Len())
	assert.Assert(t, cmp.Panics(func() { q2.Append(struct{ Name int }{0}) }))
	assert.Assert(t, q.Append([]struct{ Age int }{{0}}).Len() == q.Len()+1)
	assert.Assert(t, q.Append([]struct{ Tall int }{{0}}).Len() == q.Len()+1)
}

func Test_Collect(t *testing.T) {
	q := tables.New(trList)
	assertTrData(t, q)
	r := q.Collect(TR{}).([]TR)
	assert.DeepEqual(t, trList, r)
	r = q.Collect(&TR{}).([]TR)
	assert.DeepEqual(t, trList, r)
	assert.Assert(t, cmp.Panics(func() {
		r = q.Collect(false).([]TR)
	}))
}

func Test_ColumnString(t *testing.T) {
	q := PrepareTable(t)

	assert.DeepEqual(t, q.Col("Name").String(0), "Ivanov")
	assert.DeepEqual(t, q.Col("Name").String(1), "Petrov")
	assert.DeepEqual(t, q.Col("Age").String(0), "32")
	assert.DeepEqual(t, q.Col("Age").String(1), "44")
	assert.DeepEqual(t, q.Col("Rate").String(0), "1.2")
	assert.DeepEqual(t, q.Col("Rate").String(1), "1.5")

	assert.Assert(t, cmp.Panics(func() { q.Col("name") }))
}

func Test_ColumnStrings(t *testing.T) {
	q := PrepareTable(t)

	assert.DeepEqual(t, q.Col("Name").Strings(), []string{"Ivanov", "Petrov"})
	assert.DeepEqual(t, q.Col("Age").Strings(), []string{"32", "44"})
	assert.DeepEqual(t, q.Col("Rate").Strings(), []string{"1.2", "1.5"})
}

func Test_ColumnInt(t *testing.T) {
	q := PrepareTable(t)

	assert.DeepEqual(t, q.Col("Age").Int(0), 32)
	assert.DeepEqual(t, q.Col("Age").Int(1), 44)
	assert.DeepEqual(t, q.Col("Rate").Int(0), 1)
	assert.DeepEqual(t, q.Col("Rate").Int(1), 1)

	assert.DeepEqual(t, q.Col("Age").Int8(0), int8(32))
	assert.DeepEqual(t, q.Col("Rate").Int8(0), int8(1))

	assert.DeepEqual(t, q.Col("Age").Int16(0), int16(32))
	assert.DeepEqual(t, q.Col("Rate").Int16(0), int16(1))

	assert.DeepEqual(t, q.Col("Age").Int32(0), int32(32))
	assert.DeepEqual(t, q.Col("Rate").Int32(0), int32(1))

	assert.DeepEqual(t, q.Col("Age").Int64(0), int64(32))
	assert.DeepEqual(t, q.Col("Rate").Int64(0), int64(1))

	assert.DeepEqual(t, q.Col("Age").Uint(0), uint(32))
	assert.DeepEqual(t, q.Col("Rate").Uint(0), uint(1))

	assert.DeepEqual(t, q.Col("Age").Uint8(0), uint8(32))
	assert.DeepEqual(t, q.Col("Rate").Uint8(0), uint8(1))

	assert.DeepEqual(t, q.Col("Age").Uint16(0), uint16(32))
	assert.DeepEqual(t, q.Col("Rate").Uint16(0), uint16(1))

	assert.DeepEqual(t, q.Col("Age").Uint32(0), uint32(32))
	assert.DeepEqual(t, q.Col("Rate").Uint32(0), uint32(1))

	assert.DeepEqual(t, q.Col("Age").Uint64(0), uint64(32))
	assert.DeepEqual(t, q.Col("Rate").Uint64(0), uint64(1))

	assert.Assert(t, cmp.Panics(func() { q.Col("age") }))
}

func Test_ColumnInts(t *testing.T) {
	q := PrepareTable(t)

	assert.DeepEqual(t, q.Col("Age").Ints(), []int{32, 44})
	assert.DeepEqual(t, q.Col("Age").Ints8(), []int8{32, 44})
	assert.DeepEqual(t, q.Col("Age").Ints16(), []int16{32, 44})
	assert.DeepEqual(t, q.Col("Age").Ints32(), []int32{32, 44})
	assert.DeepEqual(t, q.Col("Age").Ints64(), []int64{32, 44})
	assert.DeepEqual(t, q.Col("Age").Uints(), []uint{32, 44})
	assert.DeepEqual(t, q.Col("Age").Uints8(), []uint8{32, 44})
	assert.DeepEqual(t, q.Col("Age").Uints16(), []uint16{32, 44})
	assert.DeepEqual(t, q.Col("Age").Uints32(), []uint32{32, 44})
	assert.DeepEqual(t, q.Col("Age").Uints64(), []uint64{32, 44})
}

func Test_ColumnInt2(t *testing.T) {
	q := PrepareTable(t)

	c := q.Col("Age")
	assert.Assert(t, c.Index(0).Int() == 32)
	assert.Assert(t, c.Index(0).Int8() == 32)
	assert.Assert(t, c.Index(0).Int16() == 32)
	assert.Assert(t, c.Index(0).Int32() == 32)
	assert.Assert(t, c.Index(0).Int64() == 32)
	assert.Assert(t, c.Index(0).Uint() == 32)
	assert.Assert(t, c.Index(0).Uint8() == 32)
	assert.Assert(t, c.Index(0).Uint16() == 32)
	assert.Assert(t, c.Index(0).Uint32() == 32)
	assert.Assert(t, c.Index(0).Uint64() == 32)
}

func Test_ColumnFloat(t *testing.T) {
	q := PrepareTable(t)

	assert.DeepEqual(t, q.Col("Rate").Float(0), float32(1.2))
	assert.DeepEqual(t, q.Col("Rate").Float(1), float32(1.5))
	assert.DeepEqual(t, q.Col("Rate").Float64(0), float64(float32(1.2)))
	assert.DeepEqual(t, q.Col("Rate").Float64(1), float64(float32(1.5)))

	assert.DeepEqual(t, q.Col("Age").Float(0), float32(32))
	assert.DeepEqual(t, q.Col("Age").Float64(0), float64(32))

	assert.Assert(t, cmp.Panics(func() { q.Col("rate") }))
}

func Test_ColumnFloats(t *testing.T) {
	q := PrepareTable(t)

	assert.DeepEqual(t, q.Col("Age").Floats(), []float32{32, 44})
	assert.DeepEqual(t, q.Col("Rate").Floats(), []float32{1.2, 1.5})
	assert.DeepEqual(t, q.Col("Age").Floats64(), []float64{32, 44})
	assert.DeepEqual(t, q.Col("Rate").Floats64(), []float64{float64(float32(1.2)), float64(float32(1.5))})
}

func Test_ColumnLen(t *testing.T) {
	q := PrepareTable(t)

	assert.Assert(t, q.Len() == 2)

	q2 := q.Append([]struct {
		Name string
		Age  int
	}{{"Sidorov", 55}})

	assert.Assert(t, q.Len() == 2)
	assert.Assert(t, q2.Len() == 3)
}

func Test_ColumnUnique(t *testing.T) {
	q := PrepareTable(t)
	assert.DeepEqual(t, q.Col("Name").Unique().Strings(), []string{"Ivanov", "Petrov"})

	q2 := q.Append([]struct {
		Name string
		Age  int
	}{{"Sidorov", 55}, {"Ivanov", 55}})

	assert.Assert(t, q2.Len() == 4)
	assert.DeepEqual(t, q2.Col("Name").Unique().Strings(), []string{"Ivanov", "Petrov", "Sidorov"})
	assert.DeepEqual(t, q2.Col("Age").Unique().Ints(), []int{32, 44, 55})
	assert.DeepEqual(t, q2.Col("Rate").Unique().Floats(), []float32{1.2, 1.5, 0})

	q3 := q.Append([]struct {
		Name string
		Tall int
	}{{"Sidorov", 55}, {"Ivanov", 55}})

	assert.DeepEqual(t, q3.Col("Tall").Unique().Strings(), []string{"0", "55"})
}

func Test_Col0(t *testing.T) {
	r := map[int]interface{}{}
	assert.Assert(t, cmp.Panics(func() {
		tables.Col(r)
	}))
}

func Test_Col1(t *testing.T) {
	r := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	c := tables.Col(r)
	assert.Assert(t, c.Len() == len(r))
	assert.Assert(t, c.Type() == reflect.TypeOf(r[0]))
	for i, v := range r {
		assert.Assert(t, c.Int(i) == v)
		assert.Assert(t, c.Interface(i).(int) == v)
		assert.Assert(t, c.Inspect().([]int)[i] == v)
	}
}

func Test_Col2(t *testing.T) {
	r := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(r), func(i, j int) { r[i], r[j] = r[j], r[i] })
	c := tables.Col(r)
	assert.Assert(t, c.Max().Int() == 9)
	assert.Assert(t, c.Min().Int() == 0)
	assert.Assert(t, r[c.MaxIndex()] == 9)
	assert.Assert(t, c.Index(c.MaxIndex()).Int() == 9)
	assert.Assert(t, c.Index(c.MinIndex()).Int() == 0)
}

type ColR3 int

func (a ColR3) Less(b ColR3) bool {
	return b < a
}

func Test_Col3(t *testing.T) {
	r := []ColR3{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(r), func(i, j int) { r[i], r[j] = r[j], r[i] })
	c := tables.Col(r)
	assert.Assert(t, c.Max().Int() == 0)
	assert.Assert(t, c.Min().Int() == 9)
	assert.Assert(t, c.Index(c.MaxIndex()).Int() == 0)
	assert.Assert(t, c.Index(c.MinIndex()).Int() == 9)
}

type ColR4 struct {
	a int
	b uint
	c float64
	e [2]byte
	d string
}

func MkColR4(i int) *ColR4 {
	return &ColR4{
		0,
		uint(1),
		float64(2) * 0.1,
		[2]byte{0, byte(i)},
		fmt.Sprintf("col4:%d", i),
	}
}

func Test_Col4(t *testing.T) {
	r := []*ColR4{MkColR4(0), MkColR4(1), MkColR4(1), MkColR4(2), MkColR4(3), MkColR4(4), MkColR4(5)}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(r), func(i, j int) { r[i], r[j] = r[j], r[i] })
	c := tables.Col(r)
	assert.Assert(t, c.Max().Interface().(*ColR4).d == "col4:5")
	assert.Assert(t, c.Min().Interface().(*ColR4).d == "col4:0")
}

func Test_Col5(t *testing.T) {
	r := []*ColR4{MkColR4(0), MkColR4(1), MkColR4(1), nil, MkColR4(4), MkColR4(5)}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(r), func(i, j int) { r[i], r[j] = r[j], r[i] })
	c := tables.Col(r)
	assert.Assert(t, c.Max().Interface().(*ColR4).d == "col4:5")
	assert.Assert(t, c.Min().Interface().(*ColR4) == nil)
}

func Test_Col6(t *testing.T) {
	r := []*ColR4{MkColR4(0), MkColR4(0), MkColR4(0), MkColR4(1)}
	c := tables.Col(r)
	assert.Assert(t, c.Max().Interface().(*ColR4).d == "col4:1")
	assert.Assert(t, c.Min().Interface().(*ColR4).d == "col4:0")
}

func Test_Less1(t *testing.T) {
	a := map[int]interface{}{0: 0}
	assert.Assert(t, cmp.Panics(func() {
		util.Less(reflect.ValueOf(a), reflect.ValueOf(a))
	}))
	assert.Assert(t, cmp.Panics(func() {
		util.Less(reflect.ValueOf(1), reflect.ValueOf(""))
	}))
	assert.Assert(t, util.Less(reflect.ValueOf([2]int{0, 1}), reflect.ValueOf([2]int{0, 2})))
	assert.Assert(t, cmp.Panics(func() {
		util.Less(reflect.ValueOf([2]int{0, 1}), reflect.ValueOf([1]int{0}))
	}))
}

func Test_Lazy1(t *testing.T) {
	q := tables.FillUp(lazy.New(trList))
	assertTrData(t, q)

	q = tables.ConqFillUp(lazy.New(trList), 6)
	assertTrData(t, q)

	z := q.Lazy(TR{})
	z.Close()
	q = tables.FillUp(z)
	assert.Assert(t, q.Len() == 0)
}

func Test_Lazy2(t *testing.T) {
	q := tables.New(trList)
	r := lazy.New(trList).Filter(func(r TR) bool { return r.Age > 30 }).Collect().([]TR)
	q2 := tables.ConqFillUp(q.Lazy(func(r TR) bool { return r.Age > 30 }), 6)
	for i, v := range r {
		assert.DeepEqual(t, util.MapInterface(q2.Row(i)),
			map[string]interface{}{
				"Name": v.Name,
				"Age":  v.Age,
				"Rate": v.Rate,
			})
		assert.Assert(t, v.Age > 30)
	}
}

func Test_Lazy3(t *testing.T) {
	q := tables.New(trList)
	q2 := tables.ConqFillUp(q.Lazy(func(r TR) TR { return r }), 6)
	assertTrData(t, q2)
}

func Test_Lazy4(t *testing.T) {
	q := tables.New(trList)
	q2 := tables.ConqFillUp(q.Lazy(TR{}), 6)
	assertTrData(t, q2)
	q2 = tables.ConqFillUp(q.Lazy(&TR{}), 6)
	assertTrData(t, q2)
}

func Test_Lazy5(t *testing.T) {
	q := tables.New(trList)
	assert.Assert(t, cmp.Panics(func() {
		q.Lazy(func(int) int { return 0 })
	}))
}

func Test_NA1(t *testing.T) {
	q := tables.Empty()
	q2 := q.Append([]struct{ Name string }{{"Hello"}})
	q3 := q2.Append([]struct{ Age int }{})
	assert.Assert(t, q3.Col("Name").Len() == 1)
	assert.Assert(t, q3.Col("Age").Len() == 1)
	assert.Assert(t, !q3.Col("Name").Na(0))
	assert.Assert(t, q3.Col("Age").Na(0))
}

func Test_NA2(t *testing.T) {
	q := tables.Empty()
	q2 := q.Append([]struct {
		Name string
		Rate float32
	}{{"Hello", 1.2}})
	q3 := q2.Append([]struct {
		Age  int
		Rate float32
	}{{0, 0}})

	q4 := q3.Append([]struct {
		Name string
		Age  int
		Rate float32
	}{{"Hello", 0, 0}})

	q5 := q4.FillNa(struct {
		Name string
		Age  int
	}{"Empty", -1})
	q6 := q4.FillNa(map[string]interface{}{"Name": "Empty", "Age": -1})
	q7 := q4.FillNa(map[string]interface{}{"Rate": 0})
	q8 := q4.FillNa(map[string]interface{}{"Name": 0, "Age": -1.0})

	assert.Assert(t, q4.Col("Name").Len() == 3)
	assert.Assert(t, q4.Col("Age").Len() == 3)

	assert.Assert(t, !q4.Col("Name").Na(0))
	assert.Assert(t, q4.Col("Age").Na(0))
	assert.Assert(t, q4.Col("Name").Na(1))
	assert.Assert(t, !q4.Col("Age").Na(1))
	assert.Assert(t, !q4.Col("Name").Na(2))
	assert.Assert(t, !q4.Col("Age").Na(2))

	assert.Assert(t, q4.DropNa().Len() == 1)
	assert.Assert(t, q4.DropNa("Name").Len() == 2)
	assert.Assert(t, q4.DropNa("Age").Len() == 2)
	assert.Assert(t, q2.DropNa().Len() == 1)

	assert.Assert(t, cmp.Panics(func() {
		q2.DropNa("pigs")
	}))

	assert.Assert(t, cmp.Panics(func() {
		q4.FillNa("pigs")
	}))

	assert.Assert(t, cmp.Panics(func() {
		q4.FillNa(struct{ Name1 string }{})
	}))

	assert.Assert(t, q5.DropNa().Len() == 3)
	assert.Assert(t, !q5.Col("Name").Na(0))
	assert.Assert(t, !q5.Col("Age").Na(0))
	assert.Assert(t, !q5.Col("Name").Na(1))
	assert.Assert(t, !q5.Col("Age").Na(1))
	assert.Assert(t, !q5.Col("Name").Na(2))
	assert.Assert(t, !q5.Col("Age").Na(2))
	assert.Assert(t, q5.Col("Age").Int(0) == -1)
	assert.Assert(t, q5.Col("Name").String(1) == "Empty")

	assert.Assert(t, q6.DropNa().Len() == 3)
	assert.Assert(t, !q6.Col("Name").Na(0))
	assert.Assert(t, !q6.Col("Age").Na(0))
	assert.Assert(t, !q6.Col("Name").Na(1))
	assert.Assert(t, !q6.Col("Age").Na(1))
	assert.Assert(t, !q6.Col("Name").Na(2))
	assert.Assert(t, !q6.Col("Age").Na(2))
	assert.Assert(t, q6.Col("Age").Int(0) == -1)
	assert.Assert(t, q6.Col("Name").String(1) == "Empty")

	assert.Assert(t, q7.DropNa().Len() == 1)
	assert.Assert(t, q8.DropNa().Len() == 3)
	assert.Assert(t, q8.Col("Age").Int(0) == -1)
	assert.Assert(t, q8.Col("Name").String(1) == "0")
}
