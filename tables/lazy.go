package tables

import (
	"github.com/sudachen/go-foo/lazy"
	"reflect"
	"sync"
)

/*
Lazy creates new lazy transformation stream from the table and empty struct or some transformation function
*/
func (t *Table) Lazy(x interface{}) *lazy.Stream {
	v := reflect.ValueOf(x)
	vt := v.Type()
	flag := &lazy.AtomicFlag{Value: 1}
	stopf := func() { flag.Clear() }
	if v.Kind() == reflect.Struct || (v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Struct) {
		if vt.Kind() == reflect.Ptr {
			vt = vt.Elem()
		}
		getf := func(index int64) reflect.Value {
			if index < int64(t.Len()) && flag.State() {
				return t.GetRow(int(index), vt)
			}
			return reflect.ValueOf(false)
		}
		return &lazy.Stream{Getf: getf, Tp: vt, Stopf: stopf}
	} else if v.Kind() == reflect.Func &&
		vt.NumIn() == 1 && vt.NumOut() == 1 &&
		vt.In(0).Kind() == reflect.Struct &&
		(vt.Out(0).Kind() == reflect.Struct || vt.Out(0).Kind() == reflect.Bool) {
		ti := vt.In(0)
		to := vt.Out(0)
		isFilter := to.Kind() == reflect.Bool
		getf := func(index int64) reflect.Value {
			if index < int64(t.Len()) && flag.State() {
				q := []reflect.Value{t.GetRow(int(index), ti)}
				r := v.Call(q)
				if isFilter {
					if r[0].Bool() {
						return q[0]
					}
					return reflect.ValueOf(true)
				}
				return r[0]
			}
			return reflect.ValueOf(false)
		}
		if isFilter {
			return &lazy.Stream{Getf: getf, Tp: ti, Stopf: stopf}
		}
		return &lazy.Stream{Getf: getf, Tp: to, Stopf: stopf}
	} else {
		panic("only struct{...}, func(struct{...})struct{...} or func(struct{...})bool are allowed as an argument")
	}
}

/*
FillUp fills new table from the transformation source
*/
func FillUp(z *lazy.Stream) *Table {
	c := reflect.MakeChan(reflect.ChanOf(reflect.BothDir, z.Tp), 0)
	go func() {
		index := int64(0)
		for {
			v := z.Next(index)
			index++
			if v.Kind() == reflect.Bool {
				if !v.Bool() {
					break
				}
			} else {
				// not need to use select{send&stop}
				c.Send(v)
			}
		}
		c.Close()
	}()
	return New(c.Interface())
}

/*
ConqFillUp fills new table from the transformation source concurrently
*/
func ConqFillUp(z *lazy.Stream, concurrency int) *Table {
	index := &lazy.AtomicCounter{0}
	wc := &lazy.WaitCounter{Value: 0}
	c := reflect.MakeChan(reflect.ChanOf(reflect.BothDir, z.Tp), concurrency)
	gw := sync.WaitGroup{}
	gw.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func() {
			defer gw.Done()
			for {
				n := index.Inc()
				v := z.Next(n)
				wc.Wait(n)
				if v.Kind() != reflect.Bool {
					// not need to use select{send&stop}
					c.Send(v)
				}
				wc.Inc()
				if v.Kind() == reflect.Bool && !v.Bool() {
					break
				}
			}
		}()
	}
	go func() {
		gw.Wait()
		c.Close()
	}()
	return New(c.Interface())
}
