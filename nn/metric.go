package nn

type Metric interface {
	Reset()
	Collect(data, label []float32)
	Value() float32
	Satisfy() bool
}


