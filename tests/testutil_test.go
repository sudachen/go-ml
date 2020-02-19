package tests

import (
	"github.com/sudachen/go-ml/tables"
	"github.com/sudachen/go-ml/util"
	"gotest.tools/assert"
	"testing"
)

type TR struct {
	Name string
	Age  int
	Rate float32
}

func PrepareTable(t *testing.T) *tables.Table {
	q := tables.New([]TR{
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

	return q
}

var trList = []TR{
	{"Ivanov", 32, 1.2},
	{"Petrov", 44, 1.5},
	{"Sidorov", 55, 1.8},
	{"Gavrilov", 20, 0.9},
	{"Popova", 28, 1.0},
	{"Kozlov", 42, 1.3},
}

func TrTable() *tables.Table {
	return tables.New(trList)
}

func assertTrData(t *testing.T, q *tables.Table) {
	assert.Assert(t, q.Len() == len(trList))
	for i, r := range trList {
		assert.DeepEqual(t, util.MapInterface(q.Row(i)),
			map[string]interface{}{
				"Name": r.Name,
				"Age":  r.Age,
				"Rate": r.Rate,
			})
	}
}
