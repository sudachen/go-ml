package rdb

import (
	"database/sql"
	"github.com/sudachen/go-ml/internal"
	"reflect"
)

type IfExists_ int

const (
	ErrorIfExists IfExists_ = iota
	DropIfExists
	AppendIfExists
	InsertUpdateIfExists
)

type PrimaryKey string
type Schema string
type Driver string
type Batch int

type BOOLEAN string
type SMALLINT string
type INTEGER string
type AUTOINCREMENT string
type BIGINT string
type FLOAT string
type DOUBLE string
type DATE string
type DATETIME string
type TIMESTAMP string

type DECIMAL_ struct {
	Name             string
	Precision, Scale int
}

func DECIMAL(name string, prec ...int) DECIMAL_ {
	r := DECIMAL_{Name: name, Precision: -1, Scale: -1}
	if len(prec) > 0 {
		r.Precision = prec[0]
		r.Scale = 0
	}
	if len(prec) > 1 {
		r.Scale = prec[1]
	}
	return r
}

type VARCHAR_ struct {
	Name   string
	Length int
}

func VARCHAR(name string, length ...int) VARCHAR_ {
	r := VARCHAR_{Name: name, Length: 65536}
	if len(length) > 0 {
		r.Length = length[0]
	}
	return r
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
