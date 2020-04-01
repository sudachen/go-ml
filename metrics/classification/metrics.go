package classification

import (
	"github.com/sudachen/go-ml/fu"
	"github.com/sudachen/go-ml/model"
	"reflect"
)

const Accuracy = "Accuracy"
const Sensitivity = "Sensitivity"
const Precision = "Precision"
const F1Score = "F1Score"
const Total = "Total"
const Correct = "Correct"

var Names = []string{
	Accuracy,
	Sensitivity,
	Precision,
	F1Score,
	Correct,
	Total,
}

type Metrics struct {
	correct    float64
	lIncorrect map[int]float64
	rIncorrect map[int]float64
	cCorrect   map[int]float64
	count      float64
	Accuracy   float64 // accuracy goal
	Confidence float32 // threshold for binary classification
	// if not specified it's multi-class classification
}

func (m *Metrics) Copy() model.Metrics {
	return &Metrics{Accuracy: m.Accuracy, Confidence: m.Confidence}
}

func (m *Metrics) Begin() {
	m.correct = 0
	m.lIncorrect = map[int]float64{}
	m.rIncorrect = map[int]float64{}
	m.cCorrect = map[int]float64{}
	m.count = 0
}

/*
Update updates internal false/true|positive/negative counters

label - always is a class number [0..)

result - can be a single integer value in interval [0..) or tensor of float values
	if a single value, it's the class
	otherwise class is selected by hot_one function

*/
func (m *Metrics) Update(result, label reflect.Value) {
	l := fu.Cell{label}.Int()
	y := 0
	if result.Type() == fu.TensorType {
		v := result.Interface().(fu.Tensor)
		y = v.HotOne()
	} else {
		if m.Confidence > 0 {
			x := fu.Cell{result}.Real()
			if x > m.Confidence {
				y = 1
			}
		} else {
			y = fu.Cell{result}.Int()
		}
	}
	if l == y {
		m.correct++
		m.cCorrect[y] = m.cCorrect[y] + 1
	} else {
		m.lIncorrect[l] = m.lIncorrect[l] + 1
		m.rIncorrect[y] = m.rIncorrect[y] + 1
	}
	m.count++
}

func (m *Metrics) Complete() (fu.Struct, bool) {
	if m.count > 0 {
		acc := m.correct / m.count
		cno := float64(len(m.cCorrect))
		var sensitivity, precision float64
		for i, v := range m.cCorrect {
			sensitivity += v / (v + m.lIncorrect[i]) // false negative
			precision += v / (v + m.rIncorrect[i])   // false positive
		}
		sensitivity /= cno
		precision /= cno
		f1 := 2 * precision * sensitivity / (precision + sensitivity)
		columns := []reflect.Value{
			reflect.ValueOf(fu.Round32(float32(acc), 4)),
			reflect.ValueOf(fu.Round32(float32(sensitivity), 4)),
			reflect.ValueOf(fu.Round32(float32(precision), 4)),
			reflect.ValueOf(fu.Round32(float32(f1), 4)),
			reflect.ValueOf(int(m.correct)),
			reflect.ValueOf(int(m.count)),
		}
		goal := false
		if m.Accuracy > 0 {
			goal = goal || acc > m.Accuracy
		}
		return fu.Struct{Names: Names, Columns: columns}, goal
	}
	return fu.NaStruct(Names, fu.Float32), false
}
