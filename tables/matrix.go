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

type Filler struct{
	features []string
	label  string
	test   string
	value  bool
	table *Table
}

func (t *Table) For(features ...string) Filler {
	return Filler{table: t, features: features}
}

func (f Filler) If(column string) (fx Filler) {
	fx = f
	f.test = column
	f.value = true
	return
}

func (f Filler) IfNot(column string) (fx Filler) {
	fx = f
	fx.test = column
	fx.value = false
	return
}

func (f Filler) Label(column string) (fx Filler) {
	fx = f
	fx.label = column
	return
}

func (f Filler) filter() func(int)bool {
	if f.test != "" {
		x := f.table.Col(f.test)
		if y, ok := x.Inspect().([]bool); ok {
			return func(i int) bool { return y[i] == f.value }
		} else if y, ok := x.Inspect().([]int); ok {
			return func(i int) bool { return y[i] != 0 == f.value }
		}
		return func(i int) bool { return x.Index(i).Int() != 0 == f.value }
	}
	return nil
}

func (f Filler) Matrix() (matrix Matrix, err error) {
	filter := f.filter()
	matrix, err = f.table.fillMatrix(filter,f.features...)
	if err != nil { return }
	if f.label != "" {
		c := f.table.Col(f.label)
		matrix.LabelsWidth = 1
		if c.Type() == TensorType {
			matrix.LabelsWidth = c.Inspect().([]Tensor)[0].Volume()
		}
		matrix.Labels = make([]float32,matrix.LabelsWidth*matrix.Length)
		_, err = f.table.addToMatrix(filter, matrix, f.label, 0, true)
	}
	return
}

func (t *Table) addToMatrix(f func(int)bool, matrix Matrix, column string, i int, label ...bool) (width int, err error) {
	width = i
	where := fu.Ife(fu.Fnzb(label...),matrix.Labels,matrix.Features).([]float32)
	c := t.Col(column)
	switch c.Type() {
	case mlutil.Float32:
		x := c.Inspect().([]float32)
		if f != nil {
			for j, k := 0, 0; j < matrix.Length; j++ {
				if f(j) {
					where[k*matrix.Width+width] = x[j]
					k++
				}
			}
		} else {
			for j := 0; j < matrix.Length; j++ {
				where[j*matrix.Width+width] = x[j]
			}
		}
		width++
	case mlutil.Float64:
		x := c.Inspect().([]float64)
		if f != nil {
			for j, k := 0, 0; j < matrix.Length; j++ {
				if f(j) {
					where[k*matrix.Width+width] = float32(x[j])
					k++
				}
			}
		} else {
			for j := 0; j < matrix.Length; j++ {
				matrix.Features[j*matrix.Width+width] = float32(x[j])
			}
		}
		width++
	case mlutil.Int:
		x := c.Inspect().([]int)
		if f != nil {
			for j, k := 0, 0; j < matrix.Length; j++ {
				if f(j) {
					where[k*matrix.Width+width] = float32(x[j])
					k++
				}
			}
		} else {
			for j := 0; j < matrix.Length; j++ {
				matrix.Features[j*matrix.Width+width] = float32(x[j])
			}
		}
		width++
	case TensorType:
		x := c.Inspect().([]Tensor)
		vol := x[0].Volume()
		m := 0
		for j := 0; j < matrix.Length; j++ {
			if x[j].Volume() != vol {
				err = xerrors.Errorf("feature %v containes different volume tensors",column)
			}
			if f == nil || f(j) {
				switch x[j].Type {
				case TzeFloat32:
					y := *(*[]float32)(x[j].Value)
					copy(matrix.Features[m*matrix.Width+width:m*matrix.Width+width+vol],y)
				case TzeFloat64:
					y := *(*[]float64)(x[j].Value)
					for k :=0; k < vol; k++ {
						matrix.Features[m*matrix.Width+width+k] = float32(y[k])
					}
				case TzeByte:
					y := *(*[]byte)(x[j].Value)
					for k :=0; k < vol; k++ {
						matrix.Features[m*matrix.Width+width+k] = float32(y[k])/256
					}
				case TzeFixed8:
					y := *(*[]mlutil.Fixed8)(x[j].Value)
					for k :=0; k < vol; k++ {
						matrix.Features[m*matrix.Width+width+k] = y[k].Float32()
					}
				case TzeInt:
					y := *(*[]int)(x[j].Value)
					for k :=0; k < vol; k++ {
						matrix.Features[m*matrix.Width+width+k] = float32(y[k])
					}
				default:
					return width, xerrors.Errorf("unsupported tensor type %v", x[j].Type )
				}
				m++
			}
		}
		width += vol
	default:
		x := c.ExtractAs(mlutil.Float32, true).([]float32)
		if f != nil {
			for j, k := 0, 0; j < matrix.Length; j++ {
				if f(j) {
					where[k*matrix.Width+width] = x[j]
					k++
				}
			}
		} else {
			for j := 0; j < matrix.Length; j++ {
				matrix.Features[j*matrix.Width+width] = x[j]
			}
		}
		width++
	}
	return
}

func (t *Table) fillMatrix(f func(int)bool, features ...string) (matrix Matrix, err error) {

	matrix.Length = t.FilteredLen(f)

	for _,n := range features {
		c := t.Col(n)
		if c.Type() == TensorType {
			matrix.Width+=c.Inspect().([]Tensor)[0].Volume()
		} else {
			matrix.Width++
		}
	}

	matrix.Features = make([]float32, matrix.Width*matrix.Length)

	width := 0
	for _, n := range features {
		width, err = t.addToMatrix(f,matrix,n,width)
		if err != nil { return}
	}

	return
}

func FromMatrix(m Matrix, names ...string) *Table {
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
