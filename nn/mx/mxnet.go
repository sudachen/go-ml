package mx

import (
	"github.com/sudachen/go-ml/fu"
	"github.com/sudachen/go-ml/nn/mx/capi"
)

const (
	VersionMajor = 1
	VersionMinor = 5
	VersionPatch = 0
)

const Version fu.VersionType = VersionMajor*10000 + VersionMinor*100 + VersionPatch

func LibVersion() fu.VersionType {
	return fu.VersionType(capi.LibVersion)
}

func GpuCount() int {
	return capi.GpuCount
}

func RandomSeed(seed int) {
	capi.RandomSeed(seed)
}
