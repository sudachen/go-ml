package xgb

import (
	"github.com/sudachen/go-foo/fu"
	"github.com/sudachen/go-ml/tables"
	"github.com/sudachen/go-ml/xgb/capi"
	"io"
	"io/ioutil"
	"runtime"
)

type PredictionModel struct {
	features []string
	predicts string
	source   fu.Input
}

func (x *xgbinstance) MapFeatures(t *tables.Table) (*tables.Table, error) {
	matrix, err := t.Matrix(x.features)
	if err != nil {
		panic(err)
	}
	m := matrix32(matrix)
	defer m.Free()
	y := capi.Predict(x.handle, m.handle, 0)
	pred := tables.Matrix{
		Features:    y,
		Labels:      nil,
		Width:       len(y) / matrix.Length,
		Length:      matrix.Length,
		LabelsWidth: 0,
	}
	return t.Except(x.features...).With(pred.AsColumn(), x.predicts), nil
}

func (model PredictionModel) Features() []string {
	return model.features
}

func (model PredictionModel) Predicted() string {
	return model.predicts
}

func (model PredictionModel) FeaturesMapper(int) (fm tables.FeaturesMapper, err error) {
	var rd io.ReadCloser
	if rd, err = model.source.Open(); err != nil {
		return
	}
	defer rd.Close()
	var bs []byte
	if bs, err = ioutil.ReadAll(rd); err != nil {
		return
	}
	x := &xgbinstance{
		handle:   capi.Create(),
		features: model.features,
		predicts: model.predicts,
	}
	runtime.SetFinalizer(x, func(x *xgbinstance) { x.Close() })
	capi.SetModel(x.handle, bs)
	return x, nil
}
