package tests

import (
	"fmt"
	"github.com/sudachen/go-foo/fu"
	"github.com/sudachen/go-ml/dataset/mnist"
	"github.com/sudachen/go-ml/metrics/classification"
	"github.com/sudachen/go-ml/mlutil"
	"github.com/sudachen/go-ml/model"
	"github.com/sudachen/go-ml/nn"
	"github.com/sudachen/go-ml/nn/mx"
	"gotest.tools/assert"
	"testing"
)

var mnistMLP0 = nn.Connect(
	&nn.FullyConnected{Size: 128, Activation: nn.ReLU},
	&nn.FullyConnected{Size: 64, Activation: nn.Swish, BatchNorm: true},
	&nn.FullyConnected{Size: 10, Activation: nn.Softmax, BatchNorm: true})

func Test_mnistMLP0(t *testing.T) {
	modelFile := fu.File(mlutil.ModelPath("mnist_test_mlp0.zip"))
	metrics := nn.Model{
		Network:   mnistMLP0,
		Optimizer: &nn.Adam{Lr: .001},
		Loss:      &nn.LabelCrossEntropyLoss{},
		Input:     mx.Dim(1, 28, 28),
		Seed:      42,
		BatchSize: 32,
		//Context:   mx.GPU,
	}.Feed(model.Dataset{
		Source:   mnist.Data.RandomFlag("Test", 42, 0.2),
		Label:    "Label",
		Test:     "Test",
		Features: []string{"Image"},
	}).LuckyFit(5, modelFile, &classification.Metrics{Accuracy: 0.96})

	fmt.Println(metrics)
	assert.Assert(t, metrics.Last().Float("Accuracy") >= 0.96)

	net1 := nn.LuckyObjectify(modelFile) //.Gpu()
	metrics1 := model.LuckyEvaluate(mnist.T10k, "Label", net1, 32, &classification.Metrics{})
	fmt.Println(metrics1)
	assert.Assert(t, metrics1.Last().Float("Accuracy") >= 0.96)
}

var mnistConv0 = nn.Connect(
	&nn.Convolution{Channels: 24, Kernel: mx.Dim(3, 3), Activation: nn.ReLU},
	&nn.MaxPool{Kernel: mx.Dim(2, 2), Stride: mx.Dim(2, 2)},
	&nn.Convolution{Channels: 32, Kernel: mx.Dim(5, 5), Activation: nn.ReLU, BatchNorm: true},
	&nn.MaxPool{Kernel: mx.Dim(2, 2), Stride: mx.Dim(2, 2)},
	&nn.FullyConnected{Size: 32, Activation: nn.Swish, BatchNorm: true, Dropout: 0.33},
	&nn.FullyConnected{Size: 10, Activation: nn.Softmax})

func Test_mnistConv0(t *testing.T) {
	modelFile := fu.File(mlutil.ModelPath("mnist_test_conv0.zip"))

	metrics := nn.Model{
		Network:   mnistConv0,
		Optimizer: &nn.Adam{Lr: .001},
		Loss:      &nn.LabelCrossEntropyLoss{},
		Input:     mx.Dim(1, 28, 28),
		Seed:      42,
		BatchSize: 32,
		//Context:   mx.GPU,
	}.Feed(model.Dataset{
		Source:   mnist.Data.RandomFlag("Test", 42, 0.2),
		Label:    "Label",
		Test:     "Test",
		Features: []string{"Image"},
	}).LuckyFit(5, modelFile, &classification.Metrics{Accuracy: 0.98})

	fmt.Println(metrics)
	assert.Assert(t, metrics.Last().Float("Accuracy") >= 0.98)

	net1 := nn.LuckyObjectify(modelFile) //.Gpu()
	metrics1 := model.LuckyEvaluate(mnist.T10k, "Label", net1, 32, &classification.Metrics{})
	fmt.Println(metrics1)
	assert.Assert(t, metrics1.Last().Float("Accuracy") >= 0.98)
}
