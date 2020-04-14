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
	"github.com/sudachen/go-ml/xgb"
	"gotest.tools/assert"
	"testing"
)

var mnistMLP0 = nn.Connect(
	&nn.FullyConnected{Size: 128, Activation: nn.ReLU, Dropout: 0.3},
	&nn.FullyConnected{Size: 64, Activation: nn.Swish, BatchNorm: true},
	&nn.FullyConnected{Size: 10, Activation: nn.Softmax, BatchNorm: true})

func Test_mnistMLP0(t *testing.T) {
	modelFile := iokit.File(fu.ModelPath("mnist_test_mlp0.zip"))
	report := nn.Model{
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
	}).LuckyTrain(model.Training{
		Iterations: 5,
		ModelFile:  modelFile,
		Metrics:    &classification.Metrics{Accuracy: 0.961},
		Score:      classification.ErrorScore,
	})

	fmt.Println(report.TheBest, report.Score)
	fmt.Println(report.History.Round(5))
	assert.Assert(t, classification.Accuracy(report.Test) >= 0.96)

	net1 := nn.LuckyObjectify(modelFile) //.Gpu()
	lr := model.LuckyEvaluate(mnist.T10k, "Label", net1, 32, &classification.Metrics{})
	fmt.Println(lr)
	assert.Assert(t, classification.Accuracy(lr) >= 0.96)
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

	report := nn.Model{
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
	}).LuckyTrain(model.Training{
		Iterations: 15,
		ModelFile:  modelFile,
		Metrics:    &classification.Metrics{Accuracy: 0.981},
		Score:      classification.ErrorScore,
	})

	fmt.Println(report.TheBest, report.Score)
	fmt.Println(report.History.Round(5))
	assert.Assert(t, classification.Accuracy(report.Test) >= 0.98)

	net1 := nn.LuckyObjectify(modelFile) //.Gpu()
	lr := model.LuckyEvaluate(mnist.T10k, "Label", net1, 32, &classification.Metrics{})
	fmt.Println(lr)
	assert.Assert(t, classification.Accuracy(lr) >= 0.98)
}

func Test_minstXgb(t *testing.T) {
	modelFile := iokit.File(fu.ModelPath("mnist_test_xgb.zip"))
	report := xgb.Model{
		Algorithm:    xgb.TreeBoost,
		Function:     xgb.Softmax,
		LearningRate: 0.54,
		MaxDepth:     7,
		Estimators:   8,
		Extra:        map[string]interface{}{"tree_method": "hist"},
	}.Feed(model.Dataset{
		Source:   mnist.Data.RandomFlag(model.TestCol, 42, 0.1),
		Features: mnist.Features,
	}).LuckyTrain(model.Training{
		Iterations: 30,
		ModelFile:  modelFile,
		Metrics:    &classification.Metrics{Accuracy: 0.96},
		Score:      classification.AccuracyScore,
	})

	fmt.Println(report.TheBest, report.Score)
	fmt.Println(report.History.Round(5))
	assert.Assert(t, classification.Accuracy(report.Test) >= 0.96)

	pred := xgb.LuckyObjectify(modelFile)
	lr := model.LuckyEvaluate(mnist.T10k, model.LabelCol, pred, 32, &classification.Metrics{})
	fmt.Println(lr.Round(5))
	assert.Assert(t, classification.Accuracy(lr) >= 0.96)
}
