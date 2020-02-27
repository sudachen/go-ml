package rdb

import (
	"database/sql"
	"fmt"
	"github.com/sudachen/go-foo/fu"
	"github.com/sudachen/go-ml/internal"
	"github.com/sudachen/go-ml/mlutil"
	"github.com/sudachen/go-ml/tables"
	"golang.org/x/xerrors"
	"reflect"
	"strings"
	//	_ "github.com/go-sql-driver/mysql"
	//	_ "github.com/lib/pq"
	//	_ "github.com/mattn/go-sqlite3"
)

func Read(source interface{}, query string, opts ...interface{}) (*tables.Table, error) {
	var db *sql.DB
	var drv string
	var err error

	if url, ok := source.(string); ok {
		var conn string
		drv, conn = splitDriver(url)
		db, err = sql.Open(drv, conn)
		if err != nil {
			return nil, xerrors.Errorf("database connection error: %w", err)
		}
		defer db.Close()
		opts = append(opts, Driver(drv))
	} else if db, ok = source.(*sql.DB); ok {
		drv = fu.StrOption(Driver(""), opts)
	}
	rows, err := db.Query(query)
	if err != nil {
		return nil, xerrors.Errorf("query error: %w", err)
	}
	return sqlSelect(rows, opts...)
}

func sqlSelect(rows *sql.Rows, opts ...interface{}) (*tables.Table, error) {
	length := 0

	tps, err := rows.ColumnTypes()
	if err != nil {
		return nil, xerrors.Errorf("get types error: %w", err)
	}

	ns, err := rows.Columns()
	if err != nil {
		return nil, xerrors.Errorf("get names error: %w", err)
	}

	na := make([]internal.Bits, len(ns))
	columns := make([]reflect.Value, len(ns))

	x := make([]interface{}, len(columns))
	sqltps := GetDbTypes(opts)

	for i := range tps {
		var s SqlScan
		if dtn, ok := sqltps[ns[i]]; ok {
			s = scanner(dtn)
		} else {
			s = scanner(tps[i].DatabaseTypeName())
		}
		x[i] = s
		columns[i] = reflect.MakeSlice(reflect.SliceOf(s.Reflect()), 0, 0)
	}

	for rows.Next() {
		err = rows.Scan(x...)
		if err != nil {
			return nil, xerrors.Errorf("sql row scan error for filed: %w", err)
		}
		for i := range columns {
			y := x[i].(SqlScan)
			v, ok := y.Value()
			if !ok {
				na[i].Set(length, true)
			}
			columns[i] = reflect.Append(columns[i], v)
		}
		length++
	}

	return tables.MakeTable(ns, columns, na, length), nil
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

func Write(source interface{}, t *tables.Table, table string, opts ...interface{}) error {

	var db *sql.DB
	var drv string
	var err error

	if url, ok := source.(string); ok {
		var conn string
		drv, conn = splitDriver(url)
		db, err = sql.Open(drv, conn)
		if err != nil {
			return xerrors.Errorf("database connection error: %w", err)
		}
		defer db.Close()
		opts = append(opts, Driver(drv))
	} else if db, ok = source.(*sql.DB); ok {
		drv = fu.StrOption(Driver(""), opts)
	}

	tx, err := db.Begin()
	if err != nil {
		return xerrors.Errorf("database begin transaction error: %w", err)
	}

	if fu.Option(ErrorIfExists, opts).Interface().(IfExists_) == DropIfExists {
		_, err := tx.Exec(SqlDropQuery(table, opts...))
		if err != nil {
			return xerrors.Errorf("drop table error: %w", err)
		}
	}

	_, err = tx.Exec(SqlCreateQuery(t, table, opts...))
	if err != nil {
		return xerrors.Errorf("create table error: %w", err)
	}

	err = SqlInsert(t, tx, table, opts...)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return xerrors.Errorf("commit transaction error: %w", err)
	}
	return nil
}

func SqlDropQuery(table string, opts ...interface{}) string {
	schema := fu.StrOption(Schema(""), opts)
	if schema != "" {
		schema = schema + "."
	}
	return "drop table if exists " + schema + table
}

func SqlCreateQuery(t *tables.Table, table string, opts ...interface{}) string {
	ifExists := fu.Option(ErrorIfExists, opts).Interface().(IfExists_)
	schema := fu.StrOption(Schema(""), opts)
	pk := fu.StrOption(PrimaryKey(""), opts)

	if schema != "" {
		schema = schema + "."
	}

	query := "create table "

	if ifExists != ErrorIfExists && ifExists != DropIfExists {
		query += "if not exists "
	}

	query = query + schema + table + "( "
	sqltps := GetDbTypes(opts)
	driver := fu.StrOption(Driver(""), opts)

	raw := t.Raw()

	for i, n := range raw.Names {
		if i != 0 {
			query += ", "
		}
		if s, ok := sqltps[n]; ok {
			query = query + n + " " + s
		} else {
			query = query + n + " " +
				SqlTypeOf(raw.Columns[i].Type().Elem(), driver)
		}
	}

	if pk != "" {
		query = query + ", primary key (" + pk + ")"
	}

	query += " )"
	fmt.Println(query)
	return query
}

