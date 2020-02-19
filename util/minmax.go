package util

import "reflect"

/*
Min returns minimal value
*/
func Min(a reflect.Value) reflect.Value {
	return a.Index(MinIndex(a))
}

/*
MinIndex returns index of minimal value
*/
func MinIndex(a reflect.Value) int {
	if a.Kind() != reflect.Slice {
		panic("only slice is allowed as an argument")
	}
	N := a.Len()
	d := 0
	r := a.Index(0)
	for i := 1; i < N; i++ {
		j := a.Index(i)
		if Less(j, r) {
			r = j
			d = i
		}
	}
	return d
}

/*
Max returns maximal value
*/
func Max(a reflect.Value) reflect.Value {
	return a.Index(MaxIndex(a))
}

/*
MaxIndex returns index of maximal value
*/
func MaxIndex(a reflect.Value) int {
	if a.Kind() != reflect.Slice {
		panic("only slice is allowed as an argument")
	}
	N := a.Len()
	d := 0
	r := a.Index(0)
	for i := 1; i < N; i++ {
		j := a.Index(i)
		if Less(r, j) {
			r = j
			d = i
		}
	}
	return d
}
