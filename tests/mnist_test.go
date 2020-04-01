package tests

import (
	"fmt"
	"github.com/sudachen/go-iokit/iokit"
	"github.com/sudachen/go-ml/dataset/mnist"
	"github.com/sudachen/go-ml/fu"
	"github.com/sudachen/go-ml/metrics/classification"
	"github.com/sudachen/go-ml/model"
	"github.com/sudachen/go-ml/nn"
	"github.com/sudachen/go-ml/nn/mx"
	"github.com/sudachen/go-ml/notes"
	"github.com/sudachen/go-ml/xgb"
	"gotest.tools/assert"
	"testing"
)

var mnistMLP0 = nn.Connect(
	&nn.FullyConnected{Size: 128, Activation: nn.ReLU},
	&nn.FullyConnected{Size: 64, Activation: nn.Swish, BatchNorm: true},
	&nn.FullyConnected{Size: 10, Activation: nn.Softmax, BatchNorm: true})

func Test_mnistMLP0(t *testing.T) {
	modelFile := iokit.File(fu.ModelPath("mnist_test_mlp0.zip"))
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
	modelFile := iokit.File(fu.ModelPath("mnist_test_conv0.zip"))

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

func Test_minstXgb(t *testing.T) {
	np := notes.Page{
		Title:  `XGBoost Mnist Test`,
		Footer: `!(http://github.com/sudachen/go-ml)`,
	}.LuckyCreate(iokit.File("mnist_test_xgb.html"))
	defer np.Ensure()

	ds := mnist.Data.RandomFlag("Test", 43, 0.2)
	np.Head("Dataset first lines", ds, 5)
	np.Info("Dataset info", ds)

	modelFile := iokit.File(fu.ModelPath("mnist_test_xgb.zip"))
	metrics := xgb.Model{
		Algorithm:    xgb.TreeBoost,
		Function:     xgb.Softmax,
		LearningRate: 0.3,
		MaxDepth:     10,
		Estimators:   100,
	}.Feed(model.Dataset{
		Source:   ds,
		Label:    "Label",
		Test:     "Test",
		Features: []string{"Image"},
	}).LuckyFit(30, modelFile, &classification.Metrics{Accuracy: 0.96})

	np.Display("Metrics", metrics.Round(3))
	np.Plot("Accuracy evolution by iteration", metrics, &notes.Lines{X: "Iteration", Y: []string{"Accuracy"}, Z: "Test"})
	pred := xgb.LuckyObjectify(modelFile)
	metrics1 := model.LuckyEvaluate(mnist.T10k, "Label", pred, 32, &classification.Metrics{})
	assert.Assert(t, metrics1.Last().Float("Accuracy") >= 0.96)
}
