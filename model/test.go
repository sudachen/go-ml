package model

import (
	"github.com/sudachen/go-foo/fu"
	"github.com/sudachen/go-ml/mlutil"
	"github.com/sudachen/go-ml/tables"
	"reflect"
)

func Evaluate(source tables.AnyData, label string, m PredictionModel, batchsize int, mx ...Metrics) (t *tables.Table, err error) {
	mr := Measurer(mx).Begin()
	err = source.Lazy().Batch(batchsize).Transform(m.FeaturesMapper).Drain(
		func(v reflect.Value) (e error) {
			if v.Kind() == reflect.Bool {
				if v.Bool() {
					metrics, _ := mr.Complete(0, "test")
					t = tables.New([]*mlutil.Struct{&metrics})
				}
			} else {
				tr := v.Interface().(*tables.Table)
				mr.ColumnUpdate(tr.Col(m.Predicted()), tr.Col(label))
			}
			return
		})
	return
}

func LuckyEvaluate(source tables.AnyData, label string, m PredictionModel, batchsize int, mx ...Metrics) *tables.Table {
	t, err := Evaluate(source, label, m, batchsize, mx...)
	if err != nil {
		panic(fu.Panic(err))
	}
	return t
}
