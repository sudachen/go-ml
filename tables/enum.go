package tables

import (
	"fmt"
	"github.com/sudachen/go-foo/fu"
	"github.com/sudachen/go-ml/internal"
	"golang.org/x/xerrors"
	"reflect"
	"sync"
)

type Meta interface {
	Type() reflect.Type
	Convert(string) (reflect.Value, bool, error)
	Format(reflect.Value, bool) string
}

var enumType = reflect.TypeOf(Enum{})

/*
Enum encapsulate enumeration abstraction in relation to tables
*/
type Enum struct {
	Text  string
	Value int
}

// String return enum string representation
func (e Enum) String() string {
	return e.Text
}

// Enum defines enumerated meta-column with the Enum tipe
func (e Enumset) Enum() Meta {
	return Enumerator{e, &sync.Mutex{}, len(e) != 0}
}

// Enum defines enumerated meta-column with the string type
func (e Enumset) Text() Meta {
	return TextEnumerator{Enumerator{e, &sync.Mutex{}, len(e) != 0}}
}

// Enum defines enumerated meta-column with the int type
func (e Enumset) Integer() Meta {
	return IntegerEnumerator{
		Enumerator{e, &sync.Mutex{}, len(e) != 0},
		fu.KeysOf((map[string]int)(e)).([]string),
	}
}

// Enum defines enumerated meta-column with the float32 type
func (e Enumset) Float32() Meta {
	return Float32Enumerator{
		IntegerEnumerator{
			Enumerator{e, &sync.Mutex{}, len(e) != 0},
			fu.KeysOf((map[string]int)(e)).([]string),
		}}
}

// Enumset is a set of values belongs to one enumeration
type Enumset map[string]int

// Len returns length of enumset aka count of enum values
func (m Enumset) Len() int {
	return len(m)
}

// Enumerator the object enumerates enums in data stream
type Enumerator struct {
	m  Enumset
	mu *sync.Mutex
	ro bool
}

func (ce Enumerator) enumerate(v string) (e int, ok bool) {
	ce.mu.Lock()
	if e, ok = ce.m[v]; !ok {
		if ce.ro {
			panic(xerrors.Errorf("readonly enumset does not have value `%v`" + v))
		}
		ce.m[v] = len(ce.m)
	}
	ce.mu.Unlock()
	return
}

// Type returns the type of column
func (ce Enumerator) Type() reflect.Type {
	return enumType // it's the Enum meta-column
}
func (ce Enumerator) Convert(v string) (reflect.Value, bool, error) {
	if v == "" {
		return reflect.ValueOf(""), true, nil
	}
	e, _ := ce.enumerate(v)
	return reflect.ValueOf(Enum{v, e}), false, nil
}
func (ce Enumerator) Format(x reflect.Value, na bool) string {
	if na {
		return ""
	}
	if x.Type() == enumType {
		text := x.Interface().(Enum).Text
		if _, ok := ce.m[text]; ok {
			return text
		}
	}
	panic(xerrors.Errorf("`%v` is not an enumeration value", x))
}

type IntegerEnumerator struct {
	Enumerator
	rev []string
}

func (ce IntegerEnumerator) Type() reflect.Type {
	return internal.IntType
}

func (ce IntegerEnumerator) Convert(v string) (reflect.Value, bool, error) {
	if v == "" {
		return reflect.ValueOf(""), true, nil
	}
	e, ok := ce.enumerate(v)
	if !ok {
		ce.mu.Lock()
		ce.rev = append(ce.rev, v)
		ce.mu.Unlock()
	}
	return reflect.ValueOf(e), false, nil
}

func (ce IntegerEnumerator) Format(x reflect.Value, na bool) string {
	if na {
		return ""
	}
	if x.Kind() == reflect.String {
		text := x.String()
		if e, ok := ce.m[text]; ok {
			return fmt.Sprint(e)
		}
	}
	panic(xerrors.Errorf("`%v` is not an enumeration value", x))
}

type Float32Enumerator struct{ IntegerEnumerator }

func (ce Float32Enumerator) Type() reflect.Type {
	return internal.Float32Type
}

func (ce Float32Enumerator) Convert(v string) (val reflect.Value, na bool, err error) {
	if val, na, err = ce.IntegerEnumerator.Convert(v); err == nil {
		val = reflect.ValueOf(float32(val.Int()))
	}
	return
}

type TextEnumerator struct{ Enumerator }

func (ce TextEnumerator) Type() reflect.Type {
	return internal.StringType
}

func (ce TextEnumerator) Convert(v string) (reflect.Value, bool, error) {
	if v == "" {
		return reflect.ValueOf(""), true, nil
	}
	ce.enumerate(v)
	return reflect.ValueOf(v), false, nil
}
