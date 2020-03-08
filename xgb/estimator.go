package xgb

import "github.com/sudachen/go-ml/mlutil"

type Estimator struct {
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

func (e Estimator) Feed(ds mlutil.Dataset) mlutil.FatModel {
	return func(opts ...interface{}) (mlutil.Predictor, error) {
		return fit(e, ds, opts...)
	}
}

type Params map[string]interface{}
