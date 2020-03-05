package xgb

import (
	"github.com/sudachen/go-ml/mlutil"
	"github.com/sudachen/go-ml/xgb/capi"
	"reflect"
)

func (x XGBoost) Predict(lr mlutil.Struct) mlutil.Struct {
	m := matrix{make([]float32, len(x.features)), nil}.
		set(0, lr, -1, x.features).create(false, 1)
	defer m.Free()
	y := capi.Predict(x.handle, m.handle, 0)
	cols := make([]reflect.Value, len(x.predicts))
	for i, c := range y {
		cols[i] = reflect.ValueOf(c)
	}
	return mlutil.Struct{x.predicts, cols, mlutil.Bits{}}
}
