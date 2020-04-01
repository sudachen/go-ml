package xgb

import (
	"encoding/json"
	"github.com/sudachen/go-iokit/iokit"
	"github.com/sudachen/go-ml/fu"
	"github.com/sudachen/go-ml/model"
	"github.com/sudachen/go-ml/tables"
	"io"
)

type Model struct {
	Algorithm    booster
	Function     objective
	LearningRate float32
	MaxDepth     int
	Estimators   int
	Seed         int
	Predicted    string
	Extra        Params
}

func (e Model) Feed(ds model.Dataset) model.FatModel {
	return func(iterations int, output iokit.Output, mx ...model.Metrics) (*tables.Table, error) {
		iterations = fu.Fnzi(iterations, 1)
		return fit(iterations, e, ds, output, mx...)
	}
}

type Params map[string]interface{}

func ObjectifyModel(c map[string]iokit.Input) (pm model.PredictionModel, err error) {
	var rd io.ReadCloser
	if rd, err = c["info.json"].Open(); err != nil {
		return
	}
	defer rd.Close()
	cf := map[string]interface{}{}
	if err = json.NewDecoder(rd).Decode(&cf); err != nil {
		return
	}
	m := PredictionModel{
		source:   c["model.bin.xz"],
		features: fu.Strings(cf["features"]),
		predicts: cf["predicts"].(string),
	}
	return m, nil
}

func Objectify(source iokit.InputOutput, collection ...string) (fm model.PredictionModel, err error) {
	x := fu.Fnzs(fu.Fnzs(collection...), "model")
	m, err := model.Objectify(source, model.ObjectifyMap{x: ObjectifyModel})
	if err != nil {
		return
	}
	return m[x], nil
}

func LuckyObjectify(source iokit.InputOutput, collection ...string) model.PredictionModel {
	fm, err := Objectify(source, collection...)
	if err != nil {
		panic(err)
	}
	return fm
}
