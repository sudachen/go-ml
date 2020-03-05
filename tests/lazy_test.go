package tests

import (
	"bytes"
	"github.com/sudachen/go-foo/fu"
	"github.com/sudachen/go-ml/tables"
	"github.com/sudachen/go-ml/tables/csv"
	"github.com/sudachen/go-ml/tables/rdb"
	"gotest.tools/assert"
	"reflect"
	"testing"
)

func Test_LazyCsvRdb1(t *testing.T) {
	const CSV = `id,f1,f2,f3,f4
4,1,2,3,4
8,5,6,7,8
12,9,10,11,12
`
	z := csv.Source(fu.StringIO(CSV)).
		Map(func(x struct {
			Id string    `id`
			F  []float64 `f*`
		}) (y struct {
			Id     string
			Target float64
		}) {
			y.Target = fu.Maxd(0, x.F...)
			y.Id = x.Id
			return
		})

	err := z.
		Parallel().
		Drain(rdb.Sink("sqlite3:file:/tmp/test.db",
			rdb.Table("maxd"),
			rdb.DropIfExists,
			rdb.VARCHAR("Id").PrimaryKey(),
			rdb.DECIMAL("Target", 2)))

	assert.NilError(t, err)

	c := rdb.Source("sqlite3:file:/tmp/test.db",
		rdb.Table("maxd")).
		Filter(func(x struct {
			Id     string
			Target string
		}) bool {
			return x.Id == x.Target
		}).
		LuckyCount()

	assert.Assert(t, c == 3)

	c = rdb.Source("sqlite3:file:/tmp/test.db",
		rdb.Table("maxd")).
		Filter(func(x struct {
			Id     string
			Target string
		}) bool {
			return x.Id != x.Target
		}).
		LuckyCount()

	assert.Assert(t, c == 0)

	bf := bytes.Buffer{}
	err = rdb.Source("sqlite3:file:/tmp/test.db", rdb.Query("select Id from maxd")).
		Map(struct {
			Id string
			F4 float64 `Id`
		}{}).
		Transform(func(x struct{ F4 float64 }) (y struct{ F3 float64 }) {
			y.F3 = x.F4 - 1
			return
		}).
		Transform(func(x struct{ F4 float64 }) (y struct{ F2, F1 float64 }) {
			y.F1 = x.F4 - 3
			y.F2 = x.F4 - 2
			return
		}).
		Parallel().
		Map(struct {
			Id             string
			F1, F2, F3, F4 float64
		}{}).
		Drain(csv.Sink(&bf,
			csv.Column("Id").As("id"),
			csv.Column("F*").Round(2).As("f*")))

	assert.NilError(t, err)
	assert.Assert(t, bf.String() == CSV)
}

/*
func Test_LazyBatch(t *testing.T) {
	dataset := fu.External("https://datahub.io/machine-learning/iris/r/iris.csv",
		fu.Cached("go-ml/datasets/iris/iris.csv"))

	cls := tables.Enumset{}

	z := csv.Source(dataset,
		csv.Float32("sepallength").As("Feature1"),
		csv.Float32("sepalwidth").As("Feature2"),
		csv.Float32("petallength").As("Feature3"),
		csv.Float32("petalwidth").As("Feature4"),
		csv.Meta(cls.Integer(), "class").As("Label"))

	q := z.RandSkip(42, 0.3).Parallel().LuckyCollect()
	q2 := z.Rand(42, 0.3).Parallel().LuckyCollect()
	assert.Assert(t, q.Len()+q2.Len() == z.LuckyCount())

	n := 0
	l := 0
	batch := 30
	z.RandSkip(42, 0.3).Batch(batch).Parallel().LuckyDrain(func(v reflect.Value) error {
		if v.Kind() == reflect.Bool {
			return nil
		}
		k := v.Interface().(*tables.Table)
		for g := 0; g < fu.Mini(batch, k.Len()); g++ {
			a, b := k.Row(g), q.Row(n*batch+g)
			for e, c := range a {
				assert.DeepEqual(t, b[e].Interface(), c.Interface())
			}
			for e, c := range b {
				assert.DeepEqual(t, a[e].Interface(), c.Interface())
			}
		}
		n++
		l += k.Len()
		return nil
	})

	assert.Assert(t, l == q.Len())
}
*/