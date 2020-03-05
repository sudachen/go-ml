package ng

import (
	"fmt"
	"github.com/sudachen/go-foo/fu"
	"github.com/sudachen/go-ml/nn"
	"time"
)

func Measure(net *nn.Network, batchs interface{}, metric nn.Metric, verbosity Verbosity) (ok bool, err error) {

	var (
		li, ti Batchs
		count  int
	)

	switch s := batchs.(type) {
	case Batchs:
		ti = s
	case Dataset:
		if li, ti, err = s.Open(net.Input.Len(0)); err != nil {
			return
		}
		ti.Randomize(int(time.Now().Unix()))
		defer func() { _ = li.Close(); _ = ti.Close() }()
	default:
		return false, fmt.Errorf("samples for Measure function must be ng.Batchs or ng.Dataset")
	}

	if err = ti.Reset(); err != nil {
		return
	}

	metric.Reset()
	for ti.Next() {
		net.Test(ti.Data(), ti.Label(), metric)
		count++
	}

	w := net.Input.Len(0)

	verbose(fmt.Sprintf("Accuracy over %d*%d batchs: %v", count, w, fu.Round32(metric.Value(), 3)), verbosity)
	return metric.Satisfy(), nil
}
