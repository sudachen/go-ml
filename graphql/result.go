package graphql

import (
	"reflect"
	"strconv"
	"time"
)

type Result reflect.Value

func (q Result) Q(a interface{}) Result {
	v := reflect.Value(q)
	for v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	if v.Kind() == reflect.Map {
		return Result(v.MapIndex(reflect.ValueOf(a)))
	} else if v.Kind() == reflect.Slice {
		return Result(v.Index(int(reflect.ValueOf(a).Int())))
	}
	return Result{}
}

func (q Result) V(a interface{}) (r reflect.Value) {
	v := reflect.Value(q)
	for v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	if v.Kind() == reflect.Map {
		r = v.MapIndex(reflect.ValueOf(a))
	} else if v.Kind() == reflect.Slice {
		r = v.Index(int(reflect.ValueOf(a).Int()))
	}
	for r.Kind() == reflect.Interface {
		r = r.Elem()
	}
	return
}

func (q Result) List() []Result {
	v := (reflect.Value)(q)
	for v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	if v.Kind() == reflect.Slice {
		r := make([]Result, v.Len(), v.Len())
		for i := range r {
			r[i] = Result(v.Index(i))
		}
		return r
	}
	panic("is not list")
}

func (q Result) Chan() chan Result {
	v := (reflect.Value)(q)
	for v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	if v.Kind() == reflect.Slice {
		c := make(chan Result)
		go func() {
			defer close(c)
			for i := 0; i < v.Len(); i++ {
				c <- Result(v.Index(i))
			}
		}()
		return c
	}
	panic("is not list")
}

func (q Result) String(a interface{}) string {
	v := q.V(a)
	if !v.IsValid() {
		return ""
	}
	return v.String()
}

func (q Result) Float32(a interface{}) float32 {
	v := q.V(a)
	if !v.IsValid() {
		return 0
	}
	switch v.Kind() {
	case reflect.String:
		f, err := strconv.ParseFloat(v.String(), 32)
		if err != nil {
			panic(err.Error())
		}
		return float32(f)
	case reflect.Float64, reflect.Float32:
		return float32(v.Float())
	default:
		return float32(v.Int())
	}
}

func (q Result) Float64(a interface{}) float64 {
	v := q.V(a)
	if !v.IsValid() {
		return 0
	}
	switch v.Kind() {
	case reflect.String:
		f, err := strconv.ParseFloat(v.String(), 32)
		if err != nil {
			panic(err.Error())
		}
		return f
	case reflect.Float64, reflect.Float32:
		return v.Float()
	default:
		return float64(v.Int())
	}
}

func (q Result) Int(a interface{}) int {
	v := q.V(a)
	if !v.IsValid() {
		return 0
	}
	switch v.Kind() {
	case reflect.String:
		i, err := strconv.ParseInt(v.String(), 10, 32)
		if err != nil {
			panic(err.Error())
		}
		return int(i)
	case reflect.Float32, reflect.Float64:
		return int(v.Float())
	}
	return int(v.Int())
}

func (q Result) Time(a interface{}) time.Time {
	v := q.V(a)
	if !v.IsValid() {
		return time.Time{}
	}
	if v.Kind() == reflect.String {
		t, err := time.Parse(time.RFC3339, v.String())
		if err != nil {
			panic(err.Error())
		}
		return t
	}
	panic("not string")
}

func (q Result) Fill(a interface{}) {
	av := reflect.ValueOf(a).Elem() // a is a ptr to struct
	at := av.Type()
	for i := 0; i < at.NumField(); i++ {
		f := av.Field(i)
		ft := at.Field(i)
		switch f.Kind() {
		case reflect.String:
			f.Set(reflect.ValueOf(q.String(ft.Name)))
		case reflect.Float32:
			f.Set(reflect.ValueOf(q.Float32(ft.Name)))
		case reflect.Float64:
			f.Set(reflect.ValueOf(q.Float64(ft.Name)))
		case reflect.Int, reflect.Int32, reflect.Int64:
			f.Set(reflect.ValueOf(q.Int(ft.Name)))
		default:
			if f.Type() == reflect.TypeOf(time.Time{}) {
				f.Set(reflect.ValueOf(q.Time(ft.Name)))
			} else {
				panic("unsupported type of filed " + ft.Name)
			}
		}
	}
}
