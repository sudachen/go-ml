package xgb

import (
	"encoding/json"
	"github.com/sudachen/go-iokit/iokit"
	"github.com/sudachen/go-ml/fu"
	"github.com/sudachen/go-ml/model"
	"github.com/sudachen/go-ml/model/hyperopt"
	"io"
)

/*
Model is a XGBoost model definition
*/
type Model struct {
	Algorithm booster
	Function  objective

	Estimators int    // The number of trees
	Seed       int    // random generator seed
	Predicted  string // name of predicted value column

	MinChildWeight float64 //the minimum sum of weights of all observations required in a child.
	Gamma          float64 // Specifies the minimum loss reduction required to make a split.

	// Denotes the fraction of observations to be randomly samples for each tree.
	// Typical values: 0.5-1
	Subsample float64

	Lambda float64 // L2 regularization
	Alpha  float64 // L1 regularization

	// Makes the model more robust by shrinking the weights on each step
	// Typical values: 0.01-0.2
	LearningRate float64

	// The maximum depth of a tree.
	// Used to control over-fitting as higher depth will allow model
	// to learn relations very specific to a particular sample.
	// Typical values: 3-10
	MaxDepth int

	Extra Params
}

// Params - xgboost model extra parameters
type Params map[string]interface{}

/*
Feed model with data
*/
func (e Model) Feed(ds model.Dataset) model.FatModel {
	return func(iterations int, output iokit.Output, metricsf model.Metrics, scoref model.Score) (model.Report, error) {
		iterations = fu.Fnzi(iterations, 1)
		return fit(iterations, e, ds, output, metricsf, scoref)
	}
}

/*
ModelFunc upadtes xgboost model with parameters for hyper-optimization
*/
func (m Model) ModelFunc(p hyperopt.Params) model.HungryModel {
	return m.Apply(p)
}

/*
Apply parameters to define model specific
*/
func (m Model) Apply(p hyperopt.Params) Model {
	hyperopt.Apply(&m, p)
	return m
}

/*
ObjectifyModel creates xgboost predictor from the model collection
*/
func ObjectifyModel(c map[string]iokit.Input) (pm model.PredictionModel, err error) {
	var rd io.ReadCloser
	if rd, err = c["info.json"].Open(); err != nil {
		return
	}
	defer rd.Close()
	cf := map[string]interface{}{}
	if err = json.NewDecoder(rd).Decode(&cf); err != nil {
		return
	}
	m := predictionModel{
		source:   c["model.bin.xz"],
		features: fu.Strings(cf["features"]),
		predicts: cf["predicts"].(string),
	}
	return m, nil
}

/*
Objectify creates xgboost prediction object from an input
*/
func Objectify(source iokit.InputOutput, collection ...string) (fm model.PredictionModel, err error) {
	x := fu.Fnzs(fu.Fnzs(collection...), "model")
	m, err := model.Objectify(source, model.ObjectifyMap{x: ObjectifyModel})
	if err != nil {
		return
	}
	return m[x], nil
}

/*
LuckyObjectify is the errorless version of Objectify
*/
func LuckyObjectify(source iokit.InputOutput, collection ...string) model.PredictionModel {
	fm, err := Objectify(source, collection...)
	if err != nil {
		panic(err)
	}
	return fm
}
