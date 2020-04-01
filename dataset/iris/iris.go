package iris

import (
	"github.com/sudachen/go-iokit/iokit"
	"github.com/sudachen/go-ml/lazy"
	"github.com/sudachen/go-ml/tables"
	"github.com/sudachen/go-ml/tables/csv"
)

func source(x string) iokit.Input {
	const base = "https://datahub.io/machine-learning/iris/r/"
	return iokit.Url(base+x, iokit.Cache("go-ml/dataset/iris/"+x))
}

var dataset = source("iris.csv")

var Data tables.Lazy = func() lazy.Stream {
	var cls = tables.Enumset{}
	return csv.Source(dataset,
		csv.Float32("sepallength").As("Feature1"),
		csv.Float32("sepalwidth").As("Feature2"),
		csv.Float32("petallength").As("Feature3"),
		csv.Float32("petalwidth").As("Feature4"),
		csv.Meta(cls.Integer(), "class").As("Label"))()
}
