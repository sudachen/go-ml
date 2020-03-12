package csv

import (
	"github.com/sudachen/go-foo/fu"
	"github.com/sudachen/go-ml/mlutil"
	"golang.org/x/xerrors"
	"reflect"
)

type formatter func(value reflect.Value, na bool) string
type converter func(value string, field *reflect.Value, index,width int) (bool,error)
type mapper struct {
	CsvCol, TableCol string
	valueType        reflect.Type
	convert          converter
	format           formatter
	group            bool
	field,index      int
	width 			 int
	name             string
}

func Mapper(ccol, tcol string, t reflect.Type, conv converter, form formatter) mapper {
	return mapper{ccol,tcol, t, conv, form, false, 0 ,0, 0, "" }
}

func (m mapper) Group() bool {
	return m.group
}

func (m mapper) Type() reflect.Type {
	if m.valueType != reflect.Type(nil) {
		return m.valueType
	}
	return mlutil.String
}

func (m mapper) Convert(value string, field *reflect.Value, index, width int) (na bool, err error) {
	if m.convert != nil {
		return m.convert(value, field, index, width)
	}
	*field = reflect.ValueOf(value)
	return
}

func (m mapper) Format(v reflect.Value, na bool) string {
	return format(v, na, m.format)
}

func mapFields(header []string, opts []interface{}) (fm []mapper, names []string, err error) {
	fm = make([]mapper, len(header))
	names = make([]string, 0, len(header))
	mask := mlutil.Bits{}
	for _, o := range opts {
		if x, ok := o.(resolver); ok {
			v := x()
			exists := false
			if v.group {
				like := mlutil.Pattern(v.CsvCol)
				for i, n := range header {
					if !mask.Bit(i) && like(n) {
						v.name = v.TableCol
						fm[i] = v
						mask.Set(i,true)
						exists = true
						v.index++
					}
				}
			} else {
				starsub := mlutil.Starsub(v.CsvCol, v.TableCol)
				for i, n := range header {
					if !mask.Bit(i) {
						if c, ok := starsub(n); ok {
							v.name = c
							fm[i] = v
							exists = true
						}
					}
				}
			}
			if !exists {
				return nil, nil, xerrors.Errorf("field %v does not exist in CSV file", v.CsvCol)
			}
		}
	}
	width := make([]int,len(header))
	for i := range fm {
		if fm[i].name == "" { fm[i].name = header[i] }
		j := fu.IndexOf(fm[i].name,names)
		if j < 0 {
			j = len(names)
			names = append(names,fm[i].name)
		}
		fm[i].field = j
		width[j]++
	}
	for i := range fm {
		if fm[i].group {
			fm[i].width = width[fm[i].field]
		}
	}
	return
}
