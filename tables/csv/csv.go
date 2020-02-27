package csv

import (
	"encoding/csv"
	"github.com/sudachen/go-foo/fu"
	"github.com/sudachen/go-ml/internal"
	"github.com/sudachen/go-ml/tables"
	"golang.org/x/xerrors"
	"io"
	"os"
	"reflect"
)

type Column string
type RenamedColumn struct{ CsvCol, TableCol string }
type String string
type RenamedString RenamedColumn
type Int string
type RenamedInt RenamedColumn
type Float32 string
type RenamedFloat32 RenamedColumn
type Float64 string
type RenamedFloat64 RenamedColumn
type Time string // RFC 3339
type TimeLayout struct {
	Col    string
	Layout string
}
type RenamedTime RenamedColumn
type RenamedTimeLayout struct{ CsvCol, TableCol, Layout string }
type Comma rune

const initialCapacity = 101

/*
	// detects compression automatically
    // can be gzip, bzip2, xz/lzma2
	csv.Read("file.csv.xz",
				csv.Float64("feature_1").As("Feature1"),
				csv.Time("feature_2").Like(time.RFC3339Nano).As("Feature2"))

	// will be downloaded every time
	csv.Read(fu.External("s3://profile@bucket/testfile.csv.xz"))

	// will be downloaded only once
	csv.Read(fu.External("http://sudachen.xyz/testfile.xz",
				fu.Cached("external-files/sudachen.xyz/testfile.xz")))

	// loads file from the Zip archive
	csv.Read(fu.ZipFile("dataset1.csv","file.zip"))

	csv.Read(fu.ZipFile("dataset1.csv"
				fu.External("http://sudachen.xyz/testfile.zip",
					fu.Cached("external-files/sudachen.xyz/testfile.zip")))

	var csvContent =
    `s1,f_*,f_1,f_2
  	"the first",100,0,0.1
	"another one",200,3,0.2`

	csv.Read(fu.StringIO(csvContent),
                csv.Int("f_**").As("Number"), // hide f_* for next rules
				csv.Float64("f_*").As("Feature*"),
				csv.String("s*").As("Text*"))
*/

func Read(source interface{}, opts ...interface{}) (t *tables.Table, err error) {
	var f io.ReadCloser
	if e, ok := source.(fu.Input); ok {
		if f, err = e.Open(); err != nil { return }
		defer f.Close()
		return dqread(f, opts...)
	}
	if e, ok := source.(string); ok {
		f, err = os.Open(e)
		defer f.Close()
		return dqread(f, opts...)
	}
	if rd, ok := source.(io.Reader); ok {
		return dqread(rd, opts...)
	}
	return nil, xerrors.Errorf("csv reader does not know source type %v", reflect.TypeOf(source).String())
}

func dqread(source io.Reader, opts ...interface{}) (t *tables.Table, err error) {
	dq := fu.Decompress(source)
	defer dq.Close()
	return read(dq, opts...)
}

func read(source io.Reader, opts ...interface{}) (t *tables.Table, err error) {
	rdr := csv.NewReader(source)
	rdr.Comma = fu.RuneOption(Comma(rdr.Comma), opts)
	var vals []string
	if vals, err = rdr.Read(); err != nil {
		return
	}
	fm, names, err := mapFields(vals, opts)
	if err != nil {
		return
	}
	columns := make([]reflect.Value, len(names))
	na := make([]internal.Bits, len(names))
	rdr.FieldsPerRecord = len(names)
	for i := range columns {
		columns[i] = reflect.MakeSlice(reflect.TypeOf(fm[i].Type), 0, initialCapacity)
	}

	stopC := make(chan []string)
	csvC := make(chan []string)
	go func() {
		for {
			var vx []string
			vx, err = rdr.Read() // function err
			if err != nil {
				if err == io.EOF {
					err = nil
				}
				close(csvC)
				return
			}
			select {
			case csvC <- vx:
			case <-stopC:
				return
			}
		}
	}()

	length := 0
	for {
		vals, ok := <-csvC
		if !ok {
			break
		}
		for j, v := range vals {
			x, xna, e := fm[j].Convert(v)
			if e != nil {
				close(stopC)
				return nil, e
			}
			columns[j] = reflect.Append(columns[j], x)
			na[j].Set(length, xna)
		}
		length++
	}
	if err != nil {
		return
	}
	for i,m := range fm {
		m.AutoConvert(&columns[i],&na[i])
	}
	return tables.MakeTable(names, columns, na, length), nil
}

/*
	csv.Write(t,"file.csv.xz",
				csv.Column("feature_1").As("Feature1"))

	csv.Write(t,fu.Lzma2("file.csv.xz"),
				csv.Column("feature_1").As("Feature1"))

	bf := bytes.Buffer{}
	csv.Write(t,fu.Gzip(&bf),
				csv.Comma('|'),
				csv.Column("feature_1").As("Feature1"))

	csv.Write(t,fu.S3upload("s3://profile@bucket/testfile.csv.xz",
					fu.Lzma2("bucket/myfile.txt.xz")),
				csv.Comma('|'),
				csv.Column("feature_1").As("Feature1"))
*/
func Write(t *tables.Table, dest interface{}, opts ...interface{}) (err error) {
	return
}
