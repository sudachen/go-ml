package ml

import (
	"github.com/sudachen/go-ml/tables"
)

/*
HungryModel is an ML algorithm grows from a data to predict something
Needs to be fattened by Feed method to fit.
*/
type HungryModel interface {
	Feed(Dataset) FatModel
}

// FatModel is fattened model (a training function of model instance bounded to a dataset)
type FatModel func(...Fitparam) (FeaturesMapper, error)

/*
Fit trains a fattened (Fat) model
*/
func (f FatModel) Fit(opts ...Fitparam) (FeaturesMapper, error) {
	return f(opts...)
}

/*
LukyFit trains fattened (Fat) model and trows any occored errors as a panic
*/
func (f FatModel) LuckyFit(opts ...Fitparam) FeaturesMapper {
	e, err := f.Fit(opts...)
	if err != nil {
		panic(err)
	}
	return e
}

/*
FeaturesMapper is able to predict by the same features it's trained
It maps features to predicted value transforming the table

If model can fit more (NN for example), it implements the HungryModel interface as wall as the FeaturesMapper.
*/
type FeaturesMapper interface {
	// returns new table with all original columns except features
	// adding one new column with prediction
	MapFeatures(*tables.Table) (*tables.Table, error)
	// release all native resources bounded to prediction backend
	// like mnxnet NN executor, symbols, arrays or XGboost objects
	Close() error
	// features model uses when maps features
	//  the same as Features in the training dataset
	Features() []string
	// column name model adds to result table when maps features
	//  by default it's <Model>Result where <Model> is a model specific Prefix
	Result() string
}

type Fitparam interface { Fitparam() }

// Result specifies prediction column name model adds to the table when maps features
type Result string
func (Result) Fitparam() {}

// Iterations specifies counf of training iterations (epochs for NN)
type Iterations int
func (Iterations) Fitparam() {}
