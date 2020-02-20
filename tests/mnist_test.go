package tests

import (
	"github.com/sudachen/go-ml/nn"
	"github.com/sudachen/go-ml/nn/mx"
	"github.com/sudachen/go-ml/nn/ng"
	"github.com/sudachen/go-ml/util"
	"github.com/sudachen/go-ml/util/mnist"
	"gotest.tools/assert"
	"testing"
	"time"
)

var mnistMLP0 = nn.Connect(
	&nn.FullyConnected{Size: 128, Activation: nn.ReLU},
	&nn.FullyConnected{Size: 64, Activation: nn.Swish, BatchNorm: true },
	&nn.FullyConnected{Size: 10, Activation: nn.Softmax, BatchNorm: true})

func Test_mnistMLP0(t *testing.T) {

	gym := &ng.Gym{
		Optimizer: &nn.Adam{Lr: .001},
		Loss:      &nn.LabelCrossEntropyLoss{},
		Input:     mx.Dim(32, 1, 28, 28),
		Epochs:    10,
		Verbose:   ng.Printing,
		Every:     1 * time.Second,
		Dataset:   &mnist.Dataset{},
		Metric:    &ng.Classification{Accuracy: 0.96},
		Seed:      42,
	}

	acc, params, err := gym.Train(mx.CPU, mnistMLP0)
	assert.NilError(t, err)
	assert.Assert(t, acc >= 0.96)

	net1 := nn.Bind(mx.CPU, mnistMLP0, mx.Dim(50, 1, 28, 28), nil)
	defer net1.Release()
	net1.PrintSummary(false)
	err = net1.SetParams(params, false)
	assert.NilError(t, err)
	ok, err := ng.Measure(net1, &mnist.Dataset{}, &ng.Classification{Accuracy: 0.96}, ng.Printing)
	assert.Assert(t, ok)
	err = net1.SaveParamsFile(util.CacheFile("tests/mnistMLP0.params"))
	assert.NilError(t, err)

	net2 := nn.Bind(mx.CPU, mnistMLP0, mx.Dim(10, 1, 28, 28), nil)
	err = net2.LoadParamsFile(util.CacheFile("tests/mnistMLP0.params"), false)
	assert.NilError(t, err)
	ok, err = ng.Measure(net2, &mnist.Dataset{}, &ng.Classification{Accuracy: 0.96}, ng.Printing)
	assert.Assert(t, ok)
}

var mnistConv0 = nn.Connect(
	&nn.Convolution{Channels: 24, Kernel: mx.Dim(3, 3), Activation: nn.ReLU},
	&nn.MaxPool{Kernel: mx.Dim(2, 2), Stride: mx.Dim(2, 2)},
	&nn.Convolution{Channels: 32, Kernel: mx.Dim(5, 5), Activation: nn.ReLU, BatchNorm: true},
	&nn.MaxPool{Kernel: mx.Dim(2, 2), Stride: mx.Dim(2, 2)},
	&nn.FullyConnected{Size: 32, Activation: nn.Swish, BatchNorm: true, Dropout: 0.33},
	&nn.FullyConnected{Size: 10, Activation: nn.Softmax})

func Test_mnistConv0(t *testing.T) {

	gym := &ng.Gym{
		Optimizer: &nn.Adam{Lr: .001},
		Loss:      &nn.LabelCrossEntropyLoss{},
		Input:     mx.Dim(32, 1, 28, 28),
		Epochs:    5,
		Verbose:   ng.Printing,
		Every:     1 * time.Second,
		Dataset:   &mnist.Dataset{},
		Metric:    &ng.Classification{Accuracy: 0.98},
		Seed:      42,
	}

	acc, params, err := gym.Train(mx.CPU, mnistConv0)
	assert.NilError(t, err)
	assert.Assert(t, acc >= 0.98)
	err = params.Save(util.CacheFile("tests/mnistConv0.params"))
	assert.NilError(t, err)

	net := nn.Bind(mx.CPU, mnistConv0, mx.Dim(10, 1, 28, 28), nil)
	assert.NilError(t, err)
	err = net.LoadParamsFile(util.CacheFile("tests/mnistConv0.params"), false)
	assert.NilError(t, err)
	net.PrintSummary(false)

	ok, err := ng.Measure(net, &mnist.Dataset{}, &ng.Classification{Accuracy: 0.98}, ng.Printing)
	assert.Assert(t, ok)
}
