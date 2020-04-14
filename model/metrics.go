package model

import (
	"github.com/sudachen/go-ml/fu"
	"github.com/sudachen/go-ml/tables"
	"reflect"
)

/*
TestCol is the default name of Test column containing a boolean flag
*/
const TestCol = "Test"

/*
PredictedCol is the default name of column containing a result of prediction
*/
const PredictedCol = "Predicted"

/*
LabelCol is the default name of column containing a training label
*/
const LabelCol = "Label"

/*
SubsetCol is the Subset column naime
*/
const SubsetCol = "Subset"

/*
IterationCol is the Iteration column name
*/
const IterationCol = "Iteration"

/*
TestSubset is the Subset column item value for test rows
*/
const TestSubset = "test"

/*
TrainSubset is the Subset column item value for train rows
*/
const TrainSubset = "train"

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
}

/*
Score is the type of function calculating a metrics score
*/
type Score func(train, test fu.Struct) float64

/*
func EvaluateMetrics(iteration int, subset string, result, label *tables.Column, metricsf Metrics) (fu.Struct, bool) {
	mu := metricsf.New(iteration, subset)
	BatchUpdateMetrics(result, label, mu)
	return mu.Complete()
}
*/

/*
BatchUpdateMetrics updates metrics for a batch of training results
*/
func BatchUpdateMetrics(result, label *tables.Column, mu MetricsUpdater) {
	rc, _ := result.Raw()
	lc, _ := label.Raw()
	for i := 0; i < rc.Len(); i++ {
		mu.Update(rc.Index(i), lc.Index(i))
	}
}
