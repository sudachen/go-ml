package zinc

import (
	"fmt"
	"github.com/sudachen/go-iokit/iokit"
	"github.com/sudachen/go-ml/dataset/iris"
	"github.com/sudachen/go-ml/fu"
	"github.com/sudachen/go-ml/fu/verbose"
	"github.com/sudachen/go-ml/metrics/classification"
	"github.com/sudachen/go-ml/model"
	"github.com/sudachen/go-ml/model/hyperopt"
	"github.com/sudachen/go-ml/xgb"
	"testing"
)

func Test_Optimize_Iris1(t *testing.T) {
	defer verbose.BeVerbose(verbose.Print).Revert()

	par := hyperopt.Space{
		Source:     iris.Data,
		Features:   []string{"Feature*"},
		Label:      "Label",
		Kfold:      3,
		Iterations: 19,
		Metrics:    &classification.Metrics{History: 2},
		Score:      classification.AccuracyScore,
		ModelFunc:  xgb.Model{Algorithm: xgb.LinearBoost, Function: xgb.Softmax}.ModelFunc,
		Variance: hyperopt.Variance{
			"MaxDepth":     hyperopt.IntRange{1, 5},
			"Estimators":   hyperopt.LogIntRange{1, 20},
			"LearningRate": hyperopt.Value(0.6),
		},
	}.LuckyOptimize(30)

	fmt.Println(par)

	modelFile := iokit.File(fu.ModelPath("xgboost_test_v1.xgb"))
	report := xgb.Model{
		Algorithm: xgb.TreeBoost,
		Function:  xgb.Softmax,
	}.Apply(par.Params).Feed(model.Dataset{
		Source:   iris.Data.RandomFlag("Test", 42, 0.2),
		Label:    "Label",
		Test:     "Test",
		Features: []string{"Feature*"},
	}).LuckyFit(30, modelFile, &classification.Metrics{History: 5}, classification.AccuracyScore)

	fmt.Println(report.History.Round(4))
}
