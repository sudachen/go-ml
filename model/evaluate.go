package model

import (
	"github.com/sudachen/go-ml/fu"
	"github.com/sudachen/go-ml/tables"
	"github.com/sudachen/go-zorros/zorros"
	"reflect"
)

/*
Evaluate metrics of the given source with the prediction model
*/
func Evaluate(source tables.AnyData, label string, m PredictionModel, batchsize int, metricsf Metrics) (lr fu.Struct, err error) {
	mu := metricsf.New(0, TestSubset)
	err = source.Lazy().Batch(batchsize).Transform(m.FeaturesMapper).Drain(
		func(v reflect.Value) (e error) {
			if v.Kind() == reflect.Bool {
				if v.Bool() {
					lr, _ = mu.Complete()
				}
			} else {
				tr := v.Interface().(*tables.Table)
				BatchUpdateMetrics(tr.Col(m.Predicted()), tr.Col(label), mu)
			}
			return
		})
	return
}

/*
LuckyEvaluate is the same as Evaluate function with handling error as a panic
*/
func LuckyEvaluate(source tables.AnyData, label string, m PredictionModel, batchsize int, metricsf Metrics) fu.Struct {
	lr, err := Evaluate(source, label, m, batchsize, metricsf)
	if err != nil {
		panic(zorros.Panic(err))
	}
	return lr
}
