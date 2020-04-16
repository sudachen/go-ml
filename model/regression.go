package model

import (
	"github.com/sudachen/go-ml/fu"
	"math"
	"reflect"
)

/*
Regression - the regression metrics factory
*/
type Regression struct {
	Error float64 // error goal
}

/*
New iteration metrics
*/
func (m Regression) New(iteration int, subset string) MetricsUpdater {
	return &rgupdater{
		Regression: m,
		iteration:  iteration,
		subset:     subset,
	}
}

/*
Names is the list of calculating metrics
*/
func (m Regression) Names() []string {
	return []string{
		IterationCol,
		SubsetCol,
		ErrorCol,
		LossCol,
		RmseCol,
		MaeCol,
		TotalCol,
	}
}

type rgupdater struct {
	Regression
	iteration int
	subset    string
	loss      float64
	error     float64 // sum{|result-label|}
	error2    float64 // sum{(result-label)^2}
	count     float64
}

func (m *rgupdater) Complete() (fu.Struct, bool) {
	if m.count > 0 {
		squrederr := m.error2 / m.count
		errsqrt := math.Sqrt(squrederr)
		abserr := m.error / m.count
		columns := []reflect.Value{
			reflect.ValueOf(m.iteration),
			reflect.ValueOf(m.subset),
			reflect.ValueOf(squrederr),
			reflect.ValueOf(m.loss / m.count),
			reflect.ValueOf(errsqrt),
			reflect.ValueOf(abserr),
			reflect.ValueOf(m.count),
		}
		goal := false
		if m.Error > 0 {
			goal = goal || squrederr < m.Error
		}
		return fu.Struct{Names: m.Names(), Columns: columns}, goal
	}
	return fu.
			NaStruct(m.Names(), fu.Float64).
			Set(IterationCol, fu.IntZero).
			Set(SubsetCol, fu.EmptyString),
		false
}

func error1(a, b []float32) float64 {
	c := 0.
	for i, v := range a {
		c += math.Abs(float64(v - b[i]))
	}
	return c / float64(len(a))
}

func error2(a, b []float32) float64 {
	c := 0.
	for i, v := range a {
		q := float64(v - b[i])
		c += q * q
	}
	return c / float64(len(a))
}

func (m *rgupdater) Update(result, label reflect.Value, loss float64) {
	var e, e2 float64
	if result.Type() == fu.TensorType {
		vr := result.Interface().(fu.Tensor).Floats32()
		vl := label.Interface().(fu.Tensor).Floats32()
		e = error1(vr, vl)
		e2 = error2(vr, vl)
	} else {
		r := fu.Cell{result}.Float()
		l := fu.Cell{result}.Float()
		e = math.Abs(r - l)
		e2 = e * e
	}
	m.error += e
	m.error2 += e2
	m.loss += loss
	m.count++
}
