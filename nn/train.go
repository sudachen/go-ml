package nn

import (
	"github.com/sudachen/go-ml/fu"
	"github.com/sudachen/go-ml/model"
	"github.com/sudachen/go-ml/tables"
	"github.com/sudachen/go-zorros/zorros"
	"reflect"
)

func train(e Model, dataset model.Dataset, w model.Workout) (report *model.Report, err error) {
	t, err := dataset.Source.Lazy().First(1).Collect()
	if err != nil {
		return
	}

	features := t.OnlyNames(dataset.Features...)

	Test := fu.Fnzs(dataset.Test, model.TestCol)
	if fu.IndexOf(Test, t.Names()) < 0 {
		err = zorros.Errorf("dataset does not have column `%v`", Test)
		return
	}

	Label := fu.Fnzs(dataset.Label,model.LabelCol)
	if fu.IndexOf(Label, t.Names()) < 0 {
		err = zorros.Errorf("dataset does not have column `%v`", Label)
		return
	}

	predicts := fu.Fnzs(e.Predicted, model.PredictedCol)

	network := New(e.Context.Upgrade(), e.Network, e.Input, e.Loss, e.BatchSize, e.Seed)
	train := dataset.Source.Lazy().IfNotFlag(dataset.Test).Batch(e.BatchSize).Parallel()
	full := dataset.Source.Lazy().Batch(e.BatchSize).Parallel()
	out := make([]float32, network.Graph.Output.Dim().Total())

	for done := false; w != nil && !done ; w = w.Next() {
		opt := e.Optimizer.Init(w.Iteration())

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

		trainmu := w.TrainMetrics()
		testmu := w.TestMetrics()
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
					testmu.Update(resultCol.Value(i), labelCol.Value(i))
				} else {
					trainmu.Update(resultCol.Value(i), labelCol.Value(i))
				}
			}
			return nil
		}); err != nil {
			return
		}

		lr0, _ := trainmu.Complete()
		lr1, d := testmu.Complete()
		memorize := model.MemorizeMap{"model":  mnemosyne{network, features, predicts}}
		if report, done, err = w.Complete(memorize, lr0, lr1, d); err != nil {
			return nil, zorros.Wrapf(err, "tailed to complete model: %s", err.Error())
		}
	}

	return
}

