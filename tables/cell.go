package tables

import (
	"fmt"
	"github.com/sudachen/go-ml/internal"
	"github.com/sudachen/go-ml/util"
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
	return util.Convert(c.Value, internal.IntType).(int)
}

func (c Cell) Int8() int8 {
	return util.Convert(c.Value, internal.Int8Type).(int8)
}

func (c Cell) Int16() int16 {
	return util.Convert(c.Value, internal.Int16Type).(int16)
}

func (c Cell) Int32() int32 {
	return util.Convert(c.Value, internal.Int32Type).(int32)
}

func (c Cell) Int64() int64 {
	return util.Convert(c.Value, internal.Int64Type).(int64)
}

func (c Cell) Uint() uint {
	return util.Convert(c.Value, internal.UintType).(uint)
}

func (c Cell) Uint8() uint8 {
	return util.Convert(c.Value, internal.Uint8Type).(uint8)
}

func (c Cell) Uint16() uint16 {
	return util.Convert(c.Value, internal.Uint16Type).(uint16)
}

func (c Cell) Uint32() uint32 {
	return util.Convert(c.Value, internal.Uint32Type).(uint32)
}

func (c Cell) Uint64() uint64 {
	return util.Convert(c.Value, internal.Uint64Type).(uint64)
}

func (c Cell) Float() float32 {
	return util.Convert(c.Value, internal.FloatType).(float32)
}

func (c Cell) Float64() float64 {
	return util.Convert(c.Value, internal.Float64Type).(float64)
}
