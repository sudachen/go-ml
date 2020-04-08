package hyperopt

import (
	"github.com/sudachen/go-ml/fu"
	"github.com/sudachen/go-zorros/zorros"
	"gonum.org/v1/gonum/floats"
	"math"
	"math/rand"
	"sort"
)

type sampler struct {
	numberOfStartupTrials int
	numberOfEICandidates  int

	rng *rand.Rand

	priorWeight float64
}

func (s *sampler) sample(name string, dist distribution, opt *optimizer) (value float64) {
	values, scores := opt.observationPairs(name)
	n := len(values)

	if n < s.numberOfStartupTrials {
		value = dist.sample1(s)
	} else {
		belowParamValues, aboveParamValues := s.splitObservationPairs(values, scores)
		value = dist.sample2(s, belowParamValues, aboveParamValues)
	}

	opt.update(name, value)
	return
}

func weights(x int) []float64 {
	if x == 0 {
		return []float64{}
	} else if x < 25 {
		return Ones1d(x)
	} else {
		ramp := Linspace(1.0/float64(x), 1.0, x-25, true)
		flat := Ones1d(25)
		return append(ramp, flat...)
	}
}

// multinomial draw samples from a multinomial distribution like numpy.random.multinomial.
// See https://docs.scipy.org/doc/numpy-1.15.0/reference/generated/numpy.random.multinomial.html
func multinomial(n int, pvals []float64, size int) [][]int {
	result := make([][]int, size)
	l := len(pvals)
	x := make([]float64, l)
	floats.CumSum(x, pvals)

	for i := range result {
		result[i] = make([]int, l)

		for j := 0; j < n; j++ {

			var index int
			r := rand.Float64()
			for i := range x {
				if x[i] > r {
					index = i
					break
				}
			}
			result[i][index]++
		}
	}
	return result
}

func argmaxMultinomial(pvals []float64) (int, error) {
	x := make([]float64, len(pvals))
	floats.CumSum(x, pvals)

	r := rand.Float64()
	for i := range x {
		if x[i] > r {
			return i, nil
		}
	}
	return 0, zorros.Errorf("invalid pvals")
}

func gamma(x int) int {
	a := int(math.Ceil(0.1 * float64(x)))
	if a > 25 {
		return 25
	}
	return a
}

func (s *sampler) splitObservationPairs(configVals []float64, lossVals [][2]float64) ([]float64, []float64) {
	nbelow := gamma(len(configVals))
	lossAscending := ArgSort2d(lossVals)

	sort.Ints(lossAscending[:nbelow])
	below := Choice(configVals, lossAscending[:nbelow])

	sort.Ints(lossAscending[nbelow:])
	above := Choice(configVals, lossAscending[nbelow:])
	return below, above
}

func (s *sampler) sampleFromCategoricalDist(probabilities []float64, size int) []int {
	if size == 0 {
		return []int{}
	}
	sample := multinomial(1, probabilities, size)

	returnVals := make([]int, size)
	for i := 0; i < size; i++ {
		for j := range sample[i] {
			returnVals[i] += sample[i][j] * j
		}
	}
	return returnVals
}

func (s *sampler) categoricalLogPDF(sample []int, p []float64) []float64 {
	if len(sample) == 0 {
		return []float64{}
	}

	result := make([]float64, len(sample))
	for i := 0; i < len(sample); i++ {
		result[i] = math.Log(p[sample[i]])
	}
	return result
}

func (s *sampler) compare(samples []float64, logL []float64, logG []float64) []float64 {
	if len(samples) == 0 {
		return []float64{}
	}
	if len(logL) != len(logG) {
		panic("the size of the log_l and log_g should be same")
	}
	score := make([]float64, len(logL))
	for i := range score {
		score[i] = logL[i] - logG[i]
	}
	if len(samples) != len(score) {
		panic("the size of the samples and score should be same")
	}

	argMax := func(s []float64) int {
		max := s[0]
		maxIdx := 0
		for i := range s {
			if i == 0 {
				continue
			}
			if s[i] > max {
				max = s[i]
				maxIdx = i
			}
		}
		return maxIdx
	}
	best := argMax(score)
	results := make([]float64, len(samples))
	for i := range results {
		results[i] = samples[best]
	}
	return results
}

