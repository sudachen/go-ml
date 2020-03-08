package xgb

import (
	"github.com/sudachen/go-ml/mlutil"
	"github.com/sudachen/go-ml/xgb/capi"
	"unsafe"
)

func LibVersion() mlutil.VersionType {
	return mlutil.VersionType(capi.LibVersion)
}

type XGBoost struct {
	handle   unsafe.Pointer
	features map[string]int
	predicts []string
}

func (x XGBoost) Close() (err error) {
	return
}

/*func (x XGBoost) Acquire() (base.Predictor,error) {
	return x, nil
}*/
