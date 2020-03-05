package xgb

import (
	"github.com/sudachen/go-ml/mlutil"
	"github.com/sudachen/go-ml/xgb/capi"
	"unsafe"
)

func LibVersion() mlutil.VersionType {
	return mlutil.VersionType(capi.LibVersion)
}

type XGBoost struct {
	handle   unsafe.Pointer
	features map[string]int
	predicts []string
}

func (x XGBoost) Close() (err error) {
	return
}

/*func (x XGBoost) Acquire() (base.Predictor,error) {
	return x, nil
}*/

type Model []xgbparam

func (m Model) Feed(ds mlutil.Dataset) mlutil.FatModel {
	return func(opts ...interface{}) (mlutil.Predictor, error) {
		return m.Fit(ds, opts...)
	}
}

func GBLinear(par ...xgbparam) Model {
	return append(par, booster("gblinear"))
}

func GBTree(par ...xgbparam) Model {
	return append(par, booster("gbtree"))
}

func Dart(par ...xgbparam) Model {
	return append(par, booster("dart"))
}

type booster string

func (b booster) pair() (string, string) { return "booster", string(b) }
func (b booster) name() string           { return "booster" }
