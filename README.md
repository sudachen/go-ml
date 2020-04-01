[![CircleCI](https://circleci.com/gh/sudachen/go-ml.svg?style=svg)](https://circleci.com/gh/sudachen/go-ml)
[![Maintainability](https://api.codeclimate.com/v1/badges/c66e0431917e286fe342/maintainability)](https://codeclimate.com/github/sudachen/go-ml/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/c66e0431917e286fe342/test_coverage)](https://codeclimate.com/github/sudachen/go-ml/test_coverage)
[![Go Report Card](https://goreportcard.com/badge/github.com/sudachen/go-ml)](https://goreportcard.com/report/github.com/sudachen/go-ml)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)


```golang
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
```
