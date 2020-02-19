package ng

import (
	"github.com/sudachen/go-ml/logger"
	"github.com/sudachen/go-fp/fu"
	"github.com/sudachen/go-ml/nn"
	"path/filepath"
)

type NfoState struct {
	Epoch     int
	EndEpoch  int
	Cicles    int // if EndEpoch is zero => EndEpoch = (last)Epoch + 1 + Cicles
	Metric    DetailedMetric
	AvgLoss   AvgLoss
	Optimizer nn.OptimizerConf
	Gnfo      GymInfo
	GnfoDir   string
	OnEpoch   func(epoch int, net *nn.Network)error
}

func (st *NfoState) Setup(net *nn.Network, iniSeed int) (seed int, err error) {
	logger.Infof("setup nfo state for network %v", net.Identity().String()[:12])
	dir := filepath.Join(st.GnfoDir,net.Identity().String()[:12])
	if st.Gnfo.Exists(dir) {
		if err = st.Gnfo.Load(dir); err != nil {
			return
		}
		st.Epoch = st.Gnfo.Epoch + 1
		if st.EndEpoch == 0 {
			st.EndEpoch = st.Epoch + fu.Fnzi(st.Cicles,1)
		}

		if err = st.Gnfo.LoadParams(st.Epoch-1, net, dir); err != nil {
			return
		}
		return st.Gnfo.Seed, nil
	}
	seed = fu.Fnzi(iniSeed,42)
	if st.EndEpoch == 0 {
		st.EndEpoch = st.Epoch + fu.Fnzi(st.Cicles,1)
	}
	st.Gnfo.Init(net, seed)
	net.Ctx.RandomSeed(seed)
	net.Initialize(nil)
	return seed, nil
}

func (st *NfoState) Preset(net *nn.Network) (opt nn.Optimizer, err error) {
	if st.OnEpoch != nil {
		if err = st.OnEpoch(st.Epoch,net); err != nil {
			return
		}
	}
	st.Optimizer.Init(st.Epoch)
}

func (st *NfoState) LogBatchLoss(loss float32) (err error) {
	st.AvgLoss.Add(loss)
	return
}

func (st *NfoState) NextEpoch(maxEpochs int) (int, error) {
	st.AvgLoss.Reset()
	if st.Epoch < st.EndEpoch {
		return st.Epoch, nil
	}
	return StopTraining, nil
}

func (st *NfoState) FinishEpoch(net *nn.Network, test Batchs) (acc float32, ok bool, err error) {
	dir := filepath.Join(st.GnfoDir,net.Identity().String()[:12])
	st.Metric.Reset()
	if ok, err = Measure(net, test, &st.Metric, Silent); err != nil {
		return
	}
	acc = st.Metric.Value()
	logger.Infof("Epoch(%d) loss:%.4f/%.4f metric:%.3f/%v",
		st.Epoch, st.AvgLoss.Avg, st.AvgLoss.Last(),
		acc,
		st.Metric.Details(),
	)
	st.Gnfo.Update(st.Epoch, acc, st.Metric.Details(), st.AvgLoss)
	if !st.Gnfo.Exists(dir) {
		_ = st.Gnfo.WriteNetwork(net, dir)
	}
	if err = st.Gnfo.SaveParams(st.Epoch, net, dir); err != nil {
		return
	}
	if err = st.Gnfo.Save(dir); err != nil {
		return
	}
	st.Epoch++
	return acc, ok, nil
}
