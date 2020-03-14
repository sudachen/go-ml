package xgb

import (
	"github.com/sudachen/go-ml/tables"
	"github.com/sudachen/go-ml/xgb/capi"
)

func (x xgbinstance) MapFeatures(t *tables.Table) (*tables.Table, error) {
	matrix, err := t.Matrix(x.features)
	if err != nil { panic(err) }
	m := matrix32(matrix)
	defer m.Free()
	y := capi.Predict(x.handle, m.handle, 0)
	pred := tables.Matrix{
		Features:    y,
		Labels:      nil,
		Width:       len(y)/matrix.Length,
		Length:      matrix.Length,
		LabelsWidth: 0,
	}
	return t.Except(x.features...).With(pred.AsColumn(), x.predicts), nil
}
