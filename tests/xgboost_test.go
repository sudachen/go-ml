package tests

import (
	"fmt"
	"github.com/sudachen/go-ml/datasets/iris"
	"github.com/sudachen/go-ml/mlutil"
	"github.com/sudachen/go-ml/xgb"
	"testing"
)

func Test_XgboostVersion(t *testing.T) {
	v := xgb.LibVersion()
	fmt.Println(v)
}

func Test_Linear(t *testing.T) {
	fmt.Println(iris.Data.RandomFlag("Test", 42, 0.3).Rand(13, 0.1).LuckyCollect())

	estimator := xgb.Estimator{
		Algorithm:    xgb.TreeBoost,
		Function:     xgb.Softmax,
		Iterations:   20,
		LearningRate: 0.1,
		MaxDepth:     1,
		Estimators:   5,
		Extra:        xgb.Params{"aga": 1},
	}.
		Feed(mlutil.Dataset{
			Source:   iris.Data.RandomFlag("Test", 42, 0.3),
			Label:    "Label",
			Test:     "Test",
			Features: []string{"Feature*"},
		}).
		LuckyFit()

	w2 := iris.Data.Rand(42, 0.1).Transform(estimator).Round(2).LuckyCollect()
	fmt.Println(w2)
}
