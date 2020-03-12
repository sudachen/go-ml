package xgb

import (
	"github.com/sudachen/go-ml/tables"
	"github.com/sudachen/go-ml/xgb/capi"
)

func (x xgbinstance) Predict(t *tables.Table) *tables.Table {
	matrix, err := t.For(x.features...).Matrix()
	if err != nil { panic(err) }
	m := matrix32(matrix)
	defer m.Free()
	y := capi.Predict(x.handle, m.handle, 0)
	matrix2 := tables.Matrix{
		Features:    y,
		Labels:      nil,
		Width:       len(x.predicts),
		Length:      matrix.Length,
		LabelsWidth: 0,
	}
	return tables.FromMatrix(matrix2, x.predicts...)
}
