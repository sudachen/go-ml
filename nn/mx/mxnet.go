package mx

import (
	"github.com/sudachen/go-ml/nn/mx/capi"
	"github.com/sudachen/go-ml/util"
)

const (
	VersionMajor = 1
	VersionMinor = 5
	VersionPatch = 0
)

const Version util.VersionType = VersionMajor*10000 + VersionMinor*100 + VersionPatch

func LibVersion() util.VersionType {
	return util.VersionType(capi.LibVersion)
}

func GpuCount() int {
	return capi.GpuCount
}

func RandomSeed(seed int) {
	capi.RandomSeed(seed)
}
