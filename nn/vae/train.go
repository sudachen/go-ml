package vae

import (
	"github.com/sudachen/go-ml/fu"
	"github.com/sudachen/go-ml/model"
	"github.com/sudachen/go-ml/nn"
	"github.com/sudachen/go-ml/nn/mx"
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

	if e.Width <= 0 {
		m, er := t.Matrix(features, 1)
		if er != nil {
			return nil, zorros.Wrapf(er, "failed to get features matrix: %s", er.Error())
		}
		e.Width = m.Width
	}

	if e.Optimizer == nil {
		e.Optimizer = &nn.Adam{Lr: .001}
	}

	if e.BatchSize <= 0 {
		e.BatchSize = DefaultBatchSize
	}

	network := nn.New(
		e.Context.Upgrade(),
		&nn.Lambda{e.autoencoder},
		mx.Dim(e.Width),
		nn.LossFunc(e.loss),
		e.BatchSize,
		e.Seed)

	//network.PrintSummary(true)

	memorize := e.modelmap(network, features)
	train := dataset.Source.Lazy().IfNotFlag(dataset.Test).Batch(e.BatchSize).Parallel()
	full := dataset.Source.Lazy().Batch(e.BatchSize).Parallel()
	out := make([]float32, network.Graph.Output.Dim().Total())
	loss := make([]float32, network.Graph.Loss.Dim().Total())

	for done := false; w != nil && !done; w = w.Next() {
		opt := e.Optimizer.Init(w.Iteration())

		network.Params["_sampling"].Ones()

		if err = train.Drain(func(value reflect.Value) error {
			if value.Kind() == reflect.Bool {
				return nil
			}
			t := value.Interface().(*tables.Table)
			m, err := t.Matrix(features, e.BatchSize)
			if err != nil {
				return err
			}
			network.Train(m.Features, nil, opt)
			return nil
		}); err != nil {
			return
		}

		network.Params["_sampling"].Zeros()

		trainmu := w.TrainMetrics()
		testmu := w.TestMetrics()
		if err = full.Drain(func(value reflect.Value) error {
			if value.Kind() == reflect.Bool {
				return nil
			}
			t := value.Interface().(*tables.Table)
			m, err := t.Matrix(features, e.BatchSize)
			if err != nil {
				return err
			}
			network.Forward(m.Features, out)
			network.Loss.CopyValuesTo(loss)
			lc := m.AsColumn()
			rt := tables.MatrixColumn(out, e.BatchSize)
			for i, c := range t.Col(dataset.Test).ExtractAs(fu.Bool, true).([]bool) {
				if c {
					testmu.Update(rt.Value(i), lc.Value(i), float64(loss[i]))
				} else {
					trainmu.Update(rt.Value(i), lc.Value(i), float64(loss[i]))
				}
			}
			return nil
		}); err != nil {
			return
		}

		lr0, _ := trainmu.Complete()
		lr1, d := testmu.Complete()
		if report, done, err = w.Complete(memorize, lr0, lr1, d); err != nil {
			return nil, zorros.Wrapf(err, "tailed to complete model: %s", err.Error())
		}
	}

	return
}
