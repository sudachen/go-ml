package xgb

import (
	"github.com/sudachen/go-ml/ml"
)

type Model struct {
	Algorithm    booster
	Function     objective
	Iterations   int
	LearningRate float64
	MaxDepth     int
	Estimators   int
	Seed         int
	Result       string
	Extra        Params
}

func (e Model) Feed(ds ml.Dataset) ml.FatModel {
	return func(opts ...interface{}) (ml.Predictor, error) {
		return fit(e, ds, opts...)
	}
}

type Params map[string]interface{}