func (s *sampler) gmmLogPDF(samples []float64, pe estimator, low, high float64, q float64, isLog bool) []float64 {

	if len(samples) == 0 {
		return []float64{}
	}

	highNormalCdf := s.normalCDF(high, pe.mus, pe.sigmas)
	lowNormalCdf := s.normalCDF(low, pe.mus, pe.sigmas)
	if len(pe.weights) != len(highNormalCdf) {
		panic("the length should be the same with weights")
	}

	paccept := 0.0
	for i := 0; i < len(highNormalCdf); i++ {
		paccept += highNormalCdf[i]*pe.weights[i] - lowNormalCdf[i]
	}

	if q > 0 {
		probabilities := make([]float64, len(samples))
		for i := range pe.weights {
			w := pe.weights[i]
			mu := pe.mus[i]
			sigma := pe.sigmas[i]
			upperBound := make([]float64, len(samples))
			lowerBound := make([]float64, len(samples))
			for i := range upperBound {
				if isLog {
					upperBound[i] = math.Min(samples[i]+q/2.0, math.Exp(high))
					lowerBound[i] = math.Max(samples[i]-q/2.0, math.Exp(low))
					lowerBound[i] = math.Max(0, lowerBound[i])
				} else {
					upperBound[i] = math.Min(samples[i]+q/2.0, high)
					lowerBound[i] = math.Max(samples[i]-q/2.0, low)
				}
			}

			incAmt := make([]float64, len(samples))
			for j := range upperBound {
				if isLog {
					incAmt[j] = w * s.logNormalCDF(upperBound[j], []float64{mu}, []float64{sigma})[0]
					incAmt[j] -= w * s.logNormalCDF(lowerBound[j], []float64{mu}, []float64{sigma})[0]
				} else {
					incAmt[j] = w * s.normalCDF(upperBound[j], []float64{mu}, []float64{sigma})[0]
					incAmt[j] -= w * s.normalCDF(lowerBound[j], []float64{mu}, []float64{sigma})[0]
				}
			}
			for j := range probabilities {
				probabilities[j] += incAmt[j]
			}
		}
		returnValue := make([]float64, len(samples))
		for i := range probabilities {
			returnValue[i] = math.Log(probabilities[i]+epsilon) + math.Log(paccept+epsilon)
		}
		return returnValue
	}

	var (
		jacobian []float64
		distance [][]float64
	)
	if isLog {
		jacobian = samples
	} else {
		jacobian = Ones1d(len(samples))
	}
	distance = make([][]float64, len(samples))
	for i := range samples {
		distance[i] = make([]float64, len(pe.mus))
		for j := range pe.mus {
			if isLog {
				distance[i][j] = math.Log(samples[i]) - pe.mus[j]
			} else {
				distance[i][j] = samples[i] - pe.mus[j]
			}
		}
	}
	mahalanobis := make([][]float64, len(distance))
	for i := range distance {
		mahalanobis[i] = make([]float64, len(distance[i]))
		for j := range distance[i] {
			mahalanobis[i][j] = distance[i][j] / math.Pow(math.Max(pe.sigmas[j], epsilon), 2)
		}
	}
	z := make([][]float64, len(distance))
	for i := range distance {
		z[i] = make([]float64, len(distance[i]))
		for j := range distance[i] {
			z[i][j] = math.Sqrt(2*math.Pi) * pe.sigmas[j] * jacobian[i]
		}
	}
	coefficient := make([][]float64, len(distance))
	for i := range distance {
		coefficient[i] = make([]float64, len(distance[i]))
		for j := range distance[i] {
			coefficient[i][j] = pe.weights[j] / z[i][j] / paccept
		}
	}

	y := make([][]float64, len(distance))
	for i := range distance {
		y[i] = make([]float64, len(distance[i]))
		for j := range distance[i] {
			y[i][j] = -0.5*mahalanobis[i][j] + math.Log(coefficient[i][j])
		}
	}
	return s.logsumRows(y)
}

func (s *sampler) normalCDF(x float64, mu []float64, sigma []float64) []float64 {
	l := len(mu)
	results := make([]float64, l)
	for i := 0; i < l; i++ {
		denominator := x - mu[i]
		numerator := math.Max(math.Sqrt(2)*sigma[i], epsilon)
		z := denominator / numerator
		results[i] = 0.5 * (1 + math.Erf(z))
	}
	return results
}

