package hyperopt

import (
	"github.com/sudachen/go-ml/fu"
	"github.com/sudachen/go-ml/fu/verbose"
	"github.com/sudachen/go-ml/model"
	"github.com/sudachen/go-zorros/zorros"
	"math"
	"math/rand"
)

type kfoldMetrics struct {
	Test, Train fu.Struct
	Score       float64
}

type trailMetrics []*kfoldMetrics

type optimizer struct {
	params  []Params
	scores  []float64
	metrics []trailMetrics
}

func (opt *optimizer) observationPairs(name string) ([]float64, [][2]float64) {
	var sign float64 = -1 // maximize score
	L := len(opt.scores)
	if L == 0 {
		return []float64{}, [][2]float64{}
	}

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
		opt.params = append(opt.params, Params{})
	}
	opt.params[L][name] = value
}

func (opt *optimizer) current() Params {
	return opt.params[len(opt.params)-1]
}

func (opt *optimizer) complete(value float64, metrics trailMetrics) {
	opt.scores = append(opt.scores, value)
	opt.metrics = append(opt.metrics, metrics)
}

const kfoldTest = model.TestCol

func (ss Space) Optimize(trails int) (best BestParams, err error) {

	if len(ss.Features) == 0 {
		err = zorros.Errorf("dataset features is not specified")
		return
	}

	Label := fu.Fnzs(ss.Label,model.LabelCol)

	opt := &optimizer{}
	seed := fu.Seed(ss.Seed)
	sm := &sampler{10, 24, rand.New(rand.NewSource(0)), 1}
	for rno := 0; rno < trails; rno++ {

		for k, d := range ss.Variance {
			sm.sample(k, d, opt)
		}

		params := opt.current()
		verbose.Printf("[%3d] sampled params: %#v", rno, params)
		hm := ss.ModelFunc(params)
		var trail = trailMetrics{}

		// k-fold cross-validation
		for k := 0; k < ss.Kfold; k++ {
			var report *model.Report
			report, err = hm.Feed(model.Dataset{
				Source:   ss.Source.Lazy().Kfold(seed, ss.Kfold, k, kfoldTest),
				Label:    Label,
				Features: ss.Features,
				Test:     kfoldTest,
			}).Train(model.Training{
				Iterations:   ss.Iterations,
				Metrics:      ss.Metrics,
				Score:        ss.Score,
			})
			if err != nil {
				return
			}
			t := &kfoldMetrics{report.Test,report.Train,report.Score}
			trail = append(trail, t)
			verbose.Printf("[%3d/%3d] k-fold test: %v", rno, k, t.Test)
			verbose.Printf("[%3d/%3d] k-fold train: %v", rno, k, t.Train)
		}

		score := 0.0
		for _, v := range trail {
			score += v.Score
		}
		score /= float64(len(trail))

		verbose.Printf("[%3d] completed: %v", rno, score)
		// complete with commutative score by k-fold metrics
		opt.complete(score, trail)
	}

	// find params with the best score
	j := fu.IndexOfMax(opt.scores)
	best = BestParams{opt.params[j], opt.scores[j]}
	return
}

func (ss Space) LuckyOptimize(trails int) BestParams {
	p, err := ss.Optimize(trails)
	if err != nil {
		zorros.Panic(err)
	}
	return p
}
