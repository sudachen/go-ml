package xgboost

import (
	"github.com/sudachen/go-ml/util"
	"github.com/sudachen/go-ml/xgboost/capi"
)

func LibVersion() util.VersionType {
	return util.VersionType(capi.LibVersion)
}

