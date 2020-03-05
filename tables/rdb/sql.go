package rdb

import (
	"database/sql"
	"fmt"
	"github.com/sudachen/go-foo/fu"
	"github.com/sudachen/go-foo/lazy"
	"github.com/sudachen/go-ml/base"
	"github.com/sudachen/go-ml/internal"
	"github.com/sudachen/go-ml/tables"
	"golang.org/x/xerrors"
	"io"
	"reflect"
	"strings"
	//	_ "github.com/go-sql-driver/mysql"
	//	_ "github.com/lib/pq"
	//	_ "github.com/mattn/go-sqlite3"
)

func Read(source interface{}, opts ...interface{}) (*tables.Table, error) {
	return Source(source, opts...).Collect()
}

func Write(source interface{}, t *tables.Table, opts ...interface{}) error {
	return t.Lazy().Drain(Sink(source, opts...))
}

type dontclose bool

func connectDB(source interface{}, opts []interface{}) (db *sql.DB, o []interface{}, err error) {
	o = opts
	if url, ok := source.(string); ok {
		drv, conn := splitDriver(url)
		o = append(o, Driver(drv))
		db, err = sql.Open(drv, conn)
	} else if db, ok = source.(*sql.DB); !ok {
		err = xerrors.Errorf("unknown database source %v", source)
	} else {
		o = append(o, dontclose(true))
	}
	return
}

func Source(source interface{}, opts ...interface{}) tables.Lazy {
	return func() lazy.Stream {
		db, opts, err := connectDB(source, opts)
		cls := io.Closer(&fu.CloserChain{})
		if !fu.BoolOption(dontclose(false), opts) {
			cls = db
		}
		if err != nil {
			tables.SourceError(xerrors.Errorf("database connection error: %w", err))
		}
		drv := fu.StrOption(Driver(""), opts)
		schema := fu.StrOption(Schema(""), opts)
		if schema != "" {
			switch drv {
			case "mysql":
				_, err = db.Exec("use " + schema)
			case "postgres":
				_, err = db.Exec("set search_path to " + schema)
			}
		}
		if err != nil {
			cls.Close()
			return lazy.Error(xerrors.Errorf("query error: %w", err))
		}
		query := fu.StrOption(Query(""), opts)
		if query == "" {
			table := fu.StrOption(Table(""), opts)
			if table != "" {
				query = "select * from " + table
			} else {
				panic("there is no query or table")
			}
		}
		rows, err := db.Query(query)
		if err != nil {
			cls.Close()
			return lazy.Error(xerrors.Errorf("query error: %w", err))
		}
		cls = &fu.CloserChain{rows, cls}
		tps, err := rows.ColumnTypes()
		if err != nil {
			cls.Close()
			return lazy.Error(xerrors.Errorf("get types error: %w", err))
		}
		ns, err := rows.Columns()
		if err != nil {
			cls.Close()
			return lazy.Error(xerrors.Errorf("get names error: %w", err))
		}
		x := make([]interface{}, len(ns))
		describe, err := Describe(ns, opts)
		if err != nil {
			cls.Close()
			return lazy.Error(err)
		}
		names := make([]string, len(ns))
		for i, n := range ns {
			var s SqlScan
			colType, colName, _ := describe(n)
			if colType != "" {
				s = scanner(colType)
			} else {
				s = scanner(tps[i].DatabaseTypeName())
			}
			x[i] = s
			names[i] = colName
		}

		wc := lazy.WaitCounter{Value: 0}
		f := lazy.AtomicFlag{Value: 0}

		return func(index uint64) (reflect.Value, error) {
			if index == lazy.STOP {
				wc.Stop()
				return reflect.ValueOf(false), nil
			}
			if wc.Wait(index) {
				end := !rows.Next()
				if !end {
					rows.Scan(x...)
					lr := base.Struct{Names: names, Columns: make([]reflect.Value, len(ns))}
					for i := range x {
						y := x[i].(SqlScan)
						v, ok := y.Value()
						if !ok {
							lr.Na.Set(i, true)
						}
						lr.Columns[i] = v
					}
					wc.Inc()
					return reflect.ValueOf(lr), nil
				}
				wc.Stop()
			}
			if f.Set() {
				cls.Close()
			}
			return reflect.ValueOf(false), nil
		}
	}
}

