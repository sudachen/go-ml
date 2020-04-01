package mnist

import (
	"encoding/binary"
	"github.com/sudachen/go-iokit/iokit"
	"github.com/sudachen/go-ml/fu"
	"github.com/sudachen/go-ml/lazy"
	"github.com/sudachen/go-ml/tables"
	"github.com/sudachen/go-zorros/zorros"
	"io/ioutil"
	"reflect"
)

func source(x string) iokit.Input {
	return iokit.Compressed(
		iokit.Url("http://yann.lecun.com/exdb/mnist/"+x,
			iokit.Cache("go-ml/dataset/mnist/"+x)))
}

var imagesSource = source("train-images-idx3-ubyte.gz")
var labelsSource = source("train-labels-idx1-ubyte.gz")
var t10kImagesSource = source("t10k-images-idx3-ubyte.gz")
var t10kLabelsSource = source("t10k-labels-idx1-ubyte.gz")

var Data tables.Lazy = func() lazy.Stream { return lazyread(imagesSource, labelsSource) }
var T10k tables.Lazy = func() lazy.Stream { return lazyread(t10kImagesSource, t10kLabelsSource) }
var Full tables.Lazy = Data.False("Test").Chain(T10k.True("Test"))

func lazyread(imagesSource, labelsSource iokit.Input) lazy.Stream {
	imr, err := imagesSource.Open()
	if err != nil {
		return lazy.Error(err)
	}
	defer imr.Close()
	lar, err := labelsSource.Open()
	if err != nil {
		return lazy.Error(err)
	}
	defer lar.Close()

	images, err := ioutil.ReadAll(imr)
	if err != nil {
		return lazy.Error(err)
	}
	labels, err := ioutil.ReadAll(lar)
	if err != nil {
		return lazy.Error(err)
	}

	if 0x00000803 != binary.BigEndian.Uint32(images[0:4]) {
		return lazy.Error(zorros.Errorf("not mnist images file"))
	}
	if 0x00000801 != binary.BigEndian.Uint32(labels[0:4]) {
		return lazy.Error(zorros.Errorf("not mnist labels file"))
	}
	count := int(binary.BigEndian.Uint32(images[4:8]))
	if count != int(binary.BigEndian.Uint32(labels[4:8])) {
		return lazy.Error(zorros.Errorf("incorrect samples count"))
	}
	width := int(binary.BigEndian.Uint32(images[8:12]))
	height := int(binary.BigEndian.Uint32(images[12:16]))
	vol := width * height

	names := []string{"Label", "Image"}
	f := fu.AtomicFlag{Value: 0}
	return func(index uint64) (value reflect.Value, err error) {
		if index == lazy.STOP {
			f.Set()
		} else if !f.State() && int(index) < count {
			tz := fu.MakeByteTensor(1, height, width, images[16+vol*int(index):16+vol*(int(index)+1)])
			return reflect.ValueOf(fu.MakeStruct(names, int(labels[8+int(index)]), tz)), nil
		}
		return fu.False, nil
	}
}
