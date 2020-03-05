package tables

import (
	"fmt"
	"github.com/sudachen/go-ml/internal"
	"github.com/sudachen/go-ml/mlutil"
	"reflect"
)

type Cell struct {
	reflect.Value
}

func (c Cell) String() string {
	if c.Kind() == reflect.String {
		return c.Interface().(string)
	}
	return fmt.Sprint(c.Interface())
}

func (c Cell) Int() int {
	return mlutil.Convert(c.Value, false, internal.IntType).Interface().(int)
}

func (c Cell) Int8() int8 {
	return mlutil.Convert(c.Value, false, internal.Int8Type).Interface().(int8)
}

func (c Cell) Int16() int16 {
	return mlutil.Convert(c.Value, false, internal.Int16Type).Interface().(int16)
}

func (c Cell) Int32() int32 {
	return mlutil.Convert(c.Value, false, internal.Int32Type).Interface().(int32)
}

func (c Cell) Int64() int64 {
	return mlutil.Convert(c.Value, false, internal.Int64Type).Interface().(int64)
}

func (c Cell) Uint() uint {
	return mlutil.Convert(c.Value, false, internal.UintType).Interface().(uint)
}

func (c Cell) Uint8() uint8 {
	return mlutil.Convert(c.Value, false, internal.Uint8Type).Interface().(uint8)
}

func (c Cell) Uint16() uint16 {
	return mlutil.Convert(c.Value, false, internal.Uint16Type).Interface().(uint16)
}

func (c Cell) Uint32() uint32 {
	return mlutil.Convert(c.Value, false, internal.Uint32Type).Interface().(uint32)
}

func (c Cell) Uint64() uint64 {
	return mlutil.Convert(c.Value, false, internal.Uint64Type).Interface().(uint64)
}

func (c Cell) Real() float32 {
	return mlutil.Convert(c.Value, false, internal.Float32Type).Interface().(float32)
}

func (c Cell) Float() float64 {
	return mlutil.Convert(c.Value, false, internal.Float64Type).Interface().(float64)
}
