package tests

import (
	"fmt"
	"github.com/sudachen/go-ml/datasets/iris"
	"github.com/sudachen/go-ml/ml"
	"github.com/sudachen/go-ml/xgb"
	"testing"
)

func Test_XgboostVersion(t *testing.T) {
	v := xgb.LibVersion()
	fmt.Println(v)
}

func Test_Linear(t *testing.T) {
	fmt.Println(iris.Data.RandomFlag("Test", 42, 0.3).Rand(13, 0.1).LuckyCollect())

	pred := xgb.Model{
		Algorithm:    xgb.TreeBoost,
		Function:     xgb.Softmax,
		Iterations:   10,
		LearningRate: 0.1,
		MaxDepth:     1,
		Estimators:   5,
		Extra:        xgb.Params{"aga": 1},
	}.
	Feed(ml.Dataset{
		Source:   iris.Data.RandomFlag("Test", 42, 0.3),
		Label:    "Label",
		Test:     "Test",
		Features: []string{"Feature*"},
	}).
	LuckyFit(ml.Result("Pred"))

	defer pred.Close()

	w2 := iris.Data.Rand(33, 0.3).Batch(64).Transform(pred.MapFeatures).Flat().Round(2).LuckyCollect()
	fmt.Println(w2)
}
