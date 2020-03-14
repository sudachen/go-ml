package tables

import (
	"github.com/sudachen/go-foo/lazy"
	"github.com/sudachen/go-ml/mlutil"
	"reflect"
)

type Batch lazy.Source

func (zf Lazy) Batch(length int) Batch {
	return func() lazy.Stream {
		z := zf()
		wc := lazy.WaitCounter{Value:0}
		columns := []reflect.Value{}
		na := []mlutil.Bits{}
		names := []string{}
		ac := lazy.AtomicCounter{Value:0}

		return func(index uint64) (v reflect.Value, err error) {
			v, err = z(index)
			if index == lazy.STOP || err != nil {
				wc.Stop()
				return
			}

			x := mlutil.True
			if wc.Wait(index) {
				if v.Kind() == reflect.Bool {
					if !v.Bool() {
						n := int(ac.Value % uint64(length))
						if ac.Value != 0 {
							if n == 0 { n = length }
							v = reflect.ValueOf(MakeTable(names,columns,na,n))
						}
						wc.Stop()
					}
					wc.Inc()
					return v, nil
				}

				lr := v.Interface().(mlutil.Struct)
				ndx := ac.PostInc()
				n := int(ndx % uint64(length))

				if n == 0 {
					if ndx != 0 {
						x = reflect.ValueOf(MakeTable(names,columns,na,length))
					}
					names = lr.Names
					width := len(names)
					columns = make([]reflect.Value,width)
					for i := range columns {
						columns[i] = reflect.MakeSlice(reflect.SliceOf(lr.Columns[i].Type()),0,length)
					}
					na = make([]mlutil.Bits,width)
				}

				for i := range lr.Names {
					columns[i] = reflect.Append(columns[i],lr.Columns[i])
					na[i].Set(n,lr.Na.Bit(i))
				}

				wc.Inc()
				return x, nil
			}
			return mlutil.False, nil
		}
	}
}

func (zf Batch) Flat() Lazy {
	return func() lazy.Stream {
		z := zf()
		wc := lazy.WaitCounter{Value:0}
		ac := lazy.AtomicCounter{Value:0}
		t := (*Table)(nil)
		row := 0
		return func(index uint64) (v reflect.Value, err error) {
			v = mlutil.False
			if index == lazy.STOP || err != nil {
				wc.Stop()
				return
			}
			if wc.Wait(index) {
				if t == nil {
					v, err = z(ac.PostInc())
					if err != nil || (v.Kind() == reflect.Bool  && !v.Bool()) {
						wc.Stop()
						return
					}
					if v.Kind() != reflect.Bool {
						t = v.Interface().(*Table)
						row = 0
					}
				}
				if t != nil {
					v = reflect.ValueOf(t.Struct(row))
					row++
					if row >= t.Len() { t = nil }
				}
				wc.Inc()
				return v, nil
			}
			return mlutil.False, nil
		}
	}
}

func (zf Batch) Transform(tx func(*Table)(*Table,error)) Batch {
	return func() lazy.Stream {
		z := zf()
		f := lazy.AtomicFlag{Value:0}
		return func(index uint64) (v reflect.Value, err error) {
			v, err = z(index)
			if index == lazy.STOP || err != nil {
				f.Set()
				return
			}
			if !f.State() {
				if v.Kind() != reflect.Bool {
					lr := v.Interface().(*Table)
					t, err := tx(lr)
					if err != nil {
						f.Set()
						return mlutil.False, err
					}
					return reflect.ValueOf(t), nil
				}
				if v.Bool() {
					return mlutil.True, nil
				}
				f.Set()
			}
			return mlutil.False, nil
		}
	}
}
