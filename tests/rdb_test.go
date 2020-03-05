package tests

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sudachen/go-ml/tables/rdb"
	"gotest.tools/assert"
	"os"
	"testing"
)

func init() {
	_ = os.Remove("/tmp/go-tables-test.sqlite3")
}

func Test_SQL1(t *testing.T) {
	var err error
	q := TrTable()
	url := "sqlite3:file:/tmp/go-tables-test.sqlite3"
	err = rdb.Write(url, q, rdb.Table("q"), rdb.DropIfExists)
	assert.NilError(t, err)
	x, err := rdb.Read(url, rdb.Query("select * from q"))
	assert.NilError(t, err)
	assertTrData(t, x)
}

func Test_SQL1b(t *testing.T) {
	var err error
	q := TrTable()
	url := "sqlite3:file:/tmp/go-tables-test.sqlite3"
	err = rdb.Write(url, q, rdb.Table("q"), rdb.DropIfExists)
	assert.NilError(t, err)
	x, err := rdb.Read(url, rdb.Query("select * from q"))
	assert.NilError(t, err)
	assertTrData(t, x)
}

func Test_SQL2(t *testing.T) {
	var err error
	q := TrTable()

	db, err := sql.Open("sqlite3", "file:/tmp/go-tables-test.sqlite3")
	assert.NilError(t, err)
	err = rdb.Write(db, q, rdb.Table("q"), rdb.DropIfExists)
	assert.NilError(t, err)
	x, err := rdb.Read(db, rdb.Query("select * from q"))
	assert.NilError(t, err)

	assertTrData(t, x)
	err = db.Close()
	assert.NilError(t, err)
}

func Test_SQL3(t *testing.T) {
	var err error
	q := TrTable()
	url := "sqlite3:file:/tmp/go-tables-test.sqlite3"

	err = rdb.Write(
		url, q, rdb.Table("q2"),
		rdb.DropIfExists,
		rdb.VARCHAR("Name", 256).As("id").PrimaryKey(),
		rdb.DECIMAL("Rate", 3, 1).As("ra"),
		rdb.FLOAT("Age").As("aa"))

	assert.NilError(t, err)

	x, err := rdb.Read(url, rdb.Query("select * from q2"),
		rdb.Column("id").As("Name"),
		rdb.FLOAT("ra").As("Rate"),
		rdb.INTEGER("aa").As("Age"))

	assert.NilError(t, err)
	assertTrData(t, x)

	y, err := rdb.Read(url, rdb.Table("q2"),
		rdb.FLOAT("ra"),
		rdb.INTEGER("aa"))
	assert.NilError(t, err)

	for i := 0; i < q.Len(); i++ {
		assert.Assert(t, y.Col("aa").Index(i).Int() == q.Col("Age").Index(i).Int())
	}
}
