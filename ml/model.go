package ml

import (
	"github.com/sudachen/go-foo/lazy"
	"github.com/sudachen/go-ml/tables"
)

type Dataset struct {
	Source   func() lazy.Stream
	Label    string   // name of float32 field containing label to use train dfata
	Test     string   // name of int field containing 1 or 0 to select test data
	Features []string // patterns of feature names to train or test model
}

/*
HungryModel is an ML alghoritm grows from a data to predict something
*/
type HungryModel interface {
	Feed(Dataset) FatModel
}

type FatModel func(...interface{}) (Predictor, error)

func (f FatModel) Fit(opts ...interface{}) (Predictor, error) {
	return f(opts...)
}

func (f FatModel) LuckyFit(opts ...interface{}) Predictor {
	e, err := f.Fit(opts...)
	if err != nil {
		panic(err)
	}
	return e
}

/*
Predictor is a trained model able to predict by the same features it's trained
*/
type Predictor interface {
	tables.Predictor
	Close() error
}
