package hyperopt

import (
	"github.com/sudachen/go-ml/fu"
	"github.com/sudachen/go-ml/model"
	"github.com/sudachen/go-ml/tables"
	"github.com/sudachen/go-zorros/zorros"
	"math"
	"math/rand"
)

type internalParams map[string]float64
func (p internalParams) params() Params {
	r := Params{}
	for k,v := range p {
		r[k] = float32(v)
	}
	return r
}

type optimizer struct {
	direct Direction
	params []internalParams
	scores []float64
	metrics []TrailMetrics
}

func (opt *optimizer) observationPairs(name string) ([]float64, [][2]float64) {
	var sign float64 = 1
	if opt.direct == MaximizeScore {
		sign = -1
	}
	L := len(opt.scores)
	if L == 0 { return []float64{}, [][2]float64{} }

	values := make([]float64, L)
	scores := make([][2]float64, L)
	for i, p := range opt.params[:L] {
		values[i] = p[name]
		score0 := math.Inf(-1) // TODO: remove this part
		score1 := sign * opt.scores[i]
		scores[i] = [2]float64{score0, score1}
	}
	return values, scores
}

func (opt *optimizer) update(name string, value float64) {
	L := len(opt.scores)
	if L == len(opt.params) {
		opt.params = append(opt.params,internalParams{})
	}
	opt.params[L][name] = value
}

func (opt *optimizer) current() Params {
	return opt.params[len(opt.params)-1].params()
}

func (opt *optimizer) complete(value float64, metrics TrailMetrics) {
	opt.scores = append(opt.scores,value)
	opt.metrics = append(opt.metrics,metrics)
}

const KfoldTest = "Test"

func thelast(t *tables.Table, test bool) fu.Struct {
	c := t.Col(model.MetricsSubset)
	L := c.Len()
	for i := L-1; i >= 0; i-- {
		if (test && c.Text(i) == model.MetricsTestSubset) ||
		   (!test && c.Text(i) == model.MetricsTrainSubset) {
			return t.Index(i)
		}
	}
	panic(zorros.Panic(zorros.Errorf("metrics does not have filed KfoldTest")))
}

func (ss Space) Optimize(trails int) (best BestParams, err error) {

	if len(ss.Features) == 0 {
		err = zorros.Errorf("dataset features is not specified")
		return
	}

	if len(ss.Label) == 0 {
		err = zorros.Errorf("dataset label is not specified")
		return
	}

	opt := &optimizer{direct: ss.Direction}
	seed := fu.Seed(ss.Seed)
	sm := &sampler{10,24,rand.New(rand.NewSource(0)), 1}
	for rno := 0; rno < trails; rno++ {
		for k, d := range ss.Variance {
			sm.sample(k, d, opt)
		}

		hm := ss.ModelFunc(opt.current())
		var trail = TrailMetrics{}

		// k-fold cross-validation
		for k := 0; k < ss.Kfold; k++ {
			var metrics *tables.Table
			metrics, err = hm.Feed(model.Dataset{
				Source: ss.Source.Lazy().Kfold(seed, ss.Kfold, k, KfoldTest),
				Label: ss.Label,
				Features: ss.Features,
				Test: KfoldTest,
			}).Fit(ss.Iterations, nil, ss.Metrics)
			if err != nil { return }
			trail = append(trail, &KfoldMetrics{
					thelast(metrics,true),
					thelast(metrics,false)})
		}

		// complete with commutative score by k-fold metrics
		opt.complete(ss.Score(trail,ss.Direction),trail)
	}

	index := fu.IndexOfMax
	if ss.Direction == MinimizeScore { index = fu.IndexOfMin }
	// find params with the best score
	j := index(opt.scores)
	best = BestParams{ opt.params[j].params(), opt.scores[j] }
	return
}

func (ss Space) LuckyOptimize(trails int) BestParams {
	p, err := ss.Optimize(trails)
	if err != nil { zorros.Panic(err) }
	return p
}
