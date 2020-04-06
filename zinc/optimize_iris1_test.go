package zinc

import (
	"fmt"
	"github.com/sudachen/go-iokit/iokit"
	"github.com/sudachen/go-ml/dataset/iris"
	"github.com/sudachen/go-ml/fu"
	"github.com/sudachen/go-ml/metrics/classification"
	"github.com/sudachen/go-ml/model"
	"github.com/sudachen/go-ml/model/hyperopt"
	"github.com/sudachen/go-ml/xgb"
	"testing"
)

func Test_Optimize_Iris1(t *testing.T) {

	par := hyperopt.Space{
		Source:     iris.Data.RandSkip(43, 0.2),
		Features:   []string{"Feature*"},
		Label:      "Label",
		Kfold:      3,
		Iterations: 25,
		Metrics:    &classification.Metrics{},
		Score:      hyperopt.TrailScore(classification.Accuracy),
		Direction:  hyperopt.MaximizeScore,
		ModelFunc: xgb.Model{ Algorithm: xgb.LinearBoost, Function: xgb.Softmax }.ModelFunc,
		Variance: hyperopt.Variance{
			"MaxDepth":     hyperopt.IntRange{1, 5},
			"Estimators":   hyperopt.IntRange{1, 20},
			"LearningRate": hyperopt.Range{0.1, 1}},
	}.LuckyOptimize(80)

	fmt.Println(par)

	modelFile := iokit.File(fu.ModelPath("xgboost_test_v1.xgb"))
	metrics := xgb.Model{
		Algorithm: xgb.LinearBoost,
		Function:  xgb.Softmax,
	}.Apply(par.Params).Feed(model.Dataset{
		Source:   iris.Data.RandomFlag("Test", 43, 0.2),
		Label:    "Label",
		Test:     "Test",
		Features: []string{"Feature*"},
	}).LuckyFit(30, modelFile, &classification.Metrics{})

	fmt.Println(metrics.Round(4))
}
