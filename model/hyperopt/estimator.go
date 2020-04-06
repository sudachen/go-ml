package hyperopt

import (
	"gonum.org/v1/gonum/floats"
	"math"
)

type estimator struct {
	weights []float64
	mus     []float64
	sigmas  []float64
}

func buildEstimator( mus []float64, low float64, high float64, s *sampler ) estimator {
	considerPrior := true
	considerMagicClip := false
	considerEndpoints := true
	priorWeight := s.priorWeight

	var sortedWeights []float64
	var sortedMus []float64
	var sigma []float64

	var order []int
	var priorPos int
	var priorSigma float64
	if considerPrior {
		priorMu := 0.5 * (low + high)
		priorSigma = 1.0 * (high - low)
		if len(mus) == 0 {
			sortedMus = []float64{priorMu}
			sigma = []float64{priorSigma}
			priorPos = 0
			order = make([]int, 0)
		} else {
			order = make([]int, len(mus))
			floats.Argsort(mus, order)
			priorPos = Location(Choice(mus, order), priorMu)
			sortedMus = make([]float64, 0, len(mus)+1)
			sortedMus = append(sortedMus, Choice(mus, order[:priorPos])...)
			sortedMus = append(sortedMus, priorMu)
			sortedMus = append(sortedMus, Choice(mus, order[priorPos:])...)
		}
	} else {
		order = make([]int, len(mus))
		floats.Argsort(mus, order)
		sortedMus = Choice(mus, order)
	}

	// we decide the sigma.
	if len(mus) > 0 {
		lowSortedMusHigh := append(sortedMus, high)
		lowSortedMusHigh = append([]float64{low}, lowSortedMusHigh...)

		l := len(lowSortedMusHigh)
		sigma = make([]float64, l)
		for i := 0; i < l-2; i++ {
			sigma[i+1] = math.Max(lowSortedMusHigh[i+1]-lowSortedMusHigh[i], lowSortedMusHigh[i+2]-lowSortedMusHigh[i+1])
		}
		if !considerEndpoints && len(lowSortedMusHigh) > 2 {
			sigma[1] = lowSortedMusHigh[2] - lowSortedMusHigh[1]
			sigma[l-2] = lowSortedMusHigh[l-2] - lowSortedMusHigh[l-3]
		}
		sigma = sigma[1 : l-1]
	}

	// we decide the weights.
	unsortedWeights := weights(len(mus))
	if considerPrior {
		sortedWeights = make([]float64, 0, len(sortedMus))
		sortedWeights = append(sortedWeights, Choice(unsortedWeights, order[:priorPos])...)
		sortedWeights = append(sortedWeights, priorWeight)
		sortedWeights = append(sortedWeights, Choice(unsortedWeights, order[priorPos:])...)
	} else {
		sortedWeights = Choice(unsortedWeights, order)
	}
	sumSortedWeights := floats.Sum(sortedWeights)
	for i := range sortedWeights {
		sortedWeights[i] /= sumSortedWeights
	}

	// We adjust the range of the 'sigma' according to the 'consider_magic_clip' flag.
	maxSigma := 1.0 * (high - low)
	var minSigma float64
	if considerMagicClip {
		minSigma = 1.0 * (high - low) / math.Min(100.0, 1.0+float64(len(sortedMus)))
	} else {
		minSigma = epsilon
	}
	Clip(sigma, minSigma, maxSigma)
	if considerPrior {
		sigma[priorPos] = priorSigma
	}

	return estimator{sortedWeights, sortedMus, sigma }
}