func (s *sampler) logNormalCDF(x float64, mu []float64, sigma []float64) []float64 {
	if x < 0 {
		panic("negative argument is given to logNormalCDF")
	}
	l := len(mu)
	results := make([]float64, l)
	for i := 0; i < l; i++ {
		denominator := math.Log(math.Max(x, epsilon)) - mu[i]
		numerator := math.Max(math.Sqrt(2)*sigma[i], epsilon)
		z := denominator / numerator
		results[i] = 0.5 + (0.5 * math.Erf(z))
	}
	return results
}

func (s *sampler) logsumRows(x [][]float64) []float64 {
	y := make([]float64, len(x))
	for i := range x {
		m := floats.Max(x[i])

		sum := 0.0
		for j := range x[i] {
			sum += math.Log(math.Exp(x[i][j] - m))
		}
		y[i] = sum + m
	}
	return y
}

func (s *sampler) sampleNumerical(low, high float64, below, above []float64, step float64, islog bool) float64 {

	if islog {
		low = math.Log(low)
		high = math.Log(high)
		for i := range below {
			below[i] = math.Log(below[i])
		}
		for i := range above {
			above[i] = math.Log(above[i])
		}
	}

	size := s.numberOfEICandidates
	estimatorBelow := buildEstimator(below, low, high, s)
	sampleBelow := s.sampleFromGMM(estimatorBelow, low, high, size, step, islog)
	logLikelihoodsBelow := s.gmmLogPDF(sampleBelow, estimatorBelow, low, high, step, islog)

	estimatorAbove := buildEstimator(above, low, high, s)
	logLikelihoodsAbove := s.gmmLogPDF(sampleBelow, estimatorAbove, low, high, step, islog)

	return s.compare(sampleBelow, logLikelihoodsBelow, logLikelihoodsAbove)[0]
}

func (s *sampler) sampleFromGMM(pe estimator, low, high float64, size int, q float64, isLog bool) []float64 {
	nsamples := size

	if low > high {
		panic("the low should be lower than the high")
	}

	samples := make([]float64, 0, nsamples)
	for {
		if len(samples) == nsamples {
			break
		}
		active, err := argmaxMultinomial(pe.weights)
		if err != nil {
			panic(err)
		}
		x := s.rng.NormFloat64()
		draw := x*pe.sigmas[active] + pe.mus[active]
		if low <= draw && draw < high {
			samples = append(samples, draw)
		}
	}

	if isLog {
		for i := range samples {
			samples[i] = math.Exp(samples[i])
		}
	}

	if q > 0 {
		for i := range samples {
			samples[i] = math.Round(samples[i]/q) * q
		}
	}
	return samples
}

func (r Range) sample1(s *sampler) (x float64) {
	if r[0] < 0 {
		zorros.Panic(zorros.Errorf("negative hyper-parameters are not allowed"))
	}
	if r[0]+epsilon >= r[1] {
		zorros.Panic(zorros.Errorf("empty hyper-parameters range"))
	}
	x = s.rng.Float64()*(r[1]-r[0]) + r[0]
	if x < epsilon {
		x = epsilon
	}
	return
}

func (r Range) sample2(s *sampler, below, above []float64) (x float64) {
	x = s.sampleNumerical(r[0], r[1], below, above, 0, false)
	if x < epsilon {
		x = epsilon
	}
	return
}

func (r IntRange) sample1(s *sampler) float64 {
	if r[0] <= 0 {
		zorros.Panic(zorros.Errorf("zero and negative hyper-parameters are not allowed"))
	}
	if r[0] >= r[1]+1 {
		zorros.Panic(zorros.Errorf("empty hyper-parameters range"))
	}
	if r[0] == r[1] {
		return float64(r[0])
	}
	return float64(s.rng.Intn(r[1]+1-r[0]) + r[0])
}

func (r IntRange) sample2(s *sampler, below, above []float64) (x float64) {
	x = math.Ceil(s.sampleNumerical(float64(r[0])-0.5, float64(r[1])+0.5, below, above, 1, false))
	if x < epsilon {
		x = epsilon
	}
	return
}

