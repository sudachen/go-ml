package tables

import (
	"github.com/sudachen/go-ml/mlutil"
	"golang.org/x/xerrors"
	"reflect"
	"unsafe"
)

type Xtensor struct{ T reflect.Type }

func (t *Xtensor) Type() reflect.Type {
	return mlutil.TensorType
}

func (t Xtensor) Convert(value string, field *reflect.Value, _, _ int) (_ bool, err error) {
	z, err := mlutil.DecodeTensor(value)
	if err != nil {
		return
	}
	*field = reflect.ValueOf(z)
	return
}

func tensorOf(field *reflect.Value, tp reflect.Type, width int) (mlutil.Tensor, error) {
	if field.IsValid() {
		return *(*mlutil.Tensor)(unsafe.Pointer(field.Pointer())), nil
	}
	var z mlutil.Tensor
	switch tp {
	case mlutil.Float64:
		z = mlutil.MakeFloat64Tensor(1, 1, width, nil)
	case mlutil.Float32:
		z = mlutil.MakeFloat32Tensor(1, 1, width, nil)
	case mlutil.Fixed8Type:
		z = mlutil.MakeFixed8Tensor(1, 1, width, nil)
	case mlutil.Int:
		z = mlutil.MakeIntTensor(1, 1, width, nil)
	case mlutil.Byte:
		z = mlutil.MakeByteTensor(1, 1, width, nil)
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
	if x.Type() == mlutil.TensorType {
		return x.String()
	}
	panic(xerrors.Errorf("`%v` is not an Xtensor value", x))
}
