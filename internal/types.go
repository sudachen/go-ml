package internal

import (
	"reflect"
	"time"
)

var IntType = reflect.TypeOf(int(0))
var Int8Type = reflect.TypeOf(int8(0))
var Int16Type = reflect.TypeOf(int16(0))
var Int32Type = reflect.TypeOf(int32(0))
var Int64Type = reflect.TypeOf(int64(0))
var UintType = reflect.TypeOf(uint(0))
var Uint8Type = reflect.TypeOf(uint8(0))
var Uint16Type = reflect.TypeOf(uint16(0))
var Uint32Type = reflect.TypeOf(uint32(0))
var Uint64Type = reflect.TypeOf(uint64(0))
var FloatType = reflect.TypeOf(float32(0))
var Float64Type = reflect.TypeOf(float64(0))
var StringType = reflect.TypeOf("")
var BoolType = reflect.TypeOf(true)
var TsType = reflect.TypeOf(time.Time{})
