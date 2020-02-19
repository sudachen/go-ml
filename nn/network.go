package nn

import (
	"fmt"
	"github.com/sudachen/go-ml/logger"
	"github.com/sudachen/go-ml/nn/mx"
)

type Network struct {
	*mx.Graph

	BatchSize int
}

func (n *Network) Release() {
	n.Graph.Release()
}

func Bind(ctx mx.Context, nb Block, input mx.Dimension, loss mx.Loss) *Network {
	mx.ResetSymbolId(0)
	sym := nb.Combine(mx.Input())
	g := mx.Compose(ctx, sym, loss, input, mx.Float32)
	f := &Network{Graph: g, BatchSize: input.Shape[0]}
	return f
}

func (f *Network) Predict1(data interface{}, out []float32) {
	f.Graph.Input.SetValues(data)
	f.Graph.Forward(false)
	f.Graph.Output.CopyValuesTo(out)
}

func (f *Network) Predict(data interface{}) [][]float32 {
	out := make([]float32, f.Graph.Output.Dim().Total())
	f.Predict1(data, out)
	r := make([][]float32, f.BatchSize)
	stride := len(out) / f.BatchSize
	for i := 0; i < f.BatchSize; i++ {
		r[i] = out[i*stride : (i+1)*stride]
	}
	return r
}

func (f *Network) Test(data, label []float32, metric Metric) {
	out := make([]float32, f.Graph.Output.Dim().Total())
	f.Predict1(data, out)
	count := f.Graph.Output.Len(0)
	outw := len(out) / count
	labelw := len(label) / count
	for i := 0; i < count; i++ {
		metric.Collect(out[outw*i:outw*(i+1)], label[labelw*i:labelw*(i+1)])
	}
}

func (f *Network) Train(data interface{}, label interface{}, opt Optimizer) {
	f.Graph.Input.SetValues(data)
	if f.Graph.Label != nil {
		f.Graph.Label.SetValues(label)
	}
	f.Graph.Forward(true)
	f.Graph.Backward()
	f.Update(opt)
}

func (f *Network) Update(opt Optimizer) {
	for k, g := range f.Graph.Grads {
		opt.Update(f.Graph.Params[k], g)
	}
}

func (f *Network) LoadParamsFile(filename string, force bool) error {
	p, err := LoadParams(filename)
	if err != nil {
		return err
	}
	return f.SetParams(p, force)
}

func (f *Network) SaveParamsFile(filename string) error {
	p := f.GetParams()
	return p.Save(filename)
}

func (f *Network) SetParams(p Params, force bool) error {
	if err := f.checkParams(p, force); err != nil {
		return err
	}
	f.Graph.Initialize(func(d *mx.NDArray, n string){
		a, ok := p.P[n]
		if ok {
			d.SetValues(a[5:])
		} else {
			f.Graph.InitParam(n)
		}
	})
	return nil
}

func (f *Network) checkParams(p Params, force bool) error {
	for n, d := range f.Params {
		a, ok := p.P[n]
		if ok {
			dm := d.Dim()
			if dm.Total() == len(a)-5 {
				x := mx.Dimension{Len: int(a[0]), Shape: [4]int{int(a[1]), int(a[2]), int(a[3]), int(a[4])}}
				if dm != x {
					msg := fmt.Sprintf("parameter %v has dim %v but network requires %v",
						n, x, dm)
					if !force {
						return fmt.Errorf("%v",msg)
					} else {
						logger.Warning(msg)
					}
				}
			} else {
				return fmt.Errorf("parameter %v has %d values but network requires %d",
					n, len(a)-5, dm.Total())
			}
		} else if n[0] != '_' {
			msg := fmt.Sprintf("nonexistent parameter %v is required by network", n)
			if !force {
				return fmt.Errorf("%v", msg)
			} else {
				logger.Warning(msg)
			}
		} else {

		}
	}
	return nil
}

func (net *Network) GetParams() Params {
	p := Params{map[string][]float32{}}
	for n, d := range net.Params {
		if n[0] != '_' {
			dm := d.Dim()
			a := make([]float32, dm.Total()+5)
			a[0] = float32(dm.Len)
			for i := 0; i < 4; i++ {
				a[i+1] = float32(dm.Shape[i])
			}
			d.ReCopyValuesTo(a[5:])
			p.P[n] = a
		}
	}
	return p
}
