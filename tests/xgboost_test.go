package tests

import (
	"fmt"
	"github.com/sudachen/go-foo/fu"
	"github.com/sudachen/go-ml/dataset/iris"
	"github.com/sudachen/go-ml/dataset/mnist"
	"github.com/sudachen/go-ml/metrics/classification"
	"github.com/sudachen/go-ml/mlutil"
	"github.com/sudachen/go-ml/model"
	"github.com/sudachen/go-ml/notes"
	"github.com/sudachen/go-ml/xgb"
	"testing"
)

func Test_XgboostVersion(t *testing.T) {
	v := xgb.LibVersion()
	fmt.Println(v)
}

func Test_Linear(t *testing.T) {
	np := &notes.Page{
		Title:  `Iris XGBoost Example`,
		Footer: `!(http://github.com/sudachen/go-ml)`,
	}

	ds := iris.Data.RandomFlag("Test", 43, 0.3)

	np.Display("Whole dataset", ds)
	np.Info("Whole dataset info", ds)
	//np.Display("Test dataset stats", ds.IfFlag("Test").Stats())
	//np.Display("Training dataset stats", ds.IfNotFlag("Test").Stats())

	modelFile := fu.File(mlutil.ModelPath("xgboost_test_v1.xgb"))

	metrics :=
		xgb.Model{
			Algorithm:    xgb.TreeBoost,
			Function:     xgb.Softmax,
			LearningRate: 0.3,
			MaxDepth:     10,
			Estimators:   0,
		}.
			Feed(model.Dataset{
				Source:   ds,
				Label:    "Label",
				Test:     "Test",
				Features: []string{"Feature*"},
			}).
			LuckyFit(3, modelFile, &classification.Metrics{})

	np.Display("Metrics", metrics.Round(3))
	np.Plot("Accuracy evolution by iteration", metrics, &notes.Lines{X: "Iteration", Y: []string{"Accuracy"}, Z: "Test"})

	fmt.Println(metrics)

	pred := xgb.LuckyObjectify(modelFile)

	w2 := iris.Data.
		//Rand(33, 0.05).
		Batch(64).
		Transform(pred.FeaturesMapper).
		Flat().
		Round(2).
		//First(5).
		Parallel().
		LuckyCollect()

	np.Display("Prediction", w2)
}

func Test_Tree_Mnist(t *testing.T) {
	np := &notes.Page{
		Title:  `XGBoost Mnist`,
		Footer: `!(http://github.com/sudachen/go-ml)`,
	}

	ds := mnist.Data.RandomFlag("Test", 43, 0.2)
	np.Head("Dataset first lines", ds, 5)
	np.Info("Dataset info", ds)

	modelFile := fu.File(mlutil.ModelPath("xgboost_test_v1.xgb"))
	metrics :=
		xgb.Model{
			Algorithm:    xgb.TreeBoost,
			Function:     xgb.Softmax,
			LearningRate: 0.3,
			MaxDepth:     10,
			Estimators:   100,
		}.
			Feed(model.Dataset{
				Source:   ds,
				Label:    "Label",
				Test:     "Test",
				Features: []string{"Image"},
			}).
			LuckyFit(30, modelFile, &classification.Metrics{Accuracy: 0.96})

	np.Display("Metrics", metrics.Round(3))
	np.Plot("Accuracy evolution by iteration", metrics, &notes.Lines{X: "Iteration", Y: []string{"Accuracy"}, Z: "Test"})

	fmt.Println(metrics)

	pred := xgb.LuckyObjectify(modelFile)
	metrics1 := model.LuckyEvaluate(mnist.T10k, "Label", pred, 32, &classification.Metrics{})
	fmt.Println(metrics1)
}
