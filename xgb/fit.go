package xgb

import (
	"fmt"
	"github.com/sudachen/go-foo/fu"
	"github.com/sudachen/go-ml/ml"
	"github.com/sudachen/go-ml/tables"
	"github.com/sudachen/go-ml/xgb/capi"
	"golang.org/x/xerrors"
)

func fit(e Model, dataset ml.Dataset, opts ...ml.Fitparam) (xgb xgbinstance, err error) {
	t,err := tables.Lazy(dataset.Source).Collect()
	features := t.OnlyNames(dataset.Features...)
	train, test, err := t.FillTrainAndTest(features, dataset.Label, dataset.Test,true)
	//train, err := t.FillTrain(features,dataset.Label)

	if err != nil {
		return
	}
	m := matrix32(train)
	defer m.Free()

	predicts := fu.Option(ml.Result(fu.Fnzs(e.Result, "Result")),opts).String()

	if test.Length > 0 {
		m2 := matrix32(test)
		defer m2.Free()
		xgb = xgbinstance{capi.Create2(m.handle, m2.handle), features, predicts}
	} else {
		xgb = xgbinstance{capi.Create2(m.handle), features, predicts}
	}

	if e.Algorithm != booster("") {
		xgb.setparam(e.Algorithm)
	}

	if e.Function != objective("") {
		xgb.setparam(e.Function)
	}

	if e.Estimators != 0 {
		capi.SetParam(xgb.handle, "n_estimators", fmt.Sprint(e.Estimators))
	}

	if e.LearningRate != 0 {
		capi.SetParam(xgb.handle, "eta", fmt.Sprint(e.LearningRate))
	}

	if e.MaxDepth != 0 {
		capi.SetParam(xgb.handle, "max_depth", fmt.Sprint(e.MaxDepth))
	}

	capi.SetParam(xgb.handle, "num_feature", fmt.Sprint(len(features)))
	if e.Function == Softprob || e.Function == Softmax {
		x := int(fu.Maxr(fu.Maxr(0, train.Labels...), test.Labels...))
		if x < 0 {
			panic(xerrors.Errorf("labels don't contain enough classes or label values is incorrect"))
		}
		capi.SetParam(xgb.handle, "num_class", fmt.Sprint(x+1))
	}

	rounds := int(fu.Option(ml.Iterations(e.Iterations),opts).Int())
	for i := 0; i < rounds; i++ {
		capi.Update(xgb.handle, i, m.handle)
	}
	return
}
