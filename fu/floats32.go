package fu

func Mean(a []float32) float32 {
	var c float64
	for _, x := range a {
		c += float64(x)
	}
	return float32(c / float64(len(a)))
}

func Mse(a, b []float32) float32 {
	var c float64
	for i, x := range a {
		q := float64(x - b[i])
		c += q * q
	}
	return float32(c / float64(len(a)))
}
