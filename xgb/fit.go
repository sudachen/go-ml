package xgb

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/sudachen/go-foo/fu"
	"github.com/sudachen/go-ml/mlutil"
	"github.com/sudachen/go-ml/model"
	"github.com/sudachen/go-ml/tables"
	"github.com/sudachen/go-ml/xgb/capi"
	"golang.org/x/xerrors"
	"io"
	"unsafe"
)

func fit(rounds int, e Model, dataset model.Dataset, output fu.Output, mx ...model.Metrics) (metrics *tables.Table, err error) {
	t, err := dataset.Source.Collect()
	if err != nil {
		return
	}

	features := t.OnlyNames(dataset.Features...)
	// test if t.Col(dataset.Train) == true otherwise it's train
	test, train, err := t.MatrixWithLabelIf(features, dataset.Label, dataset.Test, true)
	if err != nil {
		return
	}

	m := matrix32(train)
	defer m.Free()
	m2 := matrix32(test)
	defer m2.Free()

	predicts := fu.Fnzs(e.Predicted, "Predicted")

	var xgb *xgbinstance
	if test.Length > 0 {
		xgb = &xgbinstance{capi.Create2(m.handle, m2.handle), features, predicts}
	} else {
		xgb = &xgbinstance{capi.Create2(m.handle), features, predicts}
	}
	defer xgb.Close()

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

	perflog := []*mlutil.Struct{}
	var testLabels, trainLabels *tables.Column
	if len(mx) > 0 {
		testLabels = test.AsLabelColumn()
		trainLabels = train.AsLabelColumn()
	}

	for i := 0; i < rounds; i++ {
		capi.Update(xgb.handle, i, m.handle)
		if len(mx) > 0 {
			done := false
			xgb.evalMetrics(i, false, m.handle, trainLabels, &perflog, model.Measurer(mx))
			if test.Length > 0 {
				done = xgb.evalMetrics(i, true, m2.handle, testLabels, &perflog, model.Measurer(mx))
			}
			if done {
				break
			}
		}
	}

	if len(perflog) > 0 {
		metrics = tables.New(perflog)
	}

	err = model.Memorize(output, model.MemorizeMap{"model": mnemosyne{xgb}})
	return
}

func (xgb *xgbinstance) evalMetrics(i int, testSubset bool, m unsafe.Pointer, labels *tables.Column, log *[]*mlutil.Struct, mr model.Measurer) bool {
	y := capi.Predict(xgb.handle, m, 0)
	pred := tables.Matrix{
		Features:    y,
		Labels:      nil,
		Width:       len(y) / labels.Len(),
		Length:      labels.Len(),
		LabelsWidth: 0,
	}
	subset := fu.Ifes(testSubset, "test", "train")
	line, done := mr.Iterate(i, subset, pred.AsColumn(), labels)
	*log = append(*log, &line)
	return done
}

type mnemosyne struct{ *xgbinstance }

func (x mnemosyne) Memorize(c *model.CollectionWriter) (err error) {
	if err = c.Add("info.json", func(wr io.Writer) error {
		en := json.NewEncoder(wr)
		return en.Encode(map[string]interface{}{
			"features": x.features,
			"predicts": x.predicts,
		})
	}); err != nil {
		return
	}
	if err = c.Add("config.json", func(wr io.Writer) error {
		_, err := wr.Write(capi.JsonConfig(x.handle))
		return err
	}); err != nil {
		return
	}
	if err = c.AddLzma2("model.bin.xz", func(wr io.Writer) error {
		_, err := wr.Write(capi.GetModel(x.handle))
		return err
	}); err != nil {
		return
	}
	if err = c.AddLzma2("dump.txt.xz", func(wr io.Writer) error {
		w := bufio.NewWriter(wr)
		for _, s := range capi.DumpModel(x.handle) {
			_, err := w.WriteString(s)
			if err != nil {
				return err
			}
		}
		return w.Flush()
	}); err != nil {
		return
	}
	return
}
