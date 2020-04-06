package csv

import (
	"encoding/csv"
	"github.com/sudachen/go-iokit/iokit"
	"github.com/sudachen/go-ml/fu"
	"github.com/sudachen/go-ml/fu/lazy"
	"github.com/sudachen/go-ml/tables"
	"github.com/sudachen/go-zorros/zorros"
	"io"
	"reflect"
)

type Comma rune

const initialCapacity = 101

/*
	// detects compression automatically
    // can be gzip, bzip2, xz/lzma2
	csv.Read(fu.Compressed("file.csv.xz"),
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

	csv.Read(fu.External("http://sudachen.xyz/testfile.xz",fu.Streamed))

	var csvContent =
    `s1,f_*,f_1,f_2
  	"the first",100,0,0.1
	"another one",200,3,0.2`

	csv.Read(fu.StringIO(csvContent),
                csv.TzeInt("f_**").As("Number"), // hide f_* for next rules
				csv.Float64("f_*").As("Feature*"),
				csv.String("s*").As("Text*"))
*/

func Read(source interface{}, opts ...interface{}) (t *tables.Table, err error) {
	return Source(source, opts...).Collect()
}

func Source(source interface{}, opts ...interface{}) tables.Lazy {
	if e, ok := source.(iokit.Input); ok {
		return lazyread(e, opts...)
	} else if e, ok := source.(string); ok {
		return lazyread(iokit.File(e), opts...)
	} else if rd, ok := source.(io.Reader); ok {
		return lazyread(iokit.Reader(rd, nil), opts...)
	}
	return tables.SourceError(zorros.Errorf("csv reader does not know source type %v", reflect.TypeOf(source).String()))
}

func lazyread(source iokit.Input, opts ...interface{}) tables.Lazy {
	return func() lazy.Stream {
		rd, err := source.Open()
		if err != nil {
			return lazy.Error(err)
		}
		//dq := fu.Decompress(rd)
		cls := io.Closer(rd) //fu.CloserChain{dq, rd}
		rdr := csv.NewReader(rd)
		rdr.Comma = fu.RuneOption(Comma(rdr.Comma), opts)
		vals, err := rdr.Read()
		if err != nil {
			cls.Close()
			return lazy.Error(err)
		}
		fm, names, err := mapFields(vals, opts)
		if err != nil {
			cls.Close()
			return lazy.Error(err)
		}

		rdr.FieldsPerRecord = len(vals)

		type line struct {
			vals []string
			err  error
		}
		nC := make(chan line)
		stopC := make(chan struct{})
		width := len(names)

		go func() {
			defer close(nC)
			for {
				v, e := rdr.Read()
				select {
				case nC <- line{v, e}:
				case <-stopC:
					cls.Close()
					return
				}
			}
		}()

		wc := fu.WaitCounter{Value: 0}
		return func(index uint64) (reflect.Value, error) {
			if index == lazy.STOP {
				wc.Stop()
				close(stopC)
				return reflect.ValueOf(false), nil
			}
			if !wc.Wait(index) {
				return reflect.ValueOf(false), nil
			}
			l, ok := <-nC
			wc.Inc()
			x := reflect.Value{}
			if ok {
				if err = l.err; err != nil {
					if l.err == io.EOF {
						ok = false
						err = nil
					}
				} else {
					output := fu.Struct{names, make([]reflect.Value, width), fu.Bits{}}
					for i, v := range l.vals {
						var na bool
						if na, err = fm[i].Convert(v, &output.Columns[fm[i].field], fm[i].index, fm[i].width); err != nil {
							break
						}
						if na {
							output.Na.Set(fm[i].field, true)
						}
					}
					if err == nil {
						x = reflect.ValueOf(output)
					}
				}
			}
			if !ok || err != nil {
				wc.Stop()
				return reflect.ValueOf(false), err
			}
			return x, nil
		}
	}
}

/*
	csv.Write(t,"file.csv.xz",
				csv.Column("feature_1").Round(2).As("Feature1"))

	csv.Write(t,fu.Lzma2("file.csv.xz"),
				csv.Column("feature*").As("Feature*"))

	bf := bytes.Buffer{}
	csv.Write(t,fu.Gzip(&bf),
				csv.Comma('|'),
				csv.Column("feature*").Round(3).As("Feature*"))

	csv.Write(t,fu.Lzma2(fu.External("gc://$/testfile.csv.xz")),
				csv.Comma('|'),
				csv.Column("feature_1").As("Feature1"))
*/
func Write(t *tables.Table, dest iokit.Output, opts ...interface{}) (err error) {
	return t.Lazy().Drain(Sink(dest, opts...))
}

func Sink(dest iokit.Output, opts ...interface{}) tables.Sink {
	var err error
	f := iokit.Whole(nil)
	if f, err = dest.Create(); err != nil {
		return tables.SinkError(err)
	}
	cwr := csv.NewWriter(f)
	hasHeader := false
	fm := []mapper{}
	names := []string{}
	return func(v reflect.Value) (err error) {
		if v.Kind() == reflect.Bool {
			cwr.Flush()
			if v.Bool() {
				err = f.Commit()
			}
			f.End()
			return
		}
		lr := v.Interface().(fu.Struct)
		if !hasHeader {
			if fm, names, err = mapFields(lr.Names, opts); err != nil {
				return
			}
			if err = cwr.Write(names); err != nil {
				return
			}
			hasHeader = true
		}
		r := make([]string, len(lr.Names))
		for i, x := range lr.Columns {
			r[i] = fm[i].Format(x, lr.Na.Bit(i))
		}
		err = cwr.Write(r)
		return
	}
}
