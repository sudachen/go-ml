package tables

import "reflect"

type Meta interface {
	Type() reflect.Type
	Convert(string, *reflect.Value, int, int) (bool, error)
	Format(reflect.Value, bool) string
}
