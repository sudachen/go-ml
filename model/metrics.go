package model

import (
	"github.com/sudachen/go-ml/fu"
	"github.com/sudachen/go-ml/tables"
	"math"
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
LossCol is the Loss column name
*/
const LossCol = "Loss"

/*
ErrorCol is the Error column name
*/
const ErrorCol = "Error"

/*
RmseCol is the Root Mean fo Squared Error column name
*/
const RmseCol = "Rmse"

/*
MaeCol is the Mean of Absolute Error column name
*/
const MaeCol = "Mae"

/*
AccuracyCol is the Error column name
*/
const AccuracyCol = "Accuracy"

/*
SensitivityCol is the Sensitivity column name
*/
const SensitivityCol = "Sensitivity"

/*
PrecisionCol is the Precision column name
*/
const PrecisionCol = "Precision"

/*
F1ScoreCol is the F1score column name
*/
const F1ScoreCol = "F1score"

/*
TotalCol is the Total column name
*/
const TotalCol = "Total"

/*
CorrectCol i th Correct column name
*/
const CorrectCol = "Correct"

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
	// updates metrics with prediction result and label
	// loss is an optional and can be used in LossScore on the training
	Update(result, label reflect.Value, loss float64)
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
BatchUpdateMetrics updates metrics for a batch of training results
*/
func BatchUpdateMetrics(result, label *tables.Column, mu MetricsUpdater) {
	rc, _ := result.Raw()
	lc, _ := label.Raw()
	for i := 0; i < rc.Len(); i++ {
		mu.Update(rc.Index(i), lc.Index(i), 0)
	}
}

/*
Error of an ML algorithm, can have any value
*/
func Error(lr fu.Struct) float64 { return lr.Float(ErrorCol) }

/*
ErrorScore scores error in interval [0,1], Greater is better
*/
func ErrorScore(train, test fu.Struct) float64 {
	a := 1 / (1 + math.Exp(Error(train)))
	b := 1 / (1 + math.Exp(Error(test)))
	return fu.Mind(a, b)
}

/*
Accuracy of an ML algorithm, has a value in the interval [0,1]
*/
func Accuracy(lr fu.Struct) float64 { return lr.Float(AccuracyCol) }

/*
AccuracyScore scores accuracy in interval [0,1], Greater is better
*/
func AccuracyScore(train, test fu.Struct) float64 {
	a1, a2 := Accuracy(train), Accuracy(test)
	if a1 > a2 {
		a2, a1 = a1, a2
	}
	return (a1 - (a2-a1)/2)
}

/*
Loss is the maen of the ML algorithm loss function. It can have any float value
*/
func Loss(lr fu.Struct) float64 { return lr.Float(LossCol) }

/*
LossScore scores loss in interval [0,1], Greater is better
*/
func LossScore(train, test fu.Struct) float64 {
	a := 1 / (1 + math.Exp(Loss(train)))
	b := 1 / (1 + math.Exp(Loss(test)))
	return fu.Mind(a, b)
}
