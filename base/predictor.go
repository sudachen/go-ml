package base

import (
	"github.com/sudachen/go-foo/lazy"
)

type Dataset struct {
	Source   func() lazy.Stream
	Label    string   // name of float32 field containing label to use train dfata
	Test     string   // name of int field containing 1 or 0 to select test data
	Features []string // patterns of feature names to train or test model
}

/*
ModelFarm is an ML alghoritm grows from a data to predict something
*/
type ModelFarm interface {
	Feed(Dataset) FatModel
}

type FatModel func(...interface{}) (Predictor, error)

func (f FatModel) Fit(...interface{}) (Predictor, error) {
	return f()
}

func (f FatModel) LuckyFit(...interface{}) Predictor {
	e, err := f.Fit()
	if err != nil {
		panic(err)
	}
	return e
}

/*
Predictor is a trained model able to predict by the same features it's trained
*/
type Predictor interface {
	Predict(x Struct) Struct
	Close() error
}

/*
ParallelPredictor predictor able to work in concurrent environment
*/
type ParallelPredictor interface {
	Acquire() (Predictor, error)
}
