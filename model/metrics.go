package model

import (
	"github.com/sudachen/go-ml/fu"
	"github.com/sudachen/go-ml/tables"
	"reflect"
)

const Subset = "Subset"
const Iteration = "Iteration"
const Test = "test"
const Train = "train"

/*
MetricsUpdater interface
*/
type MetricsUpdater interface {
	Update(result, label reflect.Value)
	Complete() (fu.Struct, bool)
}

/*
Metrics interface
*/
type Metrics interface {
	New(iteration int, subset string) MetricsUpdater
	Names() []string
	HistoryLength() int
}

type Score func(train, test fu.Struct) float64

const HistoryLength = 3

func EvaluateMetrics(iteration int, subset string, result, label *tables.Column, metricsf Metrics) (fu.Struct, bool) {
	mu := metricsf.New(iteration, subset)
	BatchUpdateMetrics(result, label, mu)
	return mu.Complete()
}

func BatchUpdateMetrics(result, label *tables.Column, mu MetricsUpdater) {
	rc, _ := result.Raw()
	lc, _ := label.Raw()
	for i := 0; i < rc.Len(); i++ {
		mu.Update(rc.Index(i), lc.Index(i))
	}
}
