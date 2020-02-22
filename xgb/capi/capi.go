package capi

/*
#include "capi.h"
*/
import "C"
import (
	"fmt"
	"github.com/sudachen/go-dl/dl"
	"github.com/sudachen/go-ml/util"
	"runtime"
	"unsafe"
)

var LibVersion util.VersionType

func init() {
	var so dl.SO
	dlVerbose := dl.Verbose(func(text string, verbosity int){
		if verbosity < 2 {
			// verbosity 2 relates to detailed information
			fmt.Println(text)
		}
	})
	if runtime.GOOS == "linux" && runtime.GOARCH == "amd64"{
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

	so.Bind("XGBoostVersion",unsafe.Pointer(&C._godl_XGBoostVersion))
	so.Bind("XGBGetLastError",unsafe.Pointer(&C._godl_XGBGetLastError))
	so.Bind("XGDMatrixCreateFromFile",unsafe.Pointer(&C._godl_XGDMatrixCreateFromFile))
	so.Bind("XGDMatrixCreateFromDT",unsafe.Pointer(&C._godl_XGDMatrixCreateFromDT))
	so.Bind("XGDMatrixFree",unsafe.Pointer(&C._godl_XGDMatrixFree))
	so.Bind("XGDMatrixSaveBinary",unsafe.Pointer(&C._godl_XGDMatrixSaveBinary))
	so.Bind("XGDMatrixNumRow",unsafe.Pointer(&C._godl_XGDMatrixNumRow))
	so.Bind("XGDMatrixNumCol",unsafe.Pointer(&C._godl_XGDMatrixNumCol))
	so.Bind("XGBoosterCreate",unsafe.Pointer(&C._godl_XGBoosterCreate))
	so.Bind("XGBoosterFree",unsafe.Pointer(&C._godl_XGBoosterFree))
	so.Bind("XGBoosterSetParam",unsafe.Pointer(&C._godl_XGBoosterSetParam))
	so.Bind("XGBoosterLoadModel",unsafe.Pointer(&C._godl_XGBoosterLoadModel))
	so.Bind("XGBoosterSaveModel",unsafe.Pointer(&C._godl_XGBoosterSaveModel))
	so.Bind("XGBoosterLoadModelFromBuffer",unsafe.Pointer(&C._godl_XGBoosterLoadModelFromBuffer))
	so.Bind("XGBoosterGetModelRaw",unsafe.Pointer(&C._godl_XGBoosterGetModelRaw))
	so.Bind("XGBoosterSaveJsonConfig",unsafe.Pointer(&C._godl_XGBoosterSaveJsonConfig))
	so.Bind("XGBoosterLoadJsonConfig",unsafe.Pointer(&C._godl_XGBoosterLoadJsonConfig))
	so.Bind("XGBoosterGetAttr",unsafe.Pointer(&C._godl_XGBoosterGetAttr))
	so.Bind("XGBoosterSetAttr",unsafe.Pointer(&C._godl_XGBoosterSetAttr))
	so.Bind("XGBoosterGetAttrNames",unsafe.Pointer(&C._godl_XGBoosterGetAttrNames))
	so.Bind("XGBoosterBoostOneIter",unsafe.Pointer(&C._godl_XGBoosterBoostOneIter))
	so.Bind("XGBoosterUpdateOneIter",unsafe.Pointer(&C._godl_XGBoosterUpdateOneIter))
	so.Bind("XGBoosterEvalOneIter",unsafe.Pointer(&C._godl_XGBoosterEvalOneIter))
	so.Bind("XGBoosterPredict",unsafe.Pointer(&C._godl_XGBoosterPredict))
	so.Bind("XGDMatrixSetFloatInfo",unsafe.Pointer(&C._godl_XGDMatrixSetFloatInfo))
	so.Bind("XGDMatrixSetUIntInfo",unsafe.Pointer(&C._godl_XGDMatrixSetUIntInfo))
	so.Bind("XGDMatrixGetFloatInfo",unsafe.Pointer(&C._godl_XGDMatrixGetFloatInfo))
	so.Bind("XGDMatrixGetUIntInfo",unsafe.Pointer(&C._godl_XGDMatrixGetUIntInfo))

	var major, minor, patch C.int
	C.XGBoostVersion(&major,&minor,&patch);
	LibVersion = util.MakeVersion(int(major),int(minor),int(patch))
}

