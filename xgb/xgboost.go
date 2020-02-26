package xgb

import (
	"github.com/sudachen/go-ml/mlutil"
	"github.com/sudachen/go-ml/xgb/capi"
)

func LibVersion() mlutil.VersionType {
	return mlutil.VersionType(capi.LibVersion)
}
