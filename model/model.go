package model

import (
	"github.com/sudachen/go-foo/fu"
	"github.com/sudachen/go-ml/mlutil"
	"github.com/sudachen/go-ml/tables"
	"reflect"
)

/*
HungryModel is an ML algorithm grows from a data to predict something
Needs to be fattened by Feed method to fit.
*/
type HungryModel interface {
	Feed(Dataset) FatModel
}

/*
Metrics interface
*/
type Metrics interface {
	Begin()
	Update(result, label reflect.Value)
	Complete() (mlutil.Struct, bool)
	Copy() Metrics
}

// FatModel is fattened model (a training function of model instance bounded to a dataset)
type FatModel func(int, fu.Output, ...Metrics) (*tables.Table, error)

/*
Fit trains a fattened (Fat) model
*/
func (f FatModel) Fit(iterations int, output fu.Output, mx ...Metrics) (*tables.Table, error) {
	iterations = fu.Maxi(1, iterations)
	return f(iterations, output, mx...)
}

/*
LuckyFit trains fattened (Fat) model and trows any occurred errors as a panic
*/
func (f FatModel) LuckyFit(iterations int, output fu.Output, mx ...Metrics) *tables.Table {
	m, err := f.Fit(iterations, output, mx...)
	if err != nil {
		panic(fu.Panic(err))
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