package capi

/*
#include <stdlib.h>
#include <memory.h>
#include "capi.h"
*/
import "C"
import (
	"fmt"
	"github.com/sudachen/go-dl/dl"
	"github.com/sudachen/go-ml/mlutil"
	"golang.org/x/xerrors"
	"runtime"
	"unsafe"
)

var LibVersion mlutil.VersionType

func init() {
	var so dl.SO
	dlVerbose := dl.Verbose(func(text string, verbosity int) {
		if verbosity < 2 {
			// verbosity 2 relates to detailed information
			fmt.Println(text)
		}
	})
	if runtime.GOOS == "linux" && runtime.GOARCH == "amd64" {
		so = dl.Load(
			dl.Custom("/opt/xgboost/lib/libxgboost.so"),
			dl.Cached("dl/go-ml/libxgboost.so"),
			dl.System("libxgboost.so"),
			dl.LzmaExternal("https://github.com/sudachen/xgboost/releases/download/custom/libxgboost_cpu_lin64.lzma"),
			dlVerbose)
	} else if runtime.GOOS == "windows" && runtime.GOARCH == "amd64" {
		so = dl.Load(
			dl.Cached("dl/go-ml/xgboost.dll"),
			dl.System("xgboost.dll"),
			dl.LzmaExternal("https://github.com/sudachen/xgboost/releases/download/custom/libxgboost_cpu_win64.lzma"),
			dlVerbose)
	} else {
		panic("unsupported platfrom")
	}

	so.Bind("XGBoostVersion", unsafe.Pointer(&C._godl_XGBoostVersion))
	so.Bind("XGBGetLastError", unsafe.Pointer(&C._godl_XGBGetLastError))
	so.Bind("XGDMatrixCreateFromFile", unsafe.Pointer(&C._godl_XGDMatrixCreateFromFile))
	so.Bind("XGDMatrixCreateFromDT", unsafe.Pointer(&C._godl_XGDMatrixCreateFromDT))
	so.Bind("XGDMatrixCreateFromMat", unsafe.Pointer(&C._godl_XGDMatrixCreateFromMat))
	so.Bind("XGDMatrixFree", unsafe.Pointer(&C._godl_XGDMatrixFree))
	so.Bind("XGDMatrixSaveBinary", unsafe.Pointer(&C._godl_XGDMatrixSaveBinary))
	so.Bind("XGDMatrixNumRow", unsafe.Pointer(&C._godl_XGDMatrixNumRow))
	so.Bind("XGDMatrixNumCol", unsafe.Pointer(&C._godl_XGDMatrixNumCol))
	so.Bind("XGBoosterCreate", unsafe.Pointer(&C._godl_XGBoosterCreate))
	so.Bind("XGBoosterFree", unsafe.Pointer(&C._godl_XGBoosterFree))
	so.Bind("XGBoosterSetParam", unsafe.Pointer(&C._godl_XGBoosterSetParam))
	so.Bind("XGBoosterLoadModel", unsafe.Pointer(&C._godl_XGBoosterLoadModel))
	so.Bind("XGBoosterSaveModel", unsafe.Pointer(&C._godl_XGBoosterSaveModel))
	so.Bind("XGBoosterLoadModelFromBuffer", unsafe.Pointer(&C._godl_XGBoosterLoadModelFromBuffer))
	so.Bind("XGBoosterGetModelRaw", unsafe.Pointer(&C._godl_XGBoosterGetModelRaw))
	so.Bind("XGBoosterSaveJsonConfig", unsafe.Pointer(&C._godl_XGBoosterSaveJsonConfig))
	so.Bind("XGBoosterLoadJsonConfig", unsafe.Pointer(&C._godl_XGBoosterLoadJsonConfig))
	so.Bind("XGBoosterGetAttr", unsafe.Pointer(&C._godl_XGBoosterGetAttr))
	so.Bind("XGBoosterSetAttr", unsafe.Pointer(&C._godl_XGBoosterSetAttr))
	so.Bind("XGBoosterGetAttrNames", unsafe.Pointer(&C._godl_XGBoosterGetAttrNames))
	so.Bind("XGBoosterBoostOneIter", unsafe.Pointer(&C._godl_XGBoosterBoostOneIter))
	so.Bind("XGBoosterUpdateOneIter", unsafe.Pointer(&C._godl_XGBoosterUpdateOneIter))
	so.Bind("XGBoosterEvalOneIter", unsafe.Pointer(&C._godl_XGBoosterEvalOneIter))
	so.Bind("XGBoosterPredict", unsafe.Pointer(&C._godl_XGBoosterPredict))
	so.Bind("XGDMatrixSetFloatInfo", unsafe.Pointer(&C._godl_XGDMatrixSetFloatInfo))
	so.Bind("XGDMatrixSetUIntInfo", unsafe.Pointer(&C._godl_XGDMatrixSetUIntInfo))
	so.Bind("XGDMatrixGetFloatInfo", unsafe.Pointer(&C._godl_XGDMatrixGetFloatInfo))
	so.Bind("XGDMatrixGetUIntInfo", unsafe.Pointer(&C._godl_XGDMatrixGetUIntInfo))

	var major, minor, patch C.int
	C.XGBoostVersion(&major, &minor, &patch)
	LibVersion = mlutil.MakeVersion(int(major), int(minor), int(patch))
}

func Create() unsafe.Pointer {
	var q C.BoosterHandle
	if e := C.XGBoosterCreate(nil, 0, &q); e != 0 {
		s := C.GoString(C.XGBGetLastError())
		panic(xerrors.Errorf("xgbooster error: " + s))
	}
	return unsafe.Pointer(q)
}

func Create2(m ...unsafe.Pointer) unsafe.Pointer {
	var q C.BoosterHandle
	if e := C.XGBoosterCreate((*C.DMatrixHandle)(unsafe.Pointer(&m[0])), C.ulong(len(m)), &q); e != 0 {
		s := C.GoString(C.XGBGetLastError())
		panic(xerrors.Errorf("xgbooster error: " + s))
	}
	runtime.KeepAlive(m)
	return unsafe.Pointer(q)
}

func Close(h unsafe.Pointer) {
	if e := C.XGBoosterFree(C.BoosterHandle(h)); e != 0 {
		s := C.GoString(C.XGBGetLastError())
		panic(xerrors.Errorf("xgbooster error: " + s))
	}
}

func SetParam(b unsafe.Pointer, par, val string) {
	p := C.CString(par)
	defer C.free(unsafe.Pointer(p))
	v := C.CString(val)
	defer C.free(unsafe.Pointer(v))
	if e := C.XGBoosterSetParam(C.BoosterHandle(b), p, v); e != 0 {
		s := C.GoString(C.XGBGetLastError())
		panic(xerrors.Errorf("xgbooster error: " + s))
	}
}

func Matrix(data []float32, row, col int) unsafe.Pointer {
	var q C.DMatrixHandle
	if e := C.XGDMatrixCreateFromMat((*C.float)(&data[0]), C.ulong(row), C.ulong(col), 0, &q); e != 0 {
		s := C.GoString(C.XGBGetLastError())
		panic(xerrors.Errorf("xgbooster error: " + s))
	}
	runtime.KeepAlive(data)
	return unsafe.Pointer(q)
}

func Free(matrix unsafe.Pointer) {
	C.XGDMatrixSaveBinary(C.DMatrixHandle(matrix), C.CString("matrix2.txt"), 0)
	if e := C.XGDMatrixFree(C.DMatrixHandle(matrix)); e != 0 {
		s := C.GoString(C.XGBGetLastError())
		panic(xerrors.Errorf("xgbooster error: " + s))
	}
}

