package util

import "math"

func Round32(a float32, digits int) float32 {
	q := math.Pow(10, float64(digits))
	return float32(math.Round(float64(a)*q) / q)
}

func Floor32(a float32, digits int) float32 {
	q := math.Pow(10, float64(digits))
	return float32(math.Floor(float64(a)*q) / q)
}

func Round32x(a []float32, digits int) []float32 {
	r := make([]float32, len(a))
	for i, v := range a {
		r[i] = float32(Round32(v, digits))
	}
	return r
}

func Floor32x(a []float32, digits int) []float32 {
	r := make([]float32, len(a))
	for i, v := range a {
		r[i] = float32(Floor32(v, digits))
	}
	return r
}