func (r LogIntRange) sample1(s *sampler) float64 {
	if r[0] <= 0 {
		zorros.Panic(zorros.Errorf("zero and negative hyper-parameters are not allowed"))
	}
	if r[0] >= r[1]+1 {
		zorros.Panic(zorros.Errorf("empty hyper-parameters range"))
	}
	if r[0] == r[1] {
		return float64(r[0])
	}
	logLow := math.Log(float64(r[0]))
	logHigh := math.Log(float64(r[1] + 1))
	return math.Ceil(math.Exp(s.rng.Float64()*(logHigh-logLow) + logLow))
}

func (r LogIntRange) sample2(s *sampler, below, above []float64) (x float64) {
	x = math.Ceil(s.sampleNumerical(float64(r[0])-0.5, float64(r[1])+0.5, below, above, 1, true))
	if x < epsilon {
		x = epsilon
	}
	return
}

func (r LogRange) sample1(s *sampler) (x float64) {
	if r[0] < 0 {
		zorros.Panic(zorros.Errorf("negative hyper-parameters are not allowed"))
	}
	if r[0]+epsilon >= r[1] {
		zorros.Panic(zorros.Errorf("empty hyper-parameters range"))
	}
	logLow := math.Log(float64(fu.Ifed(r[0] < epsilon, epsilon, r[0])))
	logHigh := math.Log(float64(r[1]))
	x = math.Exp(s.rng.Float64()*(logHigh-logLow) + logLow)
	if math.IsNaN(x) {
		panic("")
	}
	return
}

func (r LogRange) sample2(s *sampler, below, above []float64) (x float64) {
	x = s.sampleNumerical(r[0], r[1], below, above, 0, true)
	if x < epsilon {
		x = epsilon
	}
	return
}

func (v Value) sample1(_ *sampler) float64 {
	return float64(v)
}

func (v Value) sample2(_ *sampler, _, _ []float64) float64 {
	return float64(v)
}

func (l List) sample1(s *sampler) float64 {
	if len(l) == 0 {
		zorros.Panic(zorros.Errorf("empty hyper-parameters list"))
	}
	if len(l) == 1 {
		return l[0]
	}
	return float64(l[s.rng.Intn(len(l))])
}

func (l List) sample2(s *sampler, below, above []float64) float64 {
	belowInt := make([]int, len(below))
	for i := range below {
		belowInt[i] = int(below[i])
	}
	aboveInt := make([]int, len(above))
	for i := range above {
		aboveInt[i] = int(above[i])
	}
	upper := len(l)
	size := s.numberOfEICandidates

	// below
	weightsBelow := weights(len(below))
	countsBelow := Bincount(belowInt, weightsBelow, upper)
	weightedBelowSum := 0.0
	weightedBelow := make([]float64, len(countsBelow))
	for i := range countsBelow {
		weightedBelow[i] = countsBelow[i] + s.priorWeight
		weightedBelowSum += weightedBelow[i]
	}
	for i := range weightedBelow {
		weightedBelow[i] /= weightedBelowSum
	}
	samplesBelow := s.sampleFromCategoricalDist(weightedBelow, size)
	logLikelihoodsBelow := s.categoricalLogPDF(samplesBelow, weightedBelow)

	// above
	weightsAbove := weights(len(above))
	countsAbove := Bincount(aboveInt, weightsAbove, upper)
	weightedAboveSum := 0.0
	weightedAbove := make([]float64, len(countsAbove))
	for i := range countsAbove {
		weightedAbove[i] = countsAbove[i] + s.priorWeight
		weightedAboveSum += weightedAbove[i]
	}
	for i := range weightedAbove {
		weightedAbove[i] /= weightedAboveSum
	}
	samplesAbove := s.sampleFromCategoricalDist(weightedAbove, size)
	logLikelihoodsAbove := s.categoricalLogPDF(samplesAbove, weightedAbove)

	floatSamplesBelow := make([]float64, len(samplesBelow))
	for i := range samplesBelow {
		floatSamplesBelow[i] = float64(samplesBelow[i])
	}
	return s.compare(floatSamplesBelow, logLikelihoodsBelow, logLikelihoodsAbove)[0]
}
