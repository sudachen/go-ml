package classification

import "github.com/sudachen/go-ml/fu"

func score(a1, a2 float64) float64 {
	if a1 > a2 {
		a2, a1 = a1, a2
	}
	return (a1 - (a2-a1)/2)
}

func ErrorScore(train, test fu.Struct) float64 {
	return score(1-Error(test), 1-Error(train))
}

func AccuracyScore(train, test fu.Struct) float64 {
	return score(Accuracy(test), Accuracy(train))
}
