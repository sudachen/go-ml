package tests

import (
	"github.com/sudachen/go-ml/datasets/iris"
	"github.com/sudachen/go-ml/tables"
	"gotest.tools/assert"
	"testing"
)

func Test_Batch1(t *testing.T) {
	q := iris.Data.
		Batch(10).
		Transform(func(t *tables.Table)(*tables.Table,error){
			q := t.Except("").With(tables.Col([]int{0,1,2,3,4,5,6,7,8,9}[0:t.Len()]),"Index")
			return q, nil
		}).
		Flat().
		LuckyCollect()
	x := iris.Data.LuckyCollect()
	assert.Assert(t, q.Len() == x.Len())
	for i:=0;i<q.Len();i++ {
		assert.Assert(t, q.Col("Feature1").Float(i) == x.Col("Feature1").Float(i))
	}
}

