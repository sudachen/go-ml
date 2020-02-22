package xgb

import (
	"github.com/sudachen/go-ml/util"
	"github.com/sudachen/go-ml/xgb/capi"
)

func LibVersion() util.VersionType {
	return util.VersionType(capi.LibVersion)
}

