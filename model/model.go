package model

import (
	"github.com/sudachen/go-iokit/iokit"
	"github.com/sudachen/go-ml/fu"
	"github.com/sudachen/go-ml/tables"
	"github.com/sudachen/go-zorros/zorros"
)

/*
HungryModel is an ML algorithm grows from a data to predict something
Needs to be fattened by Feed method to fit.
*/
type HungryModel interface {
	Feed(Dataset) FatModel
}

type Report struct {
	History     *tables.Table // all iterations history
	TheBest     int           // the best iteration
	Test, Train fu.Struct     // the best iteration metrics
	Score       float64       // the best score
}

// FatModel is fattened model (a training function of model instance bounded to a dataset)
type FatModel func(iterations int, file iokit.Output, metrics Metrics, score Score) (Report, error)

/*
Fit trains a fattened (Fat) model
*/
func (f FatModel) Fit(iterations int, output iokit.Output, metrics Metrics, score Score) (Report, error) {
	iterations = fu.Maxi(1, iterations)
	return f(iterations, output, metrics, score)
}

/*
LuckyFit trains fattened (Fat) model and trows any occurred errors as a panic
*/
func (f FatModel) LuckyFit(iterations int, output iokit.Output, metrics Metrics, score Score) Report {
	m, err := f.Fit(iterations, output, metrics, score)
	if err != nil {
		panic(zorros.Panic(err))
	}
	return m
}

/*
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
 */
type GpuPredictionModel interface {
	PredictionModel
	// Gpu changes context of prediction backend to gpu enabled
	// it's a recommendation only, if GPU is not available or it's impossible to use it
	// the cpu will be used instead
	Gpu(...int) PredictionModel
}
