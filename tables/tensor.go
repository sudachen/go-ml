package tables

import (
	"github.com/sudachen/go-ml/mlutil"
	"golang.org/x/xerrors"
	"reflect"
	"strconv"
	"unsafe"
)

type TensorElement int
const (
	TzeByte TensorElement = iota
	TzeInt
	TzeFixed8
	TzeFloat32
	TzeFloat64
)

type Tensor struct {
	Type TensorElement
	Channels, Height, Width int
	Value unsafe.Pointer
}

//	gets base64-encoded xz-compressed stream as a string prefixed by \xE2\x9C\x97` (✗`)
func DecodeTensor(string) (t *Tensor, err error) {
	return
}

//	returns base64-encoded xz-compressed stream as a string prefixed by \xE2\x9C\x97` (✗`)
func (t *Tensor) String() (r string) {
	return
}

func (t *Tensor) Volume() int {
	return t.Channels * t.Width * t.Height
}

func (t *Tensor) ConvertElem(val string, index int) error {
	switch t.Type {
	case TzeInt:
		v, err := strconv.ParseInt(val,10,64)
		if err != nil { return err }
		(*(*[]int)(t.Value))[index] = int(v)
	case TzeFloat64:
		v, err := strconv.ParseFloat(val,64)
		if err != nil { return err }
		(*(*[]float64)(t.Value))[index] = v
	case TzeFloat32:
		v, err := mlutil.Fast32f(val)
		if err != nil { return err }
		(*(*[]float32)(t.Value))[index] = v
	case TzeFixed8:
		v, err := mlutil.Fast8f(val)
		if err != nil { return err }
		(*(*[]mlutil.Fixed8)(t.Value))[index] = v
	default:
		return xerrors.Errorf("tensor does not konw how to work with type %v",t.Type)
	}
	return nil
}

func (t *Tensor) Interface(index int) interface{} {
	switch t.Type {
	case TzeByte:
		return (*(*[]byte)(t.Value))[index]
	case TzeInt:
		return (*(*[]int)(t.Value))[index]
	case TzeFloat64:
		return (*(*[]float64)(t.Value))[index]
	case TzeFloat32:
		return (*(*[]float32)(t.Value))[index]
	case TzeFixed8:
		return (*(*[]mlutil.Fixed8)(t.Value))[index]
	}
	panic(xerrors.Errorf("tensor does not konw how to work with type %v",t.Type))
}

var TensorType = reflect.TypeOf((*Tensor)(nil))

type Xtensor struct{ T reflect.Type }
func (t *Xtensor) Type() reflect.Type {
	return TensorType
}

func (t Xtensor) Convert(value string, field *reflect.Value, _,_ int) (_ bool, err error) {
	z ,err := DecodeTensor(value)
	if err != nil { return }
	*field =  reflect.ValueOf(z)
	return
}

func tensorOf(field *reflect.Value, tp reflect.Type, width int) (*Tensor,error) {
	if field.IsValid() {
		return (*Tensor)(unsafe.Pointer(field.Pointer())), nil
	}
	z := &Tensor{
		Channels: 1,
		Height: 1,
		Width: width,
	}
	switch tp {
	case mlutil.Float64:
		value := make([]float64,width)
		z.Value = unsafe.Pointer(&value)
		z.Type = TzeFloat64
	case mlutil.Float32:
		value := make([]float32,width)
		z.Value = unsafe.Pointer(&value)
		z.Type = TzeFloat32
	case mlutil.Fixed8Type:
		value := make([]mlutil.Fixed8,width)
		z.Value = unsafe.Pointer(&value)
		z.Type = TzeFixed8
	case mlutil.Int:
		value := make([]int,width)
		z.Value = unsafe.Pointer(&value)
		z.Type = TzeInt
	case mlutil.Byte:
		value := make([]byte,width)
		z.Value = unsafe.Pointer(&value)
		z.Type = TzeByte
	default:
		return nil, xerrors.Errorf("unknown tensor value type "+tp.String())
	}
	*field = reflect.ValueOf(z)
	return z, nil
}

func (t Xtensor) ConvertElm(value string, field *reflect.Value, index, width int) (err error) {
	z, err := tensorOf(field, t.T, width)
	if err != nil { return }
	return z.ConvertElem(value,index)
}

func (Xtensor) Format(x reflect.Value, na bool) string {
	if na {
		return ""
	}
	if x.Type() == TensorType {
		return x.String()
	}
	panic(xerrors.Errorf("`%v` is not an Xtensor value", x))
}
