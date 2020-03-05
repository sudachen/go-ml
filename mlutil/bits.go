package mlutil

import (
	"math/bits"
)

type Bits struct {
	b []uint
}

const (
	_W = bits.UintSize // word size in bits
)

func (q *Bits) Set(i int, v bool) {
	j := i / _W
	c := j + 1
	n := len(q.b)
	if !v && n < c {
		return
	}
	if n < c {
		x := make([]uint, c)
		if n > 0 {
			copy(x[:n], q.b)
		}
		q.b = x
	}
	m := uint(1) << (i % _W)
	if v {
		q.b[j] |= m
	} else {
		q.b[j] &= ^m
	}
}

func (q Bits) Bit(i int) bool {
	j := int(i / _W)
	m := uint(1) << (i % _W)
	n := len(q.b)
	if j >= n {
		return false
	}
	return (q.b[j] & m) != 0
}

func (q Bits) Slice(from, to int) Bits {
	ql := q.Len()
	if ql <= from {
		return Bits{}
	}
	if to > ql {
		to = ql
	}
	c := ((to - from + _W - 1) / _W)
	x := make([]uint, c)
	rr := from % _W
	rl := _W - rr
	of := from / _W
	for i := range x {
		x[i] = q.b[i+of] >> rr
		if i+1 < len(q.b) {
			x[i] |= q.b[i+of+1] << rl
		}
	}
	x[len(x)-1] &= ^uint(0) >> (_W - ((to - from) % _W))
	r := Bits{x}
	if r.Len() == 0 {
		return Bits{}
	}
	return r
}

func (q Bits) Len() int {
	L := len(q.b) - 1
	for i := range q.b {
		if l := bits.Len(q.b[L-i]); l != 0 {
			return (L-i)*_W + l
		}
	}
	return 0
}

func (q Bits) Copy() Bits {
	if q.Len() == 0 {
		return Bits{}
	}
	c := (q.Len() + _W - 1) / _W
	r := make([]uint, c)
	copy(r, q.b[:c])
	return Bits{r}
}

func (q Bits) Grow(cap int) Bits {
	c := (cap + _W - 1) / _W
	if len(q.b) >= c {
		return q.Copy()
	}
	r := make([]uint, c)
	copy(r[:len(q.b)], q.b)
	return Bits{r}
}

func (q Bits) Append(z Bits, at int) Bits {
	if at < q.Len() {
		panic("bits overflow")
	}
	zl := z.Len()
	if z.Len() == 0 {
		return q.Copy()
	}
	w := q.Grow(at + zl)
	k := (zl + _W - 1) / _W
	j := at / _W
	ll := at % _W
	rl := _W - ll
	for i := 0; i < k; i++ {
		w.b[j+i] |= z.b[i] << ll
		x := z.b[i] >> rl
		if rl != _W && x != 0 {
			w.b[j+i+1] |= z.b[i] >> rl
		}
	}
	return w
}

func FillBits(count int) Bits {
	if count == 0 {
		return Bits{}
	}

	L := (count + _W - 1) / _W
	r := make([]uint, L)
	v := ^uint(0)
	if count < _W {
		v = v >> (_W - count)
		r[0] = v
	} else {
		for i := 0; i < count/_W; i++ {
			r[i] = v
		}
		if count%_W != 0 {
			r[count/_W] = v >> (_W - count%_W)
		}
	}
	return Bits{r}
}

func (q Bits) Repr() string {
	n := q.Len()
	b := make([]byte, n)
	for i := 0; i < n; i++ {
		c := byte('0')
		if (q.b[i/_W]>>(i%_W))&1 != 0 {
			c = byte('1')
		}
		b[i] = c
	}
	return string(b)
}

func (q Bits) String() string {
	n := q.Len()
	b := make([]byte, n+n/8)
	for i, j := 0, 0; i < n; i++ {
		if i != 0 && i%8 == 0 {
			b[j] = byte('.')
			j++
		}
		c := byte('0')
		if (q.b[i/_W]>>(i%_W))&1 != 0 {
			c = byte('1')
		}
		b[j] = c
		j++
	}
	return string(b)
}

func Words(l int) int {
	return (l + _W - 1) / _W
}

func (q Bits) Word(i int) uint {
	if i < len(q.b) {
		return q.b[i]
	}
	return 0
}

func (q Bits) Count() (n int) {
	for _, x := range q.b {
		n += bits.OnesCount(x)
	}
	return
}
