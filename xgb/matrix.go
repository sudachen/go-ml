package xgb

import (
	"github.com/sudachen/go-ml/base"
	"github.com/sudachen/go-ml/internal"
	"github.com/sudachen/go-ml/logger"
	"github.com/sudachen/go-ml/mlutil"
	"github.com/sudachen/go-ml/xgb/capi"
	"math"
	"reflect"
	"unsafe"
)

type Matrix struct {
	handle        unsafe.Pointer
	width, length int
}

func (m Matrix) Free() {
	capi.Free(m.handle)
}

func matrix32(dat, labdat []float32, length int) (mx Matrix) {
	width := len(dat) / length
	m := capi.Matrix(dat, length, width)
	mx = Matrix{m, width, length}
	if labdat != nil {
		capi.SetInfo(m, "label", labdat)
	}
	return
}

type matrix [2][]float32

func (m matrix) create(labels bool, rows int) (mx Matrix) {
	if labels {
		return matrix32(m[0], m[1], rows)
	}
	return matrix32(m[0], nil, rows)
}

func (m matrix) set(row int, lr base.Struct, label int, features map[string]int) matrix {
	width := len(features)
	fc := 0
	for i, n := range lr.Names {
		x := float32(0)
		v := lr.Columns[i]
		if lr.Na.Bit(i) {
			x = float32(math.NaN())
		} else {
			switch v.Kind() {
			case reflect.Float32:
				x = v.Interface().(float32)
			case reflect.Float64:
				x = float32(v.Interface().(float64))
			case reflect.Int:
				x = float32(v.Interface().(int))
			default:
				x = mlutil.Convert(v, false, internal.Float32Type).Interface().(float32)
			}
		}
		if i == label {
			m[1][row] = x
		} else if j, ok := features[n]; ok {
			m[0][row*width+j] = x
			fc++
		}
	}
	if fc < width {
		logger.Warningf("not enough features in dataset, required %d but exists only %d", width, fc)
	}
	return m
}
