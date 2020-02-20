package util

import (
	"math/rand"
	"reflect"
	"unsafe"
)

func Index(i int, p interface{}) unsafe.Pointer {
	pv := reflect.ValueOf(p)
	of := pv.Elem().Type().Size() * uintptr(i)
	return unsafe.Pointer(pv.Pointer() + of)
}

func RandomIndex(ln, seed int) []int {
	return rand.New(rand.NewSource(int64(seed))).Perm(ln)
}

