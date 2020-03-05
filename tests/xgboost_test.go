package tests

import (
	"bufio"
	"fmt"
	"github.com/sudachen/go-foo/fu"
	"github.com/sudachen/go-ml/base"
	"github.com/sudachen/go-ml/tables"
	"github.com/sudachen/go-ml/tables/csv"
	"github.com/sudachen/go-ml/xgb"
	"gotest.tools/assert"
	"testing"
)

func Test_XgboostVersion(t *testing.T) {
	v := xgb.LibVersion()
	fmt.Println(v)
}


func Test_Linear(t *testing.T) {
	dataset := fu.External("https://datahub.io/machine-learning/iris/r/iris.csv",
		fu.Cached("go-ml/datasets/iris/iris8.csv"))

	fmt.Println(fu.Cached("go-ml/datasets/iris/iris8.csv"))

	rd,err := dataset.Open()
	assert.NilError(t, err)
	s,err := bufio.NewReader(rd).ReadString(byte('\n'))
	assert.NilError(t, err)
	fmt.Println(s)
	rd.Close()

	cls := tables.Enumset{}
	z := csv.Source(dataset,
		csv.Float32("sepallength").As("Feature1"),
		csv.Float32("sepalwidth").As("Feature2"),
		csv.Float32("petallength").As("Feature3"),
		csv.Float32("petalwidth").As("Feature4"),
		csv.Meta(cls.Integer(), "class").As("Label"))

	fmt.Println(z.RandomFlag("Test",42,0.3).First(15).LuckyCollect())

	estimator := xgb.GBTree(
		xgb.Softmax,
		xgb.Rounds(1000),
		xgb.LearnRate(0.1),
		xgb.MaxDepth(10),
		xgb.Nestimators(1000)).
		Feed(base.Dataset{
			Source: z.RandomFlag("Test", 42, 0.3),
			Label:  "Label",
			Test:   "Test",
			//Features: []string{"Feature*"},
		}).
		LuckyFit()

	fmt.Println("predict")
	//w1 := z.Rand(42,0.3).Map(estimator).Round(2).LuckyCollect() /fmt.Println(w1.Head(15))
	w2 := z.Rand(42, 0.2).Transform(estimator).Round(2).LuckyCollect()
	fmt.Println(w2.Head(150))
}
