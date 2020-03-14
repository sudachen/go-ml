package tables

import (
	"github.com/sudachen/go-foo/fu"
	"github.com/sudachen/go-ml/mlutil"
	"golang.org/x/xerrors"
	"reflect"
)

type Matrix struct {
	Features []float32
	Labels []float32
	Width, Length int
	LabelsWidth int // 0 means no labels defined
}

func (t *Table) Matrix(features []string) (m Matrix, err error) {
	m,_,err = t.FillTrainAndTest(features, "", "", nil)
	return
}

func (t *Table) FillFeatures(features []string, splitIfName string, splitIfValue interface{}) (m0,m1 Matrix, err error) {
	return t.FillTrainAndTest(features, "", splitIfName, splitIfValue)
}

func (t *Table) FillTrainAndTest(features []string, label string, testIfName string, testIfValue interface{}) (_,_ Matrix, err error) {
	filter := func(int) int { return 0 }
	if testIfName != "" {
		tc := t.Col(testIfName)
		if a, ok := tc.Inspect().([]bool); ok {
			vt := testIfValue.(bool)
			filter = func(i int) int {
				if a[i] == vt {
					return 1
				} // test matrix
				return 0
			}
		} else if a, ok := tc.Inspect().([]int); ok {
			vt := testIfValue.(int)
			filter = func(i int) int {
				if a[i] == vt {
					return 1
				} // test matrix
				return 0
			}
		} else {
			filter = func(i int) int {
				if tc.Index(i).Value == testIfValue {
					return 1
				}
				return 0
			}
		}
	}

	L := [2]int{}
	for i := 0; i < t.raw.Length; i++ {
		L[filter(i)]++
	}

	width := 0
	for _,n := range features {
		c := t.Col(n)
		if c.Type() == TensorType {
			width+=c.Inspect().([]Tensor)[0].Volume()
		} else {
			width++
		}
	}

	lwidth := 0

	if label != "" {
		lc := t.Col(label)
		lwidth = 1
		if lc.Type() == TensorType {
			lwidth = lc.Inspect().([]Tensor)[0].Volume()
		}
	}

	mx := []Matrix{
		{ make([]float32,L[0]*width), make([]float32,L[0]*lwidth), width, L[0], lwidth },
		{ make([]float32,L[1]*width), make([]float32,L[0]*lwidth), width, L[1], lwidth },
	}

	wc := 0

	for _,n := range features {
		if wc, err = t.addToMatrix(filter, mx, t.Col(n), false, wc, width, t.raw.Length); err != nil {
			return
		}
	}

	if lwidth > 0 {
		if _, err = t.addToMatrix(filter, mx, t.Col(label), true, 0, lwidth, t.raw.Length); err != nil {
			return
		}
	}

	return mx[0], mx[1], nil
}

func (t *Table) addToMatrix(f func(int)int, matrix []Matrix,c *Column, label bool, wc,width,length int) (_ int, err error) {
	where := [2][]float32{
		fu.Ife(label,matrix[0].Labels,matrix[0].Features).([]float32),
		fu.Ife(label,matrix[1].Labels,matrix[1].Features).([]float32),
	}
	z := [2]int{}
	switch c.Type() {
	case mlutil.Float32:
		x := c.Inspect().([]float32)
		for j := 0; j < length; j++ {
			jf := f(j)
			where[jf][z[jf]*width+wc] = x[j]
			z[jf]++
		}
		wc++
	case mlutil.Float64:
		x := c.Inspect().([]float64)
		for j := 0; j < length; j++ {
			jf := f(j)
			where[jf][z[jf]*width+wc] = float32(x[j])
			z[jf]++
		}
		wc++
	case mlutil.Int:
		x := c.Inspect().([]int)
		for j := 0; j < length; j++ {
			jf := f(j)
			where[jf][z[jf]*width+wc] = float32(x[j])
			z[jf]++
		}
		wc++
	case TensorType:
		x := c.Inspect().([]Tensor)
		vol := x[0].Volume()
		for j := 0; j < length; j++ {
			if x[j].Volume() != vol {
				err = xerrors.Errorf("tensors with different volumes found in one column")
			}
			jf := f(j)
			m := z[jf]
			t := where[jf]
			switch x[j].Type {
			case TzeFloat32:
				y := *(*[]float32)(x[j].Value)
				copy(t[m*width+wc:m*width+wc+vol],y)
			case TzeFloat64:
				y := *(*[]float64)(x[j].Value)
				for k :=0; k < vol; k++ {
					t[m*width+wc+k] = float32(y[k])
				}
			case TzeByte:
				y := *(*[]byte)(x[j].Value)
				for k :=0; k < vol; k++ {
					t[m*width+wc+k] = float32(y[k])/256
				}
			case TzeFixed8:
				y := *(*[]mlutil.Fixed8)(x[j].Value)
				for k :=0; k < vol; k++ {
					t[m*width+wc+k] = y[k].Float32()
				}
			case TzeInt:
				y := *(*[]int)(x[j].Value)
				for k :=0; k < vol; k++ {
					t[m*width+wc+k] = float32(y[k])
				}
			default:
				return width, xerrors.Errorf("unsupported tensor type %v", x[j].Type )
			}
			z[jf]++
		}
		wc += vol
	default:
		x := c.ExtractAs(mlutil.Float32, true).([]float32)
		for j := 0; j < length; j++ {
			jf := f(j)
			where[jf][z[jf]*width+wc] = x[j]
			z[jf]++
		}
		wc++
	}
	return
}

func (m Matrix) AsTable(names ...string) *Table {
	columns := make([]reflect.Value,m.Width)
	na := make([]mlutil.Bits,m.Width)
	for i := range columns {
		c := make([]float32,m.Length,m.Length)
		for j :=0; j < m.Length; j++ {
			c[j] = m.Features[m.Width*j+i]
		}
		columns[i] = reflect.ValueOf(c)
	}
	return MakeTable(names,columns,na,m.Length)
}

func (m Matrix) AsColumn() *Column {
	if m.Width == 1 {
		return &Column{column:reflect.ValueOf(m.Features[0:m.Length])}
	}
	column := make([]*Tensor,m.Length)
	for i := 0; i<m.Length; i++ {
		column[i] = MakeFloat32Tensor(1,1, m.Width, m.Features[m.Width*i:m.Width*(i+1)])
	}
	return &Column{column:reflect.ValueOf(column)}
}

