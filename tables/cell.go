package tables

import (
	"fmt"
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
	return mlutil.Convert(c.Value, false, mlutil.Int).Interface().(int)
}

func (c Cell) Int8() int8 {
	return mlutil.Convert(c.Value, false, mlutil.Int8).Interface().(int8)
}

func (c Cell) Int16() int16 {
	return mlutil.Convert(c.Value, false, mlutil.Int16).Interface().(int16)
}

func (c Cell) Int32() int32 {
	return mlutil.Convert(c.Value, false, mlutil.Int32).Interface().(int32)
}

func (c Cell) Int64() int64 {
	return mlutil.Convert(c.Value, false, mlutil.Int64).Interface().(int64)
}

func (c Cell) Uint() uint {
	return mlutil.Convert(c.Value, false, mlutil.Uint).Interface().(uint)
}

func (c Cell) Uint8() uint8 {
	return mlutil.Convert(c.Value, false, mlutil.Uint8).Interface().(uint8)
}

func (c Cell) Uint16() uint16 {
	return mlutil.Convert(c.Value, false, mlutil.Uint16).Interface().(uint16)
}

func (c Cell) Uint32() uint32 {
	return mlutil.Convert(c.Value, false, mlutil.Uint32).Interface().(uint32)
}

func (c Cell) Uint64() uint64 {
	return mlutil.Convert(c.Value, false, mlutil.Uint64).Interface().(uint64)
}

func (c Cell) Real() float32 {
	return mlutil.Convert(c.Value, false, mlutil.Float32).Interface().(float32)
}

func (c Cell) Float() float64 {
	return mlutil.Convert(c.Value, false, mlutil.Float64).Interface().(float64)
}
