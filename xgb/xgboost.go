package xgb

import (
	"github.com/sudachen/go-ml/mlutil"
	"github.com/sudachen/go-ml/xgb/capi"
	"runtime"
	"unsafe"
)

func LibVersion() mlutil.VersionType {
	return capi.LibVersion
}

type xgbinstance struct {
	handle   unsafe.Pointer
	features []string // names of features
	predicts string   // name of predicted value
}

func (x *xgbinstance) Close() (_ error) {
	runtime.SetFinalizer(x, nil)
	capi.Close(x.handle)
	x.handle = nil
	return
}
