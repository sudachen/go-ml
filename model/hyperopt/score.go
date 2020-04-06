package hyperopt

import (
	"github.com/sudachen/go-ml/fu"
)

func TrailScore(score func(fu.Struct)float64) MetricsScore{
	return func(trail TrailMetrics, direction Direction) float64 {
		var test, train float64
		for _,v := range trail {
			test += score(v.Test)
			train += score(v.Train)
		}
		test /= float64(len(trail))
		train /= float64(len(trail))
		if direction == MaximizeScore { return fu.Mind(test,train) }
		return fu.Maxd(test,train)
	}
}
