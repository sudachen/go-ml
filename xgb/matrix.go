package xgb

import (
	"github.com/sudachen/go-ml/tables"
	"github.com/sudachen/go-ml/xgb/capi"
	"unsafe"
)

type xgbmatrix struct {
	handle unsafe.Pointer
}

func (m xgbmatrix) Free() {
	if m.handle != nil {
		capi.Free(m.handle)
	}
}

func matrix32(matrix tables.Matrix) xgbmatrix {
	if matrix.Length > 0 {
		handle := capi.Matrix(matrix.Features, matrix.Length, matrix.Width)
		if matrix.LabelsWidth > 0 {
			capi.SetInfo(handle, "label", matrix.Labels)
		}
		return xgbmatrix{handle}
	} else {
		return xgbmatrix{nil}
	}
}
