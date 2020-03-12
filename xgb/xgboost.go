package xgb

import (
	"github.com/sudachen/go-ml/mlutil"
	"github.com/sudachen/go-ml/xgb/capi"
	"unsafe"
)

func LibVersion() mlutil.VersionType {
	return mlutil.VersionType(capi.LibVersion)
}

type xgbinstance struct {
	handle   unsafe.Pointer
	features []string
	predicts []string
}

func (x xgbinstance) Close() (_ error) {
	capi.Close(x.handle)
	x.handle = nil
	return
}

func (x xgbinstance) BatchSize() (min, max int) {
	return 1,64
}

/*func (x XGBoost) Acquire() (base.xgbinstance,error) {
	return x, nil
}*/
