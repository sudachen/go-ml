package xgb

import (
	"fmt"
	"github.com/sudachen/go-ml/xgb/capi"
)

type xgbparam interface{ name() string }
type capiparam interface{ pair() (string, string) }
type optparam interface{ opt() interface{} }

func (x XGBoost) setparam(par capiparam) {
	n, v := par.pair()
	capi.SetParam(x.handle, n, v)
}

type objective string

const Linear = objective("reg:linear")
const SquareLinear = objective("reg:squarederror")
const Logistic = objective("reg:logistic")
const SqureLogistic = objective("reg:squaredlogerror")
const Tweedie = objective("reg:tweedie")
const Binary = objective("binary:logistic")
const RawBinary = objective("binary:logitraw")
const HingeBinary = objective("binary:hinge")

// gamma regression with log-link. Output is a mean of gamma distribution.
// It might be useful, e.g., for modeling insurance claims severity,
// or for any outcome that might be gamma-distributed.
const GammaRegress = objective("reg:gamma")

// set XGBoost to do multiclass classification using the softmax objective,
// you also need to set num_class(number of classes)
const Softmax = objective("multi:softmax")

// same as softmax, but output a vector of ndata * nclass,
// which can be further reshaped to ndata * nclass matrix.
// The result contains predicted probability of each data point belonging to each class.
const Softprob = objective("multi:softprob")

func (o objective) pair() (string, string) { return "objective", string(o) }
func (o objective) name() string           { return "objective" }

type Param struct{ Name, Value string }

func (sp Param) pair() (string, string) { return sp.Name, sp.Value }
func (sp Param) name() string           { return sp.Name }

func LearnRate(v float64) xgbparam { return Param{"learning_rate", fmt.Sprint(v)} }

type OptParam struct {
	Name  string
	Value interface{}
}

func (sp OptParam) name() string     { return sp.Name }
func (sp OptParam) opt() interface{} { return sp.Value }

type Rounds int

func (Rounds) name() string { return "Rounds" }

type ResultName string

func (ResultName) name() string { return "ResultName" }

func MaxDepth(value int) xgbparam    { return Param{"max_depth", fmt.Sprint(value)} }
func Nestimators(value int) xgbparam { return Param{"n_estimators", fmt.Sprint(value)} }

//func Classnum(v int) xgbparam { return Param{"num_class",fmt.Sprint(v)} }