var strLabel = C.CString("label")
var strWeight = C.CString("weight")

func SetInfo(matrix unsafe.Pointer, name string, dat interface{}) {
	var n *C.char
	switch name {
	case "label":
		n = strLabel
	case "weight":
		n = strWeight
	default:
		n = C.CString(name)
		defer C.free(unsafe.Pointer(n))
	}
	var e C.int
	var p unsafe.Pointer
	if v, ok := dat.([]int); ok {
		p = unsafe.Pointer(&v[0])
		e = C.XGDMatrixSetUIntInfo(C.DMatrixHandle(matrix), n, (*C.uint)(p), C.ulong(len(v)))
	} else if v, ok := dat.([]float32); ok {
		p = unsafe.Pointer(&v[0])
		e = C.XGDMatrixSetFloatInfo(C.DMatrixHandle(matrix), n, (*C.float)(p), C.ulong(len(v)))
	} else {
		panic("dat must be []int or []float32")
	}
	if e != 0 {
		s := C.GoString(C.XGBGetLastError())
		panic(xerrors.Errorf("xgbooster error: " + s))
	}
	runtime.KeepAlive(dat)
}

func GetInfo(matrix unsafe.Pointer, name string, dat interface{}) interface{} {
	var n *C.char
	switch name {
	case "label":
		n = strLabel
	case "weight":
		n = strWeight
	default:
		n = C.CString(name)
		defer C.free(unsafe.Pointer(n))
	}
	var e C.int
	var p, x unsafe.Pointer
	var ln C.ulong
	if v, ok := dat.([]int); ok {
		p = unsafe.Pointer(&x)
		e = C.XGDMatrixGetUIntInfo(C.DMatrixHandle(matrix), n, &ln, (**C.uint)(p))
		if e != 0 {
			s := C.GoString(C.XGBGetLastError())
			panic(xerrors.Errorf("xgbooster error: " + s))
		}
		if len(v) < int(ln) {
			v = make([]int, int(ln))
		}
		C.memcpy(unsafe.Pointer(&v[0]), x, ln*C.ulong(unsafe.Sizeof(int(0))))
		return v
	} else if v, ok := dat.([]float32); ok {
		p = unsafe.Pointer(&x)
		e = C.XGDMatrixGetFloatInfo(C.DMatrixHandle(matrix), n, &ln, (**C.float)(p))
		if e != 0 {
			s := C.GoString(C.XGBGetLastError())
			panic(xerrors.Errorf("xgbooster error: " + s))
		}
		if len(v) < int(ln) {
			v = make([]float32, int(ln))
		}
		C.memcpy(unsafe.Pointer(&v[0]), x, ln*C.ulong(4))
		return v
	} else {
		panic("dat must be []int or []float32")
	}
}

func Update(b unsafe.Pointer, iter int, matrix unsafe.Pointer) {
	if e := C.XGBoosterUpdateOneIter(C.BoosterHandle(b), C.int(iter), C.DMatrixHandle(matrix)); e != 0 {
		s := C.GoString(C.XGBGetLastError())
		panic(xerrors.Errorf("xgbooster error: " + s))
	}
}

func Predict(b, matrix unsafe.Pointer, limit int) (result []float32) {
	ln := C.ulong(0)
	dt := unsafe.Pointer(nil)
	if e := C.XGBoosterPredict(
		C.BoosterHandle(b), C.DMatrixHandle(matrix),
		C.int(0), C.uint(limit), C.int(0),
		(*C.ulong)(&ln), unsafe.Pointer(&dt)); e != 0 {
		s := C.GoString(C.XGBGetLastError())
		panic(xerrors.Errorf("xgbooster error: " + s))
	}
	dtlen := uintptr(ln)
	result = make([]float32, dtlen)
	for i := range result {
		result[i] = float32(*(*C.float)(unsafe.Pointer((uintptr(dt) + uintptr(4*i)))))
	}
	return
}
