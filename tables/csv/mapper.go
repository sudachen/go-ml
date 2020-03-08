package csv

import (
	"github.com/sudachen/go-ml/mlutil"
	"golang.org/x/xerrors"
	"reflect"
)

type mapper struct {
	CsvCol, TableCol string
	valueType        reflect.Type
	convert          func(string) (value reflect.Value, na bool, err error)
	format           func(value reflect.Value, na bool) string
}

func (m mapper) AutoConvert(column *reflect.Value, na *mlutil.Bits) {
}

func (m mapper) Type() reflect.Type {
	if m.valueType != reflect.Type(nil) {
		return m.valueType
	}
	return mlutil.String
}

func (m mapper) Convert(s string) (value reflect.Value, na bool, err error) {
	if m.convert != nil {
		return m.convert(s)
	}
	return reflect.ValueOf(s), false, nil
}

func (m mapper) Format(v reflect.Value, na bool) string {
	return format(v, na, m.format)
}

func mapFields(header []string, opts []interface{}) (fm []mapper, names []string, err error) {
	fm = make([]mapper, len(header))
	names = make([]string, len(header))
	for _, o := range opts {
		if x, ok := o.(resolver); ok {
			v := x()
			starsub := mlutil.Starsub(v.CsvCol, v.TableCol)
			exists := false
			for i, n := range header {
				if names[i] == "" {
					if c, ok := starsub(n); ok {
						names[i] = c
						fm[i] = v
						exists = true
					}
				}
			}
			if !exists {
				return nil, nil, xerrors.Errorf("field %v does not exist in CSV file", v.CsvCol)
			}
		}
	}
	for i, n := range names {
		if n == "" {
			names[i] = header[i]
		}
	}
	return
}
