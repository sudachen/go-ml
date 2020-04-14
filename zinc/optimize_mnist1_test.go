package zinc

import (
	"fmt"
	"github.com/sudachen/go-iokit/iokit"
	"github.com/sudachen/go-ml/dataset/mnist"
	"github.com/sudachen/go-ml/fu"
	"github.com/sudachen/go-ml/fu/verbose"
	"github.com/sudachen/go-ml/metrics/classification"
	"github.com/sudachen/go-ml/model"
	"github.com/sudachen/go-ml/model/hyperopt"
	"github.com/sudachen/go-ml/xgb"
	"testing"
)

func Test_Optimize_Mnist1(t *testing.T) {
	defer verbose.BeVerbose(verbose.Print).Revert()

	par := hyperopt.Space{
		Source:     mnist.Data.Rand(13, 0.35),
		Features:   mnist.Features,
		Kfold:      3,
		Iterations: 19,
		Metrics:    &classification.Metrics{},
		Score:      classification.AccuracyScore,
		ModelFunc:  xgb.Model{Algorithm: xgb.TreeBoost, Function: xgb.Softmax}.ModelFunc,
		Variance: hyperopt.Variance{
			"MaxDepth":     hyperopt.IntRange{1, 10},
			"Estimators":   hyperopt.LogIntRange{1, 1000},
			"LearningRate": hyperopt.Range{0.1, 0.9},
		},
	}.LuckyOptimize(30)

	fmt.Println(par)

	modelFile := iokit.File(fu.ModelPath("xgboost_mnist_v1.zip"))
	report := xgb.Model{
		Algorithm: xgb.TreeBoost,
		Function:  xgb.Softmax,
	}.Apply(par.Params).Feed(model.Dataset{
		Source:   mnist.T10k.RandomFlag(model.TestCol, 42, 0.2),
		Features: mnist.Features,
	}).LuckyTrain(model.Training{
		Iterations: 10,
		ModelFile: modelFile,
		Metrics: &classification.Metrics{},
		Score: classification.AccuracyScore,
	})

	fmt.Println(report.History.Round(4))
}
