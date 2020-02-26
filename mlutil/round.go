package mlutil

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
	q := math.Pow(10, float64(digits))
	r := make([]float32, len(a))
	for i, v := range a {
		r[i] = float32(math.Round(float64(v)*q) / q)
	}
	return r
}

func Floor32x(a []float32, digits int) []float32 {
	q := math.Pow(10, float64(digits))
	r := make([]float32, len(a))
	for i, v := range a {
		r[i] = float32(math.Floor(float64(v)*q) / q)
	}
	return r
}

func Round(a float64, digits int) float64 {
	q := math.Pow(10, float64(digits))
	return math.Round(a*q) / q
}

func Floor(a float64, digits int) float64 {
	q := math.Pow(10, float64(digits))
	return math.Floor(a*q) / q
}

func Roundx(a []float64, digits int) []float64 {
	q := math.Pow(10, float64(digits))
	r := make([]float64, len(a))
	for i, v := range a {
		r[i] = math.Round(v*q) / q
	}
	return r
}

func Floorx(a []float64, digits int) []float64 {
	q := math.Pow(10, float64(digits))
	r := make([]float64, len(a))
	for i, v := range a {
		r[i] = math.Floor(float64(v)*q) / q
	}
	return r
}
