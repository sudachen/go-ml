package nn

import (
	"github.com/sudachen/go-foo/fu"
	"github.com/sudachen/go-ml/mlutil"
	"github.com/sudachen/go-ml/model"
	"github.com/sudachen/go-ml/tables"
	"gopkg.in/yaml.v3"
	"io"
	"reflect"
)

func fit(iteratins int, e Model, dataset model.Dataset, output fu.Output, mx ...model.Metrics) (metrics *tables.Table, err error) {
	t, err := dataset.Source.Lazy().First(1).Collect()
	if err != nil {
		return
	}
	features := t.OnlyNames(dataset.Features...)
	predicts := fu.Fnzs(e.Predicted, "Predicted")

	network := New(e.Context.Upgrade(), e.Network, e.Input, e.Loss, e.BatchSize, e.Seed)
	train := dataset.Source.Lazy().IfNotFlag(dataset.Test).Batch(e.BatchSize).Parallel()
	full := dataset.Source.Lazy().Batch(e.BatchSize).Parallel()
	trainmr := model.Measurer(mx)
	testmr := trainmr.Copy()
	mrlines := []mlutil.Struct{}
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
		trainmr.Begin()
		testmr.Begin()
		if len(mx) > 0 {
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
				if dataset.Test != "" {
					for i, c := range t.Col(dataset.Test).ExtractAs(mlutil.Bool, true).([]bool) {
						if c {
							testmr.Update(resultCol.Value(i), labelCol.Value(i))
						} else {
							trainmr.Update(resultCol.Value(i), labelCol.Value(i))
						}
					}
				} else {
					trainmr.ColumnUpdate(resultCol, labelCol)
				}
				return nil
			}); err != nil {
				return
			}
		}
		line, _ := trainmr.Complete(i, "train")
		mrlines = append(mrlines, line)
		line, done := testmr.Complete(i, "test")
		mrlines = append(mrlines, line)
		if done {
			break
		}
	}

	metrics = tables.New(mrlines)
	err = model.Memorize(output, model.MemorizeMap{"model": mnemosyne{network, features, predicts}})
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
		return mm.network.SaveParams(fu.Writer(wr))
	}); err != nil {
		return
	}
	if err = c.AddLzma2("symbol.yaml.xz", func(wr io.Writer) (e error) {
		return mm.network.SaveSymbol(fu.Writer(wr))
	}); err != nil {
		return
	}
	return
}
