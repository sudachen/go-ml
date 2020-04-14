package tables

import (
	"github.com/sudachen/go-ml/fu"
	"golang.org/x/xerrors"
	"reflect"
)

type Xtensor struct{ T reflect.Type }

func (t *Xtensor) Type() reflect.Type {
	return fu.TensorType
}

func (t Xtensor) Convert(value string, field *reflect.Value, _, _ int) (_ bool, err error) {
	z, err := fu.DecodeTensor(value)
	if err != nil {
		return
	}
	*field = reflect.ValueOf(z)
	return
}

func tensorOf(field *reflect.Value, tp reflect.Type, width int) (fu.Tensor, error) {
	if field.IsValid() {
		return (field.Interface()).(fu.Tensor), nil
	}
	var z fu.Tensor
	switch tp {
	case fu.Float64:
		z = fu.MakeFloat64Tensor(1, 1, width, nil)
	case fu.Float32:
		z = fu.MakeFloat32Tensor(1, 1, width, nil)
	case fu.Fixed8Type:
		z = fu.MakeFixed8Tensor(1, 1, width, nil)
	case fu.Int:
		z = fu.MakeIntTensor(1, 1, width, nil)
	case fu.Byte:
		z = fu.MakeByteTensor(1, 1, width, nil)
	default:
		return z, xerrors.Errorf("unknown tensor value type " + tp.String())
	}
	*field = reflect.ValueOf(z)
	return z, nil
}

func (t Xtensor) ConvertElm(value string, field *reflect.Value, index, width int) (err error) {
	z, err := tensorOf(field, t.T, width)
	if err != nil {
		return
	}
	return z.ConvertElem(value, index)
}

func (Xtensor) Format(x reflect.Value, na bool) string {
	if na {
		return ""
	}
	if x.Type() == fu.TensorType {
		return x.String()
	}
	panic(xerrors.Errorf("`%v` is not an Xtensor value", x))
}
