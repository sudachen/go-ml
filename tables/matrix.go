package tables

import (
	"github.com/sudachen/go-ml/fu"
	"golang.org/x/xerrors"
	"reflect"
)

/*
Matrix the presentation of features and labels as plane []float32 slices
*/
type Matrix struct {
	Features      []float32
	Labels        []float32
	Width, Length int
	LabelsWidth   int // 0 means no labels defined
}

/*
Matrix returns matrix without labels
*/
func (t *Table) Matrix(features []string, least ...int) (m Matrix, err error) {
	_, m, err = t.MatrixWithLabelIf(features, "", "", nil, least...)
	return
}

/*
MatrixWithLabel returns matrix with labels
*/
func (t *Table) MatrixWithLabel(features []string, label string, least ...int) (m Matrix, err error) {
	_, m, err = t.MatrixWithLabelIf(features, label, "", nil, least...)
	return
}

/*
MatrixIf returns two matrices without labels
the first one contains samples with column ifName equal ifValue
the second one - samples with column ifName not equal ifValue
*/
func (t *Table) MatrixIf(features []string, ifName string, ifValue interface{}) (m0, m1 Matrix, err error) {
	return t.MatrixWithLabelIf(features, "", ifName, ifValue)
}

/*
MatrixWithLabelIf returns two matrices with labels
the first one contains samples with column ifName equal ifValue
the second one - samples with column ifName not equal ifValue
*/
func (t *Table) MatrixWithLabelIf(features []string, label string, ifName string, ifValue interface{}, least ...int) (test, train Matrix, err error) {
	L := [2]int{0,t.Len()}
	filter := func(int) int { return 1 }
	if ifName != "" {
		if tc, ok := t.ColIfExists(ifName); ok {
			if a, ok := tc.Inspect().([]bool); ok {
				vt := ifValue.(bool)
				filter = func(i int) int {
					if a[i] == vt {
						return 0
					}
					return 1
				}
			} else if a, ok := tc.Inspect().([]int); ok {
				vt := ifValue.(int)
				filter = func(i int) int {
					if a[i] == vt {
						return 0
					}
					return 1
				}
			} else {
				filter = func(i int) int {
					if tc.Index(i).Value == ifValue {
						return 0
					}
					return 1
				}
			}
			l := t.Len()
			L = [2]int{0,0}
			for i := 0; i < l; i++ {
				L[filter(i)]++
			}
		}
	}

	width := 0
	for _, n := range features {
		c := t.Col(n)
		if c.Type() == fu.TensorType {
			width += c.Inspect().([]fu.Tensor)[0].Volume()
		} else {
			width++
		}
	}

	lwidth := 0

	if label != "" {
		lc := t.Col(label)
		lwidth = 1
		if lc.Type() == fu.TensorType {
			lwidth = lc.Inspect().([]fu.Tensor)[0].Volume()
		}
	}

	for i := range L {
		if L[i] > 0 {
			L[i] = fu.Maxi(L[i], least...)
		}
	}

	mx := []Matrix{
		{make([]float32, L[0]*width), make([]float32, L[0]*lwidth), width, L[0], lwidth},
		{make([]float32, L[1]*width), make([]float32, L[1]*lwidth), width, L[1], lwidth},
	}

	wc := 0

	for _, n := range features {
		if wc, err = t.addToMatrix(filter, mx, t.Col(n), false, wc, width, t.Len()); err != nil {
			return
		}
	}

	if lwidth > 0 {
		if _, err = t.addToMatrix(filter, mx, t.Col(label), true, 0, lwidth, t.Len()); err != nil {
			return
		}
	}

	for i, l := range L {
		if t.Len() < l && t.Len() > 0 {
			for j := t.Len() - 1; j < l; j++ {
				m := mx[i]
				copy(m.Features[j*m.Width:(j+1)*m.Width], m.Features[0:m.Width])
				if m.LabelsWidth > 0 {
					copy(m.Labels[j*m.LabelsWidth:(j+1)*m.LabelsWidth], m.Labels[0:m.LabelsWidth])
				}
			}
		}
	}

	return mx[0], mx[1], nil
}

