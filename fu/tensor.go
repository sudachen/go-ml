package fu

import (
	"reflect"
	"strconv"
)

type Tze int

const (
	TzeByte Tze = iota
	TzeInt
	TzeFixed8
	TzeFloat32
	TzeFloat64
)

type dimension struct{ Channels, Height, Width int }

func (d dimension) Volume() int              { return d.Channels * d.Width * d.Height }
func (d dimension) Dimension() (c, h, w int) { return d.Channels, d.Height, d.Width }

type tensor32f struct {
	dimension
	values []float32
}

type tensor64f struct {
	dimension
	values []float64
}

type tensor8u struct {
	dimension
	values []byte
}

type tensori struct {
	dimension
	values []int
}

type tensor8f struct {
	dimension
	values []Fixed8
}

func (t tensor32f) ConvertElem(val string, index int) (err error) {
	t.values[index], err = Fast32f(val)
	return
}

func (t tensor64f) ConvertElem(val string, index int) (err error) {
	t.values[index], err = strconv.ParseFloat(val, 64)
	return
}

func (t tensori) ConvertElem(val string, index int) (err error) {
	i, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return
	}
	t.values[index] = int(i)
	return
}

func (t tensor8f) ConvertElem(val string, index int) (err error) {
	t.values[index], err = Fast8f(val)
	return
}

func (t tensor8u) ConvertElem(val string, index int) (err error) {
	i, err := strconv.ParseInt(val, 10, 8)
	if err != nil {
		return
	}
	t.values[index] = byte(i)
	return
}

func (t tensori) Index(index int) interface{}   { return t.values[index] }
func (t tensor8f) Index(index int) interface{}  { return t.values[index] }
func (t tensor8u) Index(index int) interface{}  { return t.values[index] }
func (t tensor32f) Index(index int) interface{} { return t.values[index] }
func (t tensor64f) Index(index int) interface{} { return t.values[index] }

func (t tensori) Values() interface{}   { return t.values }
func (t tensor8f) Values() interface{}  { return t.values }
func (t tensor8u) Values() interface{}  { return t.values }
func (t tensor32f) Values() interface{} { return t.values }
func (t tensor64f) Values() interface{} { return t.values }

func (t tensori) Type() reflect.Type   { return Int }
func (t tensor8f) Type() reflect.Type  { return Fixed8Type }
func (t tensor8u) Type() reflect.Type  { return Byte }
func (t tensor32f) Type() reflect.Type { return Float32 }
func (t tensor64f) Type() reflect.Type { return Float64 }

func (t tensori) HotOne() (j int) {
	for i, v := range t.values {
		if t.values[j] < v {
			j = i
		}
	}
	return
}

func (t tensor8f) HotOne() (j int) {
	for i, v := range t.values {
		if t.values[j].int8 < v.int8 {
			j = i
		}
	}
	return
}

func (t tensor8u) HotOne() (j int) {
	for i, v := range t.values {
		if t.values[j] < v {
			j = i
		}
	}
	return
}

func (t tensor32f) HotOne() (j int) {
	for i, v := range t.values {
		if t.values[j] < v {
			j = i
		}
	}
	return
}

func (t tensor64f) HotOne() (j int) {
	for i, v := range t.values {
		if t.values[j] < v {
			j = i
		}
	}
	return
}

func (t tensori) String() string   { return "" }
func (t tensor8f) String() string  { return "" }
func (t tensor8u) String() string  { return "" }
func (t tensor32f) String() string { return "" }
func (t tensor64f) String() string { return "" }

func (t tensori) Floats32() (r []float32) {
	r = make([]float32, len(t.values))
	for i, v := range t.values {
		r[i] = float32(v)
	}
	return
}

func (t tensor8f) Floats32() (r []float32) {
	r = make([]float32, len(t.values))
	for i, v := range t.values {
		r[i] = v.Float32()
	}
	return
}

func (t tensor8u) Floats32() (r []float32) {
	r = make([]float32, len(t.values))
	for i, v := range t.values {
		r[i] = float32(v) / 256
	}
	return
}

func (t tensor64f) Floats32() (r []float32) {
	r = make([]float32, len(t.values))
	for i, v := range t.values {
		r[i] = float32(v)
	}
	return
}

func (t tensor32f) Floats32() []float32 { return t.values }

type tensor interface {
	Dimension() (c, h, w int)
	Volume() int
	Type() reflect.Type
	Values() interface{}
	Index(index int) interface{}
	String() string
	ConvertElem(val string, index int) error
	HotOne() int
	Floats32() []float32
}

type Tensor struct{ tensor }

//	gets base64-encoded xz-compressed stream as a string prefixed by \xE2\x9C\x97` (✗`)
func DecodeTensor(string) (t Tensor, err error) {
	return
}

func MakeFloat64Tensor(channels, height, width int, values []float64, docopy ...bool) Tensor {
	v := values
	if values != nil {
		if len(docopy) > 0 && docopy[0] {
			v := make([]float64, len(values))
			copy(v, values)
		}
	} else {
		v = make([]float64, channels*height*width)
	}
	x := tensor64f{
		dimension: dimension{
			Channels: channels,
			Height:   height,
			Width:    width,
		},
		values: v,
	}
	return Tensor{x}
}

func MakeFloat32Tensor(channels, height, width int, values []float32, docopy ...bool) Tensor {
	v := values
	if values != nil {
		if len(docopy) > 0 && docopy[0] {
			v := make([]float32, len(values))
			copy(v, values)
		}
	} else {
		v = make([]float32, channels*height*width)
	}
	x := tensor32f{
		dimension: dimension{
			Channels: channels,
			Height:   height,
			Width:    width,
		},
		values: v,
	}
	return Tensor{x}
}

func MakeByteTensor(channels, height, width int, values []byte, docopy ...bool) Tensor {
	v := values
	if values != nil {
		if len(docopy) > 0 && docopy[0] {
			v := make([]byte, len(values))
			copy(v, values)
		}
	} else {
		v = make([]byte, channels*height*width)
	}
	x := tensor8u{
		dimension: dimension{
			Channels: channels,
			Height:   height,
			Width:    width},
		values: v,
	}
	return Tensor{x}
}

func MakeFixed8Tensor(channels, height, width int, values []Fixed8, docopy ...bool) Tensor {
	v := values
	if values != nil {
		if len(docopy) > 0 && docopy[0] {
			v := make([]Fixed8, len(values))
			copy(v, values)
		}
	} else {
		v = make([]Fixed8, channels*height*width)
	}
	x := tensor8f{
		dimension: dimension{
			Channels: channels,
			Height:   height,
			Width:    width},
		values: v,
	}
	return Tensor{x}
}

func MakeIntTensor(channels, height, width int, values []int, docopy ...bool) Tensor {
	v := values
	if values != nil {
		if len(docopy) > 0 && docopy[0] {
			v := make([]int, len(values))
			copy(v, values)
		}
	} else {
		v = make([]int, channels*height*width)
	}
	x := tensori{
		dimension: dimension{
			Channels: channels,
			Height:   height,
			Width:    width},
		values: v,
	}
	return Tensor{x}
}

var TensorType = reflect.TypeOf(Tensor{})
