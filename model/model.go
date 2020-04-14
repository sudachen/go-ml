package model

import (
	"github.com/sudachen/go-ml/fu"
	"github.com/sudachen/go-ml/tables"
	"github.com/sudachen/go-zorros/zorros"
	"io"
)

/*
HungryModel is an ML algorithm grows from a data to predict something
Needs to be fattened by Feed method to fit.
*/
type HungryModel interface {
	Feed(Dataset) FatModel
}

/*
Report is an ML training report
*/
type Report struct {
	History     *tables.Table // all iterations history
	TheBest     int           // the best iteration
	Test, Train fu.Struct     // the best iteration metrics
	Score       float64       // the best score
}

/*
Workout is a training iteration abstraction
*/
type Workout interface {
	Iteration() int
	TrainMetrics() MetricsUpdater
	TestMetrics() MetricsUpdater
	Complete(m MemorizeMap, train, test fu.Struct, metricsDone bool) (*Report, bool, error)
	Next() Workout
}

/*
UnifiedTraining is an interface allowing to write any logging/staging backend for ML training
*/
type UnifiedTraining interface {
	// Workout returns the first iteration workout
	Workout() Workout
}

/*
FatModel is fattened model (a training function of model instance bounded to a dataset)
*/
type FatModel func(workout Workout) (*Report, error)

/*
Train a fattened (Fat) model
*/
func (f FatModel) Train(training UnifiedTraining) (*Report, error) {
	w := training.Workout()
	if c,ok := w.(io.Closer); ok {
		defer c.Close()
	}
	return f(w)
}

/*
LuckyTrain trains fattened (Fat) model and trows any occurred errors as a panic
*/
func (f FatModel) LuckyTrain(training UnifiedTraining) *Report {
	m, err := f.Train(training)
	if err != nil {
		panic(zorros.Panic(err))
	}
	return m
}

/*
PredictionModel is a predictor interface
*/
type PredictionModel interface {
	// Features model uses when maps features
	// the same as Features in the training dataset
	Features() []string
	// Column name model adds to result table when maps features.
	// By default it's 'Predicted'
	Predicted() string
	// Returns new table with all original columns except features
	// adding one new column with prediction
	FeaturesMapper(batchSize int) (tables.FeaturesMapper, error)
}

/*
GpuPredictionModel is a prediction interface able to use GPU
*/
type GpuPredictionModel interface {
	PredictionModel
	// Gpu changes context of prediction backend to gpu enabled
	// it's a recommendation only, if GPU is not available or it's impossible to use it
	// the cpu will be used instead
	Gpu(...int) PredictionModel
}
