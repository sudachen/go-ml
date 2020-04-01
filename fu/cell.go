package fu

import (
	"fmt"
	"reflect"
	"strings"
)

type Cell struct {
	reflect.Value
}

func (c Cell) Text() string {
	if c.Kind() == reflect.String {
		return c.Interface().(string)
	}
	if c.Type() == TensorType {
		z := c.Interface().(Tensor)
		s := []string{}
		for i := 0; i < 4; i++ {
			if i == 3 || i >= z.Volume() {
				s = append(s, ">")
				break
			} else if i < z.Volume() {
				s = append(s, fmt.Sprint(z.Index(i)))
			}
		}
		ch, h, w := z.Dimension()
		return fmt.Sprintf("(%dx%dx%d){%v}", ch, h, w, strings.Join(s, ","))
	}
	return fmt.Sprint(c.Interface())
}

func (c Cell) String() string { return c.Text() }

func (c Cell) Int() int {
	return Convert(c.Value, false, Int).Interface().(int)
}

func (c Cell) Int8() int8 {
	return Convert(c.Value, false, Int8).Interface().(int8)
}

func (c Cell) Int16() int16 {
	return Convert(c.Value, false, Int16).Interface().(int16)
}

func (c Cell) Int32() int32 {
	return Convert(c.Value, false, Int32).Interface().(int32)
}

func (c Cell) Int64() int64 {
	return Convert(c.Value, false, Int64).Interface().(int64)
}

func (c Cell) Uint() uint {
	return Convert(c.Value, false, Uint).Interface().(uint)
}

func (c Cell) Uint8() uint8 {
	return Convert(c.Value, false, Uint8).Interface().(uint8)
}

func (c Cell) Uint16() uint16 {
	return Convert(c.Value, false, Uint16).Interface().(uint16)
}

func (c Cell) Uint32() uint32 {
	return Convert(c.Value, false, Uint32).Interface().(uint32)
}

func (c Cell) Uint64() uint64 {
	return Convert(c.Value, false, Uint64).Interface().(uint64)
}

func (c Cell) Real() float32 {
	return Convert(c.Value, false, Float32).Interface().(float32)
}

func (c Cell) Float() float64 {
	return Convert(c.Value, false, Float64).Interface().(float64)
}