func (t *Table) addToMatrix(f func(int) int, matrix []Matrix, c *Column, label bool, xc, width, length int) (wc int, err error) {
	where := [2][]float32{
		fu.Ife(label, matrix[0].Labels, matrix[0].Features).([]float32),
		fu.Ife(label, matrix[1].Labels, matrix[1].Features).([]float32),
	}
	wc = xc
	z := [2]int{}
	switch c.Type() {
	case fu.Float32:
		x := c.Inspect().([]float32)
		for j := 0; j < length; j++ {
			jf := f(j)
			where[jf][z[jf]*width+wc] = x[j]
			z[jf]++
		}
		wc++
	case fu.Float64:
		x := c.Inspect().([]float64)
		for j := 0; j < length; j++ {
			jf := f(j)
			where[jf][z[jf]*width+wc] = float32(x[j])
			z[jf]++
		}
		wc++
	case fu.Int:
		x := c.Inspect().([]int)
		for j := 0; j < length; j++ {
			jf := f(j)
			where[jf][z[jf]*width+wc] = float32(x[j])
			z[jf]++
		}
		wc++
	case fu.TensorType:
		x := c.Inspect().([]fu.Tensor)
		vol := x[0].Volume()
		for j := 0; j < length; j++ {
			if x[j].Volume() != vol {
				err = xerrors.Errorf("tensors with different volumes found in one column")
			}
			jf := f(j)
			m := z[jf]
			t := where[jf]
			switch x[j].Type() {
			case fu.Float32:
				y := x[j].Values().([]float32)
				copy(t[m*width+wc:m*width+wc+vol], y)
			case fu.Float64:
				y := x[j].Values().([]float64)
				for k := 0; k < vol; k++ {
					t[m*width+wc+k] = float32(y[k])
				}
			case fu.Byte:
				y := x[j].Values().([]byte)
				for k := 0; k < vol; k++ {
					t[m*width+wc+k] = float32(y[k]) / 256
				}
			case fu.Fixed8Type:
				y := x[j].Values().([]fu.Fixed8)
				for k := 0; k < vol; k++ {
					t[m*width+wc+k] = y[k].Float32()
				}
			case fu.Int:
				y := x[j].Values().([]int)
				for k := 0; k < vol; k++ {
					t[m*width+wc+k] = float32(y[k])
				}
			default:
				return width, xerrors.Errorf("unsupported tensor type %v", x[j].Type)
			}
			z[jf]++
		}
		wc += vol
	default:
		x := c.ExtractAs(fu.Float32, true).([]float32)
		for j := 0; j < length; j++ {
			jf := f(j)
			where[jf][z[jf]*width+wc] = x[j]
			z[jf]++
		}
		wc++
	}
	return
}

/*
AsTable converts raw features representation into Table
*/
func (m Matrix) AsTable(names ...string) *Table {
	columns := make([]reflect.Value, m.Width)
	na := make([]fu.Bits, m.Width)
	for i := range columns {
		c := make([]float32, m.Length, m.Length)
		for j := 0; j < m.Length; j++ {
			c[j] = m.Features[m.Width*j+i]
		}
		columns[i] = reflect.ValueOf(c)
	}
	return MakeTable(names, columns, na, m.Length)
}

/*
AsColumn converts raw features representation into Column
*/
func (m Matrix) AsColumn() *Column {
	if m.Width == 1 {
		return &Column{column: reflect.ValueOf(m.Features[0:m.Length])}
	}
	column := make([]fu.Tensor, m.Length)
	for i := 0; i < m.Length; i++ {
		column[i] = fu.MakeFloat32Tensor(1, 1, m.Width, m.Features[m.Width*i:m.Width*(i+1)])
	}
	return &Column{column: reflect.ValueOf(column)}
}

/*
AsLabelColumn converts raw labels representation into Column
*/
func (m Matrix) AsLabelColumn() *Column {
	if m.LabelsWidth == 1 {
		return &Column{column: reflect.ValueOf(m.Labels[0:m.Length])}
	}
	column := make([]fu.Tensor, m.Length)
	for i := 0; i < m.Length; i++ {
		column[i] = fu.MakeFloat32Tensor(1, 1, m.LabelsWidth, m.Labels[m.LabelsWidth*i:m.LabelsWidth*(i+1)])
	}
	return &Column{column: reflect.ValueOf(column)}
}

func MatrixColumn(dat []float32, length int) *Column {
	if length > 0 {
		return Matrix{dat, nil, len(dat) / length, length, 0}.AsColumn()
	}
	return Col([]float32{})
}
