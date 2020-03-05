package rdb

import (
	"database/sql"
	"fmt"
	"github.com/sudachen/go-foo/fu"
	"github.com/sudachen/go-ml/internal"
	"github.com/sudachen/go-ml/mlutil"
	"golang.org/x/xerrors"
	"reflect"
)

type IfExists_ int

const (
	ErrorIfExists IfExists_ = iota
	DropIfExists
	AppendIfExists
	InsertUpdateIfExists
)

type Table string
type Query string
type Schema string
type Driver string
type Batch int

type SqlTypeOpt func(string) (string, string, string, bool)

func Describe(names []string, opts []interface{}) (func(string) (string, string, bool), error) {
	drv := fu.StrOption(Driver(""), opts)
	m := map[string]func() (string, string, bool){}
	for _, o := range opts {
		if sto, ok := o.(SqlTypeOpt); ok {
			v, ctp, c, cpk := sto(drv)
			starsub := mlutil.Starsub(v, c)
			exists := false
			for _, n := range names {
				if _, ok := m[n]; !ok {
					if ns, ok := starsub(n); ok {
						colType := ctp
						isPK := cpk
						m[n] = func() (string, string, bool) {
							return colType, ns, isPK
						}
						exists = true
					}
				}
			}
			if !exists {
				return nil, xerrors.Errorf("field %v does not exist in table/query file", v)
			}
		}
	}
	return func(n string) (colType string, colName string, isPK bool) {
		if f, ok := m[n]; ok {
			return f()
		} else {
			colName = n
		}
		return
	}, nil
}

func Column(v string) SqlTypeOpt {
	return func(_ string) (string, string, string, bool) {
		return v, "", v, false
	}
}
func BOOLEAN(v string) SqlTypeOpt {
	return func(_ string) (string, string, string, bool) {
		return v, "BOOLEAN", v, false
	}
}
func SMALLINT(v string) SqlTypeOpt {
	return func(_ string) (string, string, string, bool) {
		return v, "SMALLINT", v, false
	}
}
func INTEGER(v string) SqlTypeOpt {
	return func(_ string) (string, string, string, bool) {
		return v, "INTEGER", v, false
	}
}
func BIGINT(v string) SqlTypeOpt {
	return func(_ string) (string, string, string, bool) {
		return v, "BIGINT", v, false
	}
}
func FLOAT(v string) SqlTypeOpt {
	return func(drv string) (string, string, string, bool) {
		if drv == "postgres" {
			return v, "REAL", v, false
		}
		return v, "FLOAT", v, false
	}
}
func DOUBLE(v string) SqlTypeOpt {
	return func(drv string) (string, string, string, bool) {
		if drv == "postgres" {
			return v, "DOUBLE PRECISION", v, false
		}
		return v, "DOUBLE", v, false
	}
}
func DATE(v string) SqlTypeOpt {
	return func(_ string) (string, string, string, bool) {
		return v, "DATE", v, false
	}
}
func DATETIME(v string) SqlTypeOpt {
	return func(_ string) (string, string, string, bool) {
		return v, "DATETIME", v, false
	}
}
func TIMESTAMP(v string) SqlTypeOpt {
	return func(drv string) (string, string, string, bool) {
		if drv == "sqlite3" {
			return v, "DATETIME", v, false
		}
		return v, "TIMESTAMP", v, false
	}
}

func DECIMAL(v string, prec ...int) SqlTypeOpt {
	s := "DECIMAL"
	if len(prec) > 0 && prec[0] >= 0 {
		scale := 0
		if len(prec) > 1 {
			scale = prec[1]
		}
		s += fmt.Sprintf("(%d,%d)", prec[0], scale)
	}
	return func(_ string) (string, string, string, bool) {
		return v, s, v, false
	}
}

func VARCHAR(v string, length ...int) SqlTypeOpt {
	l := 65536
	if len(length) > 0 {
		l = length[0]
	}
	s := fmt.Sprintf("VARCHAR(%d)", l)
	return func(_ string) (string, string, string, bool) {
		return v, s, v, false
	}
}

