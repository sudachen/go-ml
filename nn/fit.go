package nn

import (
	"github.com/sudachen/go-iokit/iokit"
	"github.com/sudachen/go-ml/fu"
	"github.com/sudachen/go-ml/fu/lazy"
	"github.com/sudachen/go-ml/fu/verbose"
	"github.com/sudachen/go-ml/model"
	"github.com/sudachen/go-ml/tables"
	"github.com/sudachen/go-zorros/zorros"
	"gopkg.in/yaml.v3"
	"io"
	"reflect"
)

func fit(iteratins int, e Model, dataset model.Dataset, output iokit.Output, metricsf model.Metrics, scoref model.Score) (report model.Report, err error) {
	historylen := metricsf.HistoryLength()
	stash := model.NewStash(historylen, "nn-model-stash-*.zip")
	t, err := dataset.Source.Lazy().First(1).Collect()
	if err != nil {
		return
	}
	features := t.OnlyNames(dataset.Features...)
	predicts := fu.Fnzs(e.Predicted, "Predicted")
	Test := fu.Fnzs(dataset.Test, "Test")
	if fu.IndexOf(Test, t.Names()) < 0 {
		err = zorros.Errorf("dataset does not have column `%v`", Test)
		return
	}
	network := New(e.Context.Upgrade(), e.Network, e.Input, e.Loss, e.BatchSize, e.Seed)
	train := dataset.Source.Lazy().IfNotFlag(dataset.Test).Batch(e.BatchSize).Parallel()
	full := dataset.Source.Lazy().Batch(e.BatchSize).Parallel()
	perflog := [][2]fu.Struct{}
	scorlog := []float64{}
	out := make([]float32, network.Graph.Output.Dim().Total())

	for i := 0; i < iteratins; i++ {
		opt := e.Optimizer.Init(i)
		if err = train.Drain(func(value reflect.Value) error {
			if value.Kind() == reflect.Bool {
				return nil
			}
			t := value.Interface().(*tables.Table)
			m, err := t.MatrixWithLabel(features, dataset.Label, e.BatchSize)
			if err != nil {
				return err
			}
			network.Train(m.Features, m.Labels, opt)
			return nil
		}); err != nil {
			return
		}
		trainmr := metricsf.New(i, model.Train)
		testmr := metricsf.New(i, model.Test)
		if err = full.Drain(func(value reflect.Value) error {
			if value.Kind() == reflect.Bool {
				return nil
			}
			t := value.Interface().(*tables.Table)
			m, err := t.MatrixWithLabel(features, dataset.Label, e.BatchSize)
			if err != nil {
				return err
			}
			network.Forward(m.Features, out)
			resultCol := tables.MatrixColumn(out, e.BatchSize)
			labelCol := t.Col(dataset.Label)
			for i, c := range t.Col(dataset.Test).ExtractAs(fu.Bool, true).([]bool) {
				if c {
					testmr.Update(resultCol.Value(i), labelCol.Value(i))
				} else {
					trainmr.Update(resultCol.Value(i), labelCol.Value(i))
				}
			}
			return nil
		}); err != nil {
			return
		}
		done := false
		lr := [2]fu.Struct{}
		lr[0], _ = trainmr.Complete()
		lr[1], done = testmr.Complete()
		score := scoref(lr[0], lr[1])
		verbose.Printf("fit [%3d] train %v", i, lr[0])
		verbose.Printf("fit [%3d] test %v", i, lr[0])
		verbose.Printf("fit [%3d] score %v", i, score)
		if len(scorlog) > historylen {
			q := len(scorlog) - historylen
			if fu.Maxd(0, scorlog[:q]...) >= fu.Maxd(score, scorlog[q:]...) {
				break
			}
		}
		o, e := stash.Output(i)
		if e != nil {
			err = zorros.Trace(e)
			return
		}
		e = model.Memorize(o,
			model.MemorizeMap{
				"model": mnemosyne{network, features, predicts}})
		if e != nil {
			err = zorros.Trace(e)
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

type mnemosyne struct {
	network  *Network
	features []string
	predicts string
}

func (mm mnemosyne) Memorize(c *model.CollectionWriter) (err error) {
	if err = c.Add("info.yaml", func(wr io.Writer) error {
		en := yaml.NewEncoder(wr)
		return en.Encode(map[string]interface{}{
			"features": mm.features,
			"predicts": mm.predicts,
		})
	}); err != nil {
		return
	}
	if err = c.AddLzma2("params.bin.xz", func(wr io.Writer) (e error) {
		return mm.network.SaveParams(iokit.Writer(wr))
	}); err != nil {
		return
	}
	if err = c.AddLzma2("symbol.yaml.xz", func(wr io.Writer) (e error) {
		return mm.network.SaveSymbol(iokit.Writer(wr))
	}); err != nil {
		return
	}
	return
}
