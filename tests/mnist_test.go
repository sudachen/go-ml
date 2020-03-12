package tests

import (
	"github.com/sudachen/go-foo/fu"
	"github.com/sudachen/go-ml/datasets/mnist"
	"github.com/sudachen/go-ml/nn"
	"github.com/sudachen/go-ml/nn/mx"
	"github.com/sudachen/go-ml/nn/ng"
	"gotest.tools/assert"
	"testing"
	"time"
)

var mnistMLP0 = nn.Connect(
	&nn.FullyConnected{Size: 128, Activation: nn.ReLU},
	&nn.FullyConnected{Size: 64, Activation: nn.Swish, BatchNorm: true},
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

	/*
	   type Tensor struct {
	   	Type reflect.Type
	   	Value reflect.Value // slice of float64, float32, uint8, int values ordered as CHW
	   	Channels,Height,Width int
	   }
	   //	gets base64-encoded xz-compressed stream as a string prefixed by \xE2\x9C\x97` (✗`)
	   func (*) Decode(string)
	   //	returns base64-encoded xz-compressed stream as a string prefixed by \xE2\x9C\x97` (✗`)
	   func () String() string

	   	csv.Source(fu.File(),tables.Meta(mlutil.Pixmap(mlutil.Gray32f),"image").As("Image"))

	   	// network params
	   	csv.Source(fu.File(),
	   		csv.String("name").As("Name"),
	   		csv.Tensor32f("values").As("Values"))

	   	csv.Sink(fu.File(),
	      		csv.Column("Name").As("name"),
	      		csv.Column("Values").As("values"))

	   	tensor is packed into xz compressed stream and is encoded as base64 string
	   	it's prefiexed by \xE2\x9C\x97` (✗`)

	      	pred, metrics,err := ??.Model{}.Feed(...).Fit()
	   	pred, metrics := ??.model{}.Feed(...).LuckyFit()
	       metrics, err := ??.model{}.Feed(...).Estimate()
	      	metrics := ??.model{}.Feed(...).LuckyEstimate()

	   	metrics := nn.Model{
	   		Network:   mnistMLP0,
	   		Initial:   Source, // name,data
	                           // layer1,✗`MQoyCjMKNAo=
	   		Optimizer: &nn.Adam{Lr: .001},
	   		Loss:      &nn.LabelCrossEntropyLoss{},
	   		Input:     mx.Dim(1, 28, 28),
	   		Epochs:    10,
	   		Seed:      42,
	   		Batch:     32 }.
	   	Feed(mlutil.Dataset{
	   		Source: mnist.Source.Kfold(42, 5, "Fold").Parallel().MemCache(),
	   		Features: []string{"Image"},
	   		Kfold: "Fold",
	   		Label: "Label"
	   	}).
	   	//LuckyFit(mlutil.Verbose(mlutil.Printing,1*time.Second))
	   	LuckyEstimate(
	   		mlutil.Verbose(nn.Printing,1*time.Second),
	   		metrics.Accuracy{})

	   	??? LuckyGridFit?
	*/
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
	err = net1.SaveParamsFile(fu.CacheFile("tests/mnistMLP0.params"))
	assert.NilError(t, err)

	net2 := nn.Bind(mx.CPU, mnistMLP0, mx.Dim(10, 1, 28, 28), nil)
	err = net2.LoadParamsFile(fu.CacheFile("tests/mnistMLP0.params"), false)
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
	err = params.Save(fu.CacheFile("tests/mnistConv0.params"))
	assert.NilError(t, err)

	net := nn.Bind(mx.CPU, mnistConv0, mx.Dim(10, 1, 28, 28), nil)
	assert.NilError(t, err)
	err = net.LoadParamsFile(fu.CacheFile("tests/mnistConv0.params"), false)
	assert.NilError(t, err)
	net.PrintSummary(false)

	ok, err := ng.Measure(net, &mnist.Dataset{}, &ng.Classification{Accuracy: 0.98}, ng.Printing)
	assert.Assert(t, ok)
}