func AUTOINCREMENT(v string) SqlTypeOpt {
	return func(drv string) (string, string, string, bool) {
		switch drv {
		case "postgres":
			return v, "SERIAL NOT NULL", v, false
		case "mysql":
			return v, "INTEGER NOT NULL AUTO_INCREMENT", v, false
		}
		return v, "INTEGER NOT NULL AUTOINCREMENT", v, false
	}
}

func (f SqlTypeOpt) PrimaryKey() SqlTypeOpt {
	return func(drv string) (string, string, string, bool) {
		n, t, p, _ := f(drv)
		return n, t, p, true
	}
}

func (f SqlTypeOpt) As(b string) SqlTypeOpt {
	return func(drv string) (string, string, string, bool) {
		n, t, _, k := f(drv)
		return n, t, b, k
	}
}

type SqlScan interface {
	sql.Scanner
	Value() (reflect.Value, bool)
	Reflect() reflect.Type
}

type SqlSmall struct {
	sql.NullInt32
}

func (s *SqlSmall) Scan(value interface{}) error {
	return s.NullInt32.Scan(value)
}

func (s *SqlSmall) Value() (r reflect.Value, ok bool) {
	return reflect.ValueOf(int16(s.Int32)), s.Valid
}

func (s *SqlSmall) Reflect() reflect.Type {
	return internal.Int16Type
}

type SqlInteger struct {
	sql.NullInt32
}

func (s *SqlInteger) Scan(value interface{}) error {
	return s.NullInt32.Scan(value)
}

func (s *SqlInteger) Value() (reflect.Value, bool) {
	return reflect.ValueOf(int(s.Int32)), s.Valid
}

func (s *SqlInteger) Reflect() reflect.Type {
	return internal.IntType
}

type SqlBigint struct {
	sql.NullInt64
}

func (s *SqlBigint) Scan(value interface{}) error {
	return s.NullInt64.Scan(value)
}

func (s *SqlBigint) Value() (reflect.Value, bool) {
	return reflect.ValueOf(s.Int64), s.Valid
}

func (s *SqlBigint) Reflect() reflect.Type {
	return internal.Int64Type
}

type SqlBool struct {
	sql.NullBool
}

func (s *SqlBool) Scan(value interface{}) error {
	return s.NullBool.Scan(value)
}

func (s *SqlBool) Value() (reflect.Value, bool) {
	return reflect.ValueOf(s.Bool), s.Valid
}

func (s *SqlBool) Reflect() reflect.Type {
	return internal.BoolType
}

type SqlString struct {
	sql.NullString
}

func (s *SqlString) Scan(value interface{}) error {
	return s.NullString.Scan(value)
}

func (s *SqlString) Value() (reflect.Value, bool) {
	return reflect.ValueOf(s.String), s.Valid
}

func (s *SqlString) Reflect() reflect.Type {
	return internal.StringType
}

type SqlFloat struct {
	sql.NullFloat64
}

func (s *SqlFloat) Scan(value interface{}) error {
	return s.NullFloat64.Scan(value)
}

func (s *SqlFloat) Value() (reflect.Value, bool) {
	return reflect.ValueOf(float32(s.Float64)), s.Valid
}

func (s *SqlFloat) Reflect() reflect.Type {
	return internal.Float32Type
}

type SqlDouble struct {
	sql.NullFloat64
}

func (s *SqlDouble) Scan(value interface{}) error {
	return s.NullFloat64.Scan(value)
}

func (s *SqlDouble) Value() (reflect.Value, bool) {
	return reflect.ValueOf(s.Float64), s.Valid
}

func (s *SqlDouble) Reflect() reflect.Type {
	return internal.Float64Type
}

type SqlTimestamp struct {
	sql.NullTime
}

func (s *SqlTimestamp) Value() (reflect.Value, bool) {
	return reflect.ValueOf(s.NullTime.Time), s.NullTime.Valid
}

func (s *SqlTimestamp) Reflect() reflect.Type {
	return internal.TsType
}

func (s *SqlTimestamp) Scan(value interface{}) error {
	return s.NullTime.Scan(value)
}