func SqlTypeOf(tp reflect.Type, driver string) string {
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

func GetDbTypes(opts []interface{}) map[string]string {
	m := map[string]string{}
	driver := fu.StrOption(Driver(""), opts)
	for _, o := range opts {
		switch v := o.(type) {
		case DATE:
			m[string(v)] = "DATE"
		case DATETIME:
			m[string(v)] = "DATETIME"
		case TIMESTAMP:
			if driver == "sqlite3" {
				m[string(v)] = "DATETIME"
			} else {
				m[string(v)] = "TIMESTAMP"
			}
		case BOOLEAN:
			m[string(v)] = "BOOLEAN"
		case SMALLINT:
			m[string(v)] = "SMALLINT"
		case INTEGER:
			m[string(v)] = "INTEGER"
		case BIGINT:
			m[string(v)] = "BIGINT"
		case FLOAT:
			if driver == "postgres" {
				m[string(v)] = "REAL"
			} else {
				m[string(v)] = "FLOAT"
			}
		case DOUBLE:
			if driver == "postgres" {
				m[string(v)] = "DOUBLE PRECISION"
			} else {
				m[string(v)] = "DOUBLE"
			}
		case DECIMAL_:
			s := "DECIMAL"
			if v.Precision >= 0 {
				s += fmt.Sprintf("(%d,%d)", v.Precision, v.Scale)
			}
			m[v.Name] = s
		case VARCHAR_:
			m[v.Name] = fmt.Sprintf("VARCHAR(%d)", v.Length)
		case AUTOINCREMENT:
			if driver == "postgres" {
				m[string(v)] = "SERIAL NOT NULL"
			} else if driver == "mysql" {
				m[string(v)] = "INTEGER NOT NULL AUTO_INCREMENT"
			} else {
				m[string(v)] = "INTEGER NOT NULL AUTOINCREMENT"
			}
		}
	}
	fmt.Println(m)
	return m
}

func SqlInsert(t *tables.Table, tx *sql.Tx, table string, opts ...interface{}) error {
	ifExists := fu.Option(ErrorIfExists, opts).Interface().(IfExists_)
	schema := fu.StrOption(Schema(""), opts)
	batchLen := fu.IntOption(Batch(1), opts)
	drv := fu.StrOption(Driver(""), opts)
	pk := strings.Split(fu.StrOption(PrimaryKey(""), opts), ",")

	if schema != "" {
		schema = schema + "."
	}

	raw := t.Raw()

	batch := make([]interface{}, 0, batchLen*len(raw.Names))
	var stmt *sql.Stmt

	isPk := make([]bool, len(raw.Names))
	if ifExists == InsertUpdateIfExists && len(pk) > 0 {
		for j := range raw.Names {
			if mlutil.IndexOf(raw.Names[j], pk) >= 0 {
				isPk[j] = true
			}
		}
	}

	insertBatch := func() (err error) {
		L := len(raw.Names)
		BL := len(batch) / L
		if stmt == nil || BL != batchLen {
			q1 := " values "
			for j := 0; j < BL; j++ {
				q1 += "("
				if drv == "postgres" {
					for k := range raw.Names {
						q1 += fmt.Sprintf("$%d,", j*L+k+1)
					}
				} else {
					q1 += strings.Repeat("?,", L)
				}
				q1 = q1[:len(q1)-1] + "),"
			}
			q := "insert into " + schema + table + "(" + strings.Join(raw.Names, ",") + ")" + q1[:len(q1)-1]

			if ifExists == InsertUpdateIfExists && len(pk) > 0 {
				q += " on duplicate key update "
				for j, n := range raw.Names {
					if !isPk[j] {
						q += " " + n + " = values(" + n + "),"
					}
				}
				q = q[:len(q)-1]
			}
			stmt, err = tx.Prepare(q)
			if err != nil {
				return
			}
		}
		_, err = stmt.Exec(batch...)
		return
	}

	for i := 0; i < raw.Length; i++ {
		for j, c := range raw.Columns {
			if raw.Na[j].Bit(i) {
				batch = append(batch, nil)
			} else {
				batch = append(batch, c.Index(i).Interface())
			}
		}
		if (i+1)%batchLen == 0 || i+1 >= raw.Length {
			if err := insertBatch(); err != nil {
				return err
			}
			batch = batch[:0]
		}
	}

	return nil
}
