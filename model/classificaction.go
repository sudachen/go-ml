package model

import (
	"github.com/sudachen/go-ml/fu"
	"reflect"
)

/*
Classification metrics factory
*/
type Classification struct {
	Accuracy   float64 // accuracy goal
	Error      float64 // error goal
	Confidence float32 // threshold for binary classification
}

/*
Names is the list of calculating metrics
*/
func (m Classification) Names() []string {
	return []string{
		IterationCol,
		SubsetCol,
		ErrorCol,
		LossCol,
		AccuracyCol,
		SensitivityCol,
		PrecisionCol,
		F1ScoreCol,
		CorrectCol,
		TotalCol,
	}
}

/*
New metrics updater for the given iteration and subset
*/
func (m Classification) New(iteration int, subset string) MetricsUpdater {
	return &cfupdater{
		Classification: m,
		iteration:      iteration,
		subset:         subset,
		lIncorrect:     map[int]float64{},
		rIncorrect:     map[int]float64{},
		cCorrect:       map[int]float64{},
	}
}

type cfupdater struct {
	Classification
	iteration  int
	subset     string
	correct    float64
	loss       float64
	lIncorrect map[int]float64
	rIncorrect map[int]float64
	cCorrect   map[int]float64
	count      float64
}

func (m *cfupdater) Update(result, label reflect.Value, loss float64) {
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
	m.loss += loss
	m.count++
}

func (m *cfupdater) Complete() (fu.Struct, bool) {
	if m.count > 0 {
		acc := m.correct / m.count
		cno := float64(len(m.cCorrect))
		var sensitivity, precision, cerr float64
		for i, v := range m.cCorrect {
			sensitivity += v / (v + m.lIncorrect[i]) // false negative
			precision += v / (v + m.rIncorrect[i])   // false positive
			cerr += (m.rIncorrect[i] + m.lIncorrect[i]) / m.count
		}
		sensitivity /= cno
		precision /= cno
		cerr /= cno
		f1 := 2 * precision * sensitivity / (precision + sensitivity)
		columns := []reflect.Value{
			reflect.ValueOf(m.iteration),
			reflect.ValueOf(m.subset),
			reflect.ValueOf(cerr),
			reflect.ValueOf(m.loss / m.count),
			reflect.ValueOf(acc),
			reflect.ValueOf(sensitivity),
			reflect.ValueOf(precision),
			reflect.ValueOf(f1),
			reflect.ValueOf(int(m.correct)),
			reflect.ValueOf(int(m.count)),
		}
		goal := false
		if m.Accuracy > 0 {
			goal = goal || acc > m.Accuracy
		}
		if m.Error > 0 {
			goal = goal || cerr < m.Error
		}
		return fu.Struct{Names: m.Names(), Columns: columns}, goal
	}
	return fu.
			NaStruct(m.Names(), fu.Float64).
			Set(IterationCol, fu.IntZero).
			Set(SubsetCol, fu.EmptyString),
		false
}
