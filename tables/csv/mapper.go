package csv

import (
	"github.com/sudachen/go-ml/internal"
	"github.com/sudachen/go-ml/mlutil"
	"golang.org/x/xerrors"
	"reflect"
)

type mapper struct {
	CsvCol, TableCol string
	valueType        reflect.Type
	convert          func(string) (value reflect.Value, na bool, err error)
}

func (m mapper) AutoConvert(column *reflect.Value, na *internal.Bits) {
}

func (m mapper) Type() reflect.Type {
	if m.valueType != reflect.Type(nil) {
		return m.valueType
	}
	return internal.StringType
}

func (m mapper) Convert(s string) (value reflect.Value, na bool, err error) {
	if m.convert != nil {
		return m.convert(s)
	}
	return reflect.ValueOf(s), false, nil
}

func mapFields(header []string, opts []interface{}) (fm []mapper, names []string, err error) {
	fm = make([]mapper, len(header))
	names = make([]string, len(header))
	for _, o := range opts {
		if x, ok := o.(resolver); ok {
			v := x.resolve()
			i := mlutil.IndexOf(v.CsvCol, header)
			if i < 0 {
				return nil, nil, xerrors.Errorf("field %v does not exist in CSV file", v.CsvCol)
			}
			names[i] = v.TableCol
			fm[i] = v
		}
	}
	return
}