func splitDriver(url string) (string, string) {
	q := strings.SplitN(url, ":", 2)
	return q[0], q[1]
}

func scanner(q string) SqlScan {
	switch q {
	case "VARCHAR", "TEXT", "CHAR", "STRING":
		return &SqlString{}
	case "INT8", "SMALLINT", "INT2":
		return &SqlSmall{}
	case "INTEGER", "INT", "INT4":
		return &SqlInteger{}
	case "BIGINT":
		return &SqlBigint{}
	case "BOOLEAN":
		return &SqlBool{}
	case "DECIMAL", "NUMERIC", "REAL", "DOUBLE", "FLOAT8":
		return &SqlDouble{}
	case "FLOAT", "FLOAT4":
		return &SqlFloat{}
	case "DATE", "DATETIME", "TIMESTAMP":
		return &SqlTimestamp{}
	default:
		if strings.Index(q, "VARCHAR(") == 0 ||
			strings.Index(q, "CHAR(") == 0 {
			return &SqlString{}
		}
		if strings.Index(q, "DECIMAL(") == 0 ||
			strings.Index(q, "NUMERIC(") == 0 {
			return &SqlDouble{}
		}
	}
	panic("unknown column type " + q)
}

func batchInsertStmt(tx *sql.Tx, names []string, pk []bool, lines int, table string, opts []interface{}) (stmt *sql.Stmt, err error) {
	drv := fu.StrOption(Driver(""), opts)
	ifExists := fu.Option(ErrorIfExists, opts).Interface().(IfExists_)
	L := len(names)
	q1 := " values "
	for j := 0; j < lines; j++ {
		q1 += "("
		if drv == "postgres" {
			for k := range names {
				q1 += fmt.Sprintf("$%d,", j*L+k+1)
			}
		} else {
			q1 += strings.Repeat("?,", L)
		}
		q1 = q1[:len(q1)-1] + "),"
	}
	q := "insert into " + table + "(" + strings.Join(names, ",") + ")" + q1[:len(q1)-1]

	if ifExists == InsertUpdateIfExists {
		if len(pk) > 0 {
			q += " on duplicate key update "
			for i, n := range names {
				if !pk[i] {
					q += " " + n + " = values(" + n + "),"
				}
			}
			q = q[:len(q)-1]
		}
	}
	stmt, err = tx.Prepare(q)
	return
}

