/*
Package hyperopt implements SMBO/TPE hyper-parameter optimization for ML models

Many thanks to Masashi SHIBATA for his excellent work on goptuna
I used github.com/c-bata/goptuna as a reference implementation
for the paper 'Algorithms for Hyper-Parameter Optimization'
https://papers.nips.cc/paper/4443-algorithms-for-hyper-parameter-optimization.pdf

TPE sampler mostly derived from goptuna.
*/
package hyperopt

import (
	"github.com/sudachen/go-ml/fu"
	"github.com/sudachen/go-ml/model"
	"github.com/sudachen/go-ml/tables"
	"github.com/sudachen/go-zorros/zorros"
	"reflect"
)

const epsilon = 1e-12

/*
Range is a open float range specified by min and max values (min,max)
*/
type Range [2]float64

/*
LogRange is a open float logarithmic range specified by min and max values (min,max)
*/
type LogRange [2]float64

/*
IntRange is a close integer range specified by min and max values [min,max]
*/
type IntRange [2]int

/*
LogRange is a close logarithmic integer range specified by min and max values [min,max]
*/
type LogIntRange [2]int

/*
List is a list of possible parameter values
*/
type List []float64

/*
Value is a single value parameter
*/
type Value float64

// type limitation interface
type distribution interface {
	sample1(*sampler) float64
	sample2(*sampler, []float64, []float64) float64
}

/*
Variance is a space of hyper-parameters used in *Search functions
*/
type Variance map[string]distribution

/*
Params is a set of hyper-parameters used by *SearchCV functions to generate new model
*/
type Params map[string]float64

/*
Get value of the parameter by name if exists and dflt value otherwise
*/
func (p Params) Get(name string, dflt float64) float64 {
	if v, ok := p[name]; ok {
		return v
	}
	return dflt
}

/*
BestParams is a result of Hyper-parameters Optimization
*/
type BestParams struct {
	Params
	Score float64
}

/*
Space is a definition of hyper-parameters optimization space
*/
type Space struct {
	Source     tables.AnyData // dataset source
	Features   []string       // dataset features
	Label      string         // dataset label
	Seed       int            // random seed
	Kfold      int            // count of dataset folds
	Iterations int            // model fitting iterations
	Metrics    model.Metrics  // model evaluation metrics
	Score      model.Score    // function to calculate score of train/test metrics

	// the model generation function
	ModelFunc func(Params) model.HungryModel

	// hyper-parameters variance
	Variance Variance
}

/*
Apply apples params to a model
*/
func Apply(m interface{}, p Params) {
	x := reflect.ValueOf(m).Elem()
	for k, v := range p {
		z := x.FieldByName(k)
		if !z.IsValid() {
			panic(zorros.Panic(zorros.Errorf("model does not have field `%v`", k)))
		}
		z.Set(fu.Convert(reflect.ValueOf(v), false, z.Type()))
	}
}
