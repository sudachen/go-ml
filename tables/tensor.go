package tables

import (
	"github.com/sudachen/go-ml/mlutil"
	"golang.org/x/xerrors"
	"reflect"
)

type Tensor struct {
	Type                    reflect.Type
	Value                   reflect.Value // slice of float64, float32, uint8, int values ordered as CHW
	Channels, Height, Width int
}

//	gets base64-encoded xz-compressed stream as a string prefixed by \xE2\x9C\x97` (✗`)
func (t *Tensor) Decode(string) (err error) {
	if t.Type == nil {
		t.Type = mlutil.Float32
	}
	return
}

//	returns base64-encoded xz-compressed stream as a string prefixed by \xE2\x9C\x97` (✗`)
func (t Tensor) String() (r string) {
	return
}

var tensorType = reflect.TypeOf(Tensor{})

type Xtensor struct{ T reflect.Type }

func (t Xtensor) Type() reflect.Type {
	return tensorType // it's the Enum meta-column
}
func (t Xtensor) Convert(v string) (reflect.Value, bool, error) {
	x := Tensor{}
	return reflect.ValueOf(x), false, nil
}
func (Xtensor) Format(x reflect.Value, na bool) string {
	if na {
		return ""
	}
	if x.Type() == tensorType {
	}
	panic(xerrors.Errorf("`%v` is not an Xtensor value", x))
}