func Sink(source interface{}, opts ...interface{}) tables.Sink {
	db, opts, err := connectDB(source, opts)
	cls := io.Closer(&fu.CloserChain{})
	if !fu.BoolOption(dontclose(false), opts) {
		cls = db
	}
	if err != nil {
		return tables.SinkError(xerrors.Errorf("database connection error: %w", err))
	}
	drv := fu.StrOption(Driver(""), opts)

	schema := fu.StrOption(Schema(""), opts)
	if schema != "" {
		switch drv {
		case "mysql":
			_, err = db.Exec("use " + schema)
		case "postgres":
			_, err = db.Exec("set search_path to " + schema)
		}
	}
	if err != nil {
		cls.Close()
		return tables.SinkError(xerrors.Errorf("query error: %w", err))
	}

	tx, err := db.Begin()
	if err != nil {
		cls.Close()
		return tables.SinkError(xerrors.Errorf("database begin transaction error: %w", err))
	}

	table := fu.StrOption(Table(""), opts)
	if table == "" {
		panic("there is no table")
	}
	if fu.Option(ErrorIfExists, opts).Interface().(IfExists_) == DropIfExists {
		_, err := tx.Exec(sqlDropQuery(table, opts...))
		if err != nil {
			cls.Close()
			return tables.SinkError(xerrors.Errorf("drop table error: %w", err))
		}
	}

	batchLen := fu.IntOption(Batch(1), opts)
	var stmt *sql.Stmt
	created := false
	batch := []interface{}{}
	names := []string{}
	pk := []bool{}
	return func(val reflect.Value) (err error) {
		var describe func(int) (string, string, bool)
		if val.Kind() == reflect.Bool {
			if val.Bool() {
				if len(batch) > 0 {
					if stmt, err = batchInsertStmt(tx, names, pk, len(batch)/len(names), table, opts); err == nil {
						if _, err = stmt.Exec(batch...); err == nil {
							cls = fu.CloserChain{stmt, cls}
						}
					}
				}
				if err == nil {
					err = tx.Commit()
				}
			}
			cls.Close()
			return
		}
		lr := val.Interface().(base.Struct)
		names = make([]string, len(lr.Names))
		pk = make([]bool, len(lr.Names))
		drv := fu.StrOption(Driver(""), opts)
		dsx, err := Describe(lr.Names, opts)
		if err != nil {
			cls.Close()
			return
		}
		describe = func(i int) (colType, colName string, isPk bool) {
			v := lr.Names[i]
			colType, colName, isPk = dsx(v)
			if colType == "" {
				colType = sqlTypeOf(lr.Columns[i].Type(), drv)
			}
			return
		}
		for i := range names {
			_, names[i], pk[i] = describe(i)
		}
		if !created {
			_, err = tx.Exec(sqlCreateQuery(lr, table, describe, opts))
			if err != nil {
				cls.Close()
				return xerrors.Errorf("create table error: %w", err)
			}
			created = true
		}
		if len(batch)/len(names) >= batchLen {
			if stmt == nil {
				stmt, err = batchInsertStmt(tx, names, pk, len(batch)/len(names), table, opts)
				if err != nil {
					return err
				}
				cls = &fu.CloserChain{stmt, cls}
			}
			_, err = stmt.Exec(batch...)
			if err != nil {
				return err
			}
			batch = batch[:0]
		}
		for i := range lr.Names {
			if lr.Na.Bit(i) {
				batch = append(batch, nil)
			} else {
				batch = append(batch, lr.Columns[i].Interface())
			}
		}
		return
	}
}

func sqlCreateQuery(lr base.Struct, table string, describe func(int) (string, string, bool), opts []interface{}) string {
	pk := []string{}
	query := "create table "

	ifExists := fu.Option(ErrorIfExists, opts).Interface().(IfExists_)
	if ifExists != ErrorIfExists && ifExists != DropIfExists {
		query += "if not exists "
	}

	query = query + table + "( "
	for i := range lr.Names {
		if i != 0 {
			query += ", "
		}
		colType, colName, isPK := describe(i)
		query = query + colName + " " + colType
		if isPK {
			pk = append(pk, colName)
		}
	}

	if len(pk) > 0 {
		query = query + ", primary key (" + strings.Join(pk, ",") + ")"
	}

	query += " )"
	return query
}

func sqlDropQuery(table string, opts ...interface{}) string {
	schema := fu.StrOption(Schema(""), opts)
	if schema != "" {
		schema = schema + "."
	}
	return "drop table if exists " + schema + table
}

func sqlTypeOf(tp reflect.Type, driver string) string {
	switch tp.Kind() {
	case reflect.String:
		if driver == "postgres" {
			return "VARCHAR(65535)" /* redshift TEXT == VARCHAR(256) */
		}
		return "TEXT"
	case reflect.Int8, reflect.Uint8, reflect.Int16:
		return "SMALLINT"
	case reflect.Uint16, reflect.Int32, reflect.Int:
		return "INTEGER"
	case reflect.Uint, reflect.Uint32, reflect.Int64, reflect.Uint64:
		return "BIGINT"
	case reflect.Float32:
		if driver == "postgres" {
			return "REAL" /* redshift does not FLOAT */
		}
		return "FLOAT"
	case reflect.Float64:
		if driver == "postgres" {
			return "DOUBLE PRECISION" /* redshift does not have DOUBLE */
		}
		return "DOUBLE"
	case reflect.Bool:
		return "BOOLEAN"
	default:
		if tp == internal.TsType {
			return "DATETIME"
		}
	}
	panic("unsupported data type " + fmt.Sprintf("%v %v", tp.String(), tp.Kind()))
}
