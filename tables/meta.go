package tables

import "reflect"

type Meta interface {
	Type() reflect.Type
	Convert(string) (reflect.Value, bool, error)
	Format(reflect.Value, bool) string
}
