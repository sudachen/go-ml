package xgb

import (
	"fmt"
	"github.com/sudachen/go-foo/fu"
	"github.com/sudachen/go-foo/lazy"
	"github.com/sudachen/go-ml/internal"
	"github.com/sudachen/go-ml/mlutil"
	"github.com/sudachen/go-ml/xgb/capi"
	"golang.org/x/xerrors"
	"reflect"
)

func (par Model) Fit(dataset mlutil.Dataset, opts ...interface{}) (xgb XGBoost, err error) {
	var dat, labdat []float32
	var dat2, labdat2 []float32

	lj := -1
	kj := -2
	length := 0
	length2 := 0

	features := map[string]int{}
	err = lazy.Source(dataset.Source).Drain(func(v reflect.Value) error {

		if v.Kind() == reflect.Bool {
			return nil
		}
		lr := v.Interface().(mlutil.Struct)

		if lj < 0 {
			lj = fu.IndexOf(dataset.Label, lr.Names)
			if lj < 0 {
				panic(xerrors.Errorf("there is no label column `%v`", dataset.Label))
			}
		}
		if kj == -2 {
			if dataset.Test != "" {
				kj = fu.IndexOf(dataset.Test, lr.Names)
			} else {
				kj = -1
			}
		}

		crss := false
		if kj >= 0 {
			crss = lr.Columns[kj].Int() > 0
		}

		if len(features) == 0 {
			lx := fu.Lexic{}
			for _, k := range dataset.Features {
				lx = append(lx, mlutil.Pattern(k))
			}
			i := 0
			for j, n := range lr.Names {
				if j != lj && j != kj {
					if lx.Accepted(n, true) {
						features[n] = i
						i++
					}
				}
			}
			if len(features) == 0 {
				return xerrors.Errorf("there are no features to learn")
			}
		}

		for i, c := range lr.Columns {
			if c.Kind() != reflect.Float32 {
				c = mlutil.Convert(c, lr.Na.Bit(i), internal.Float32Type)
			}
			if i == lj {
				if crss {
					labdat2 = append(labdat2, c.Interface().(float32))
				} else {
					labdat = append(labdat, c.Interface().(float32))
				}
			} else if i != kj {
				if crss {
					dat2 = append(dat2, c.Interface().(float32))
				} else {
					dat = append(dat, c.Interface().(float32))
				}
			}
		}

		if crss {
			length2++
		} else {
			length++
		}

		return nil
	})

	if err != nil {
		return
	}
	m := matrix32(dat, labdat, length)
	defer m.Free()

	predictName := fu.Option(ResultName("Prediction"), par).String()
	predicts := []string{predictName}

	if kj >= 0 {
		m2 := matrix32(dat2, labdat2, length2)
		defer m2.Free()
		xgb = XGBoost{capi.Create2(m.handle, m2.handle), features, predicts}
	} else {
		xgb = XGBoost{capi.Create2(m.handle), features, predicts}
	}

	for _, o := range par {
		if param, ok := o.(capiparam); ok {
			p, a := param.pair()
			capi.SetParam(xgb.handle, p, a)
		}
	}

	capi.SetParam(xgb.handle, "num_feature", fmt.Sprint(len(features)))
	if obj := fu.Option(objective(""), par).Interface(); obj == Softprob || obj == Softmax {
		x := int(fu.Maxr(fu.Maxr(0, labdat...), labdat2...))
		if x < 0 {
			panic(xerrors.Errorf("labels don't contain enough classes or label values is incorrect"))
		}
		capi.SetParam(xgb.handle, "num_class", fmt.Sprint(x+1))
		if obj == Softprob {
			xgb.predicts = []string{}
			for i := 1; i <= x+1; i++ {
				xgb.predicts = append(xgb.predicts, fmt.Sprintf("%v%v", predictName, i))
			}
		}
	}

	rounds := int(fu.Option(Rounds(1), par).Int())
	for i := 0; i < rounds; i++ {
		capi.Update(xgb.handle, i, m.handle)
	}
	return
}
