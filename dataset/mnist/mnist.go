package mnist

import (
	"encoding/binary"
	"github.com/sudachen/go-foo/fu"
	"github.com/sudachen/go-foo/lazy"
	"github.com/sudachen/go-ml/mlutil"
	"github.com/sudachen/go-ml/tables"
	"golang.org/x/xerrors"
	"io/ioutil"
	"reflect"
)

func source(x string) fu.Input {
	return fu.Compressed(
		fu.External("http://yann.lecun.com/exdb/mnist/"+x,
			fu.Cached("go-ml/dataset/mnist/"+x)))
}

var imagesSource = source("train-images-idx3-ubyte.gz")
var labelsSource = source("train-labels-idx1-ubyte.gz")
var t10kImagesSource = source("t10k-images-idx3-ubyte.gz")
var t10kLabelsSource = source("t10k-labels-idx1-ubyte.gz")

var Data tables.Lazy = func() lazy.Stream { return lazyread(imagesSource, labelsSource) }
var T10k tables.Lazy = func() lazy.Stream { return lazyread(t10kImagesSource, t10kLabelsSource) }
var Full tables.Lazy = Data.False("Test").Chain(T10k.True("Test"))

func lazyread(imagesSource, labelsSource fu.Input) lazy.Stream {
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
		return lazy.Error(xerrors.Errorf("not mnist images file"))
	}
	if 0x00000801 != binary.BigEndian.Uint32(labels[0:4]) {
		return lazy.Error(xerrors.Errorf("not mnist labels file"))
	}
	count := int(binary.BigEndian.Uint32(images[4:8]))
	if count != int(binary.BigEndian.Uint32(labels[4:8])) {
		return lazy.Error(xerrors.Errorf("incorrect samples count"))
	}
	width := int(binary.BigEndian.Uint32(images[8:12]))
	height := int(binary.BigEndian.Uint32(images[12:16]))
	vol := width * height

	names := []string{"Label", "Image"}
	f := lazy.AtomicFlag{Value: 0}
	return func(index uint64) (value reflect.Value, err error) {
		if index == lazy.STOP {
			f.Set()
		} else if !f.State() && int(index) < count {
			tz := mlutil.MakeByteTensor(1, height, width, images[16+vol*int(index):16+vol*(int(index)+1)])
			return reflect.ValueOf(mlutil.MakeStruct(names, int(labels[8+int(index)]), tz)), nil
		}
		return mlutil.False, nil
	}
}
