package iris

import (
	"github.com/sudachen/go-foo/fu"
	"github.com/sudachen/go-foo/lazy"
	"github.com/sudachen/go-ml/tables"
	"github.com/sudachen/go-ml/tables/csv"
)

var dataset = fu.External("https://datahub.io/machine-learning/iris/r/iris.csv",
	fu.Cached("go-ml/dataset/iris/iris.csv"))

var Data tables.Lazy = func() lazy.Stream {
	var cls = tables.Enumset{}
	return csv.Source(dataset,
		csv.Float32("sepallength").As("Feature1"),
		csv.Float32("sepalwidth").As("Feature2"),
		csv.Float32("petallength").As("Feature3"),
		csv.Float32("petalwidth").As("Feature4"),
		csv.Meta(cls.Integer(), "class").As("Label"))()
}
