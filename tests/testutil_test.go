package tests

import (
	"fmt"
	"github.com/sudachen/go-foo/fu"
	"github.com/sudachen/go-ml/tables"
	"gotest.tools/assert"
	"gotest.tools/assert/cmp"
	"strings"
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
	assert.DeepEqual(t, fu.MapInterface(q.Row(0)),
		map[string]interface{}{
			"Name": "Ivanov",
			"Age":  32,
			"Rate": float32(1.2),
		})
	assert.DeepEqual(t, fu.MapInterface(q.Row(1)),
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
		assert.DeepEqual(t, fu.MapInterface(q.Row(i)),
			map[string]interface{}{
				"Name": r.Name,
				"Age":  r.Age,
				"Rate": r.Rate,
			})
	}
}

func PanicWith(text string, f func()) cmp.Comparison {
	return func() (result cmp.Result) {
		defer func() {
			if err := recover(); err != nil {
				s := fmt.Sprint(err)
				if strings.Index(s, text) >= 0 {
					result = cmp.ResultSuccess
					return
				}
				result = cmp.ResultFailure("panic `" + s + "` does not contain `" + text + "`")
			}
		}()
		f()
		return cmp.ResultFailure("did not panic")
	}

}
