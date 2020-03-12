package mlutil

import (
	"fmt"
	"golang.org/x/xerrors"
	"strconv"
)

// value in range [-1..1] with 2 digits precession
type Fixed8 struct{int8}

func (v Fixed8) String() string {
	return fmt.Sprint(v.Float32())
}

func (v Fixed8) Float32() float32 {
	return float32(v.int8)/100
}

func AsFixed8(v float32) Fixed8 {
	return Fixed8{int8(v*100) }
}

func Fast8f(s string) (v8 Fixed8, err error) {
	v,err := Fast32f(s)
	if err != nil {return}
	if v > 1.27 || v < -1.27 {
		return Fixed8{}, xerrors.Errorf("fixed8 value out of range [-1.27...1.27] : %v", v)
	}
	v8 = AsFixed8(v)
	return
}

var fast32iTable = []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
func Fast32f(s string) (float32, error) {
	sign := float32(1)
	q := 0
	exp := 0
	if s[0] == '-' {
		sign = float32(-1)
		s = s[1:]
	}
	for _, c := range s {
		if c == '.' {
			exp = 1
		} else if c >= '0' && c <= '9' {
			q = q*10 + fast32iTable[c-'0']
			exp *= 10
		} else {
			return Slow32f(s)
		}
	}
	f := float32(q)
	if exp > 0 {
		f /= float32(exp)
	}
	return f * sign, nil
}

func Slow32f(s string) (float32, error) {
	v, err := strconv.ParseFloat(s, 32)
	return float32(v), err
}
