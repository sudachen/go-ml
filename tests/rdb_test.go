package tests

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sudachen/go-ml/tables/rdb"
	"gotest.tools/assert"
	"testing"
)

func Test_SQL1(t *testing.T) {
	var err error
	q := TrTable()
	url := "sqlite3:file:/tmp/go-tables-test.sqlite3"
	err = rdb.Write(url, q, "q", rdb.DropIfExists)
	assert.NilError(t, err)
	x, err := rdb.Read(url, "select * from q")
	assert.NilError(t, err)
	assertTrData(t, x)
}

func Test_SQL2(t *testing.T) {
	var err error
	q := TrTable()

	db, err := sql.Open("sqlite3", "file:/tmp/go-tables-test.sqlite3")

	assert.NilError(t, err)
	err = rdb.Write(db, q, "q", rdb.DropIfExists)
	assert.NilError(t, err)

	x, err := rdb.Read(db, "select * from q")

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
		url, q, "q2",
		rdb.DropIfExists,
		rdb.VARCHAR("Name", 256),
		rdb.DECIMAL("Rate", 3, 1),
		rdb.FLOAT("Age"))

	assert.NilError(t, err)

	x, err := rdb.Read(url, "select * from q2", rdb.FLOAT("Rate"), rdb.INTEGER("Age"))

	assert.NilError(t, err)
	assertTrData(t, x)
}
