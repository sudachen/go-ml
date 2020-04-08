package xgb

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/sudachen/go-iokit/iokit"
	"github.com/sudachen/go-ml/fu"
	"github.com/sudachen/go-ml/fu/lazy"
	"github.com/sudachen/go-ml/model"
	"github.com/sudachen/go-ml/tables"
	"github.com/sudachen/go-ml/xgb/capi"
	"github.com/sudachen/go-zorros/zorros"
	"io"
	"unsafe"
)

func fit(rounds int, e Model, dataset model.Dataset, output iokit.Output, metricsf model.Metrics, scoref model.Score) (report model.Report, err error) {
	historylen := metricsf.HistoryLength()
	stash := model.NewStash(historylen, "xgb-model-stash-*")
	defer stash.Close()

	t, err := dataset.Source.Collect()
	if err != nil {
		return
	}
	Test := fu.Fnzs(dataset.Test, "Test")
	if fu.IndexOf(Test, t.Names()) < 0 {
		err = zorros.Errorf("dataset does not have column `%v`", Test)
		return
	}

	features := t.OnlyNames(dataset.Features...)
	test, train, err := t.MatrixWithLabelIf(features, dataset.Label, dataset.Test, true)
	if err != nil {
		return
	}

	m := matrix32(train)
	defer m.Free()
	m2 := matrix32(test)
	defer m2.Free()

	predicts := fu.Fnzs(e.Predicted, "Predicted")

	xgb := &xgbinstance{capi.Create2(m.handle, m2.handle), features, predicts}
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
			panic(zorros.Errorf("labels don't contain enough classes or label values is incorrect"))
		}
		capi.SetParam(xgb.handle, "num_class", fmt.Sprint(x+1))
	}

	perflog := [][2]fu.Struct{}
	scorlog := []float64{}
	testLabels := test.AsLabelColumn()
	trainLabels := train.AsLabelColumn()
	rounds = fu.Maxi(rounds, 1)
	var o iokit.Output

	for i := 0; i < rounds; i++ {
		capi.Update(xgb.handle, i, m.handle)
		done := false
		lr := [2]fu.Struct{}
		lr[0], _ = xgb.evalMetrics(i, model.Train, m.handle, trainLabels, metricsf)
		lr[1], done = xgb.evalMetrics(i, model.Test, m2.handle, testLabels, metricsf)
		score := scoref(lr[0], lr[1])
		//verbose.Printf("fit [%3d] train %v",i,lr[0])
		//verbose.Printf("fit [%3d] test %v",i,lr[0])
		//verbose.Printf("fit [%3d] score %v",i,score)
		if len(scorlog) > historylen {
			q := len(scorlog) - historylen
			if fu.Maxd(0, scorlog[:q]...) >= fu.Maxd(score, scorlog[q:]...) {
				break
			}
		}
		if o, err = stash.Output(i); err != nil {
			return
		}
		if err = model.Memorize(o, model.MemorizeMap{"model": mnemosyne{xgb}}); err != nil {
			return
		}
		scorlog = append(scorlog, score)
		perflog = append(perflog, lr)
		if done {
			break
		}
	}

	if len(perflog) > 0 {
		report.History = tables.Lazy(lazy.Flatn(perflog)).LuckyCollect()
		j := len(scorlog) - stash.Length()
		i := fu.Indmaxd(scorlog[j:]) + j
		report.Train = perflog[i][0]
		report.Test = perflog[i][1]
		report.Score = scorlog[i]
		if output != nil {
			rd, e := stash.Reader(i)
			if e != nil {
				err = zorros.Trace(e)
				return
			}
			wh, e := output.Create()
			if e != nil {
				err = zorros.Trace(e)
				return
			}
			defer wh.End()
			_, e = io.Copy(wh, rd)
			if e != nil {
				err = zorros.Trace(e)
				return
			}
			if e = wh.Commit(); e != nil {
				err = zorros.Trace(e)
				return
			}
		}
	} else {
		report.History = tables.NewEmpty(metricsf.Names(), nil)
	}
	return
}

func (xgb *xgbinstance) evalMetrics(i int, subset string, m unsafe.Pointer, labels *tables.Column, metricsf model.Metrics) (fu.Struct, bool) {
	y := capi.Predict(xgb.handle, m, 0)
	pred := tables.Matrix{
		Features:    y,
		Labels:      nil,
		Width:       len(y) / labels.Len(),
		Length:      labels.Len(),
		LabelsWidth: 0,
	}
	return model.EvaluateMetrics(i, subset, pred.AsColumn(), labels, metricsf)
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
