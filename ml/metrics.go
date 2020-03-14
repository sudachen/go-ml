package ml

type Metric interface{

}

type Metrics []Metric
func (Metrics) Fitparam() {}

