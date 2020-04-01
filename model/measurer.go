package model

import (
	"github.com/sudachen/go-ml/fu"
	"github.com/sudachen/go-ml/tables"
	"reflect"
)

type Measurer []Metrics

func (metrics Measurer) Iterate(iteration int, subset string, result, label *tables.Column) (fu.Struct, bool) {
	metrics.Begin()
	metrics.ColumnUpdate(result, label)
	return metrics.Complete(iteration, subset)
}

func (metrics Measurer) Begin() Measurer {
	for _, m := range metrics {
		m.Begin()
	}
	return metrics
}

func (metrics Measurer) Copy() Measurer {
	r := make([]Metrics, len(metrics))
	for i, m := range metrics {
		r[i] = m.Copy()
	}
	return r
}

func (metrics Measurer) ColumnUpdate(result, label *tables.Column) {
	for _, m := range metrics {
		rc, _ := result.Raw()
		lc, _ := label.Raw()
		for i := 0; i < rc.Len(); i++ {
			m.Update(rc.Index(i), lc.Index(i))
		}
	}
}

func (metrics Measurer) Update(result, label reflect.Value) {
	for _, m := range metrics {
		m.Update(result, label)
	}
}

func (metrics Measurer) Complete(iteration int, subset string) (fu.Struct, bool) {
	line := fu.MakeStruct([]string{"Iteration", "Subset"}, iteration, subset)
	done := false
	for _, m := range metrics {
		r, ok := m.Complete()
		line = line.With(r)
		done = done || ok
	}
	return line, done
}
