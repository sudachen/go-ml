package csv

import (
	"encoding/csv"
	"github.com/sudachen/go-foo/fu"
	"github.com/sudachen/go-foo/lazy"
	"github.com/sudachen/go-ml/mlutil"
	"github.com/sudachen/go-ml/tables"
	"golang.org/x/xerrors"
	"io"
	"os"
	"reflect"
)

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

	csv.Read(fu.External("http://sudachen.xyz/testfile.xz",fu.Streamed))

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
		if f, err = e.Open(); err != nil {
			return
		}
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
	na := make([]mlutil.Bits, len(names))
	rdr.FieldsPerRecord = len(names)
	for i := range columns {
		columns[i] = reflect.MakeSlice(reflect.SliceOf(fm[i].Type()), 0, initialCapacity)
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
	for i, m := range fm {
		m.AutoConvert(&columns[i], &na[i])
	}
	return tables.MakeTable(names, columns, na, length), nil
}

func Source(source interface{}, opts ...interface{}) tables.Lazy {
	if e, ok := source.(fu.Input); ok {
		return lazyread(e, opts...)
	} else if e, ok := source.(string); ok {
		return lazyread(fu.File(e), opts...)
	} else if rd, ok := source.(io.Reader); ok {
		return lazyread(fu.WrapClose(rd, nil), opts...)
	}
	return tables.SourceError(xerrors.Errorf("csv reader does not know source type %v", reflect.TypeOf(source).String()))
}

func lazyread(source fu.Input, opts ...interface{}) tables.Lazy {
	return func() lazy.Stream {
		rd, err := source.Open()
		if err != nil {
			return lazy.Error(err)
		}
		dq := fu.Decompress(rd)
		cls := fu.CloserChain{dq, rd}
		rdr := csv.NewReader(dq)
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

		rdr.FieldsPerRecord = len(names)
		stopC := make(chan struct{})
		csvC := make(chan reflect.Value)

		go func() {
			type line struct {
				vals []string
				err  error
			}
			width := len(names)
			output := &mlutil.Struct{}
			input := []string{}
			concurrency := 2
			iC := make(chan int, concurrency)
			oC := make(chan struct {
				mlutil.Bits
				error
			}, concurrency)
			nC := make(chan line)
			for i := 0; i < concurrency; i++ {
				go func() {
					for {
						if i, ok := <-iC; ok {
							err := error(nil)
							na := mlutil.Bits{}
							for j := i; j < width; j += concurrency {
								v, b, e := fm[j].Convert(input[j])
								if e != nil {
									err = e
									break
								}
								output.Columns[j] = v
								na.Set(j, b)
							}
							oC <- struct {
								mlutil.Bits
								error
							}{na, err}
						}
					}
				}()
			}
			stopf := make(chan struct{})
			go func() {
				for {
					v, e := rdr.Read()
					select {
					case nC <- line{v, e}:
					case <-stopf:
						return
					}
				}
			}()
		loop:
			for {
				l := <-nC
				x := reflect.Value{}
				if l.err != nil {
					if l.err == io.EOF {
						break loop
					}
					x = reflect.ValueOf(l.err)
				} else {
					output = &mlutil.Struct{names, make([]reflect.Value, width), mlutil.Bits{}}
					input = l.vals
					for i := 0; i < concurrency; i++ {
						iC <- i
					}
					count := 0
					err = nil
					x = reflect.ValueOf(*output)
					for count < concurrency {
						fe := <-oC
						err = fu.Fnze(err, fe.error)
						output.Na.Or_(fe.Bits)
						count++
					}
					if err != nil {
						x = reflect.ValueOf(err)
					}
				}
				select {
				case csvC <- x:
				case <-stopC:
					break loop
				}
			}
			cls.Close()
			close(stopf)
			close(iC)
			close(csvC)
		}()

		wc := lazy.WaitCounter{Value: 0}
		return func(index uint64) (reflect.Value, error) {
			if index == lazy.STOP {
				wc.Stop()
				close(stopC)
				return reflect.ValueOf(false), nil
			}
			if !wc.Wait(index) {
				return reflect.ValueOf(false), nil
			}
			val, ok := <-csvC
			if ok {
				err, _ = val.Interface().(error)
			}
			if !ok || err != nil {
				wc.Stop()
				return reflect.ValueOf(false), err
			}
			wc.Inc()
			return val, nil
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

	csv.Write(t,fu.Lzma2(s3.External("profile:bucket/testfile.csv.xz")),
				csv.Comma('|'),
				csv.Column("feature_1").As("Feature1"))
*/
func Write(t *tables.Table, dest interface{}, opts ...interface{}) (err error) {
	return t.Lazy().Drain(Sink(dest, opts...))
}

func Sink(dest interface{}, opts ...interface{}) tables.Sink {
	var err error
	f := io.Writer(nil)
	cls := io.Closer(nil)
	if e, ok := dest.(fu.Output); ok {
		if f, err = e.Create(); err == nil {
			cls = f.(io.Closer)
		}
	} else if e, ok := dest.(string); ok {
		if f, err = os.Create(e); err == nil {
			cls = f.(io.Closer)
		}
	} else if wr, ok := dest.(io.Writer); ok {
		f = wr
	} else {
		return tables.SinkError(xerrors.Errorf("csv writer does not know dest type %v", reflect.TypeOf(dest).String()))
	}
	if err != nil {
		return tables.SinkError(err)
	}
	cwr := csv.NewWriter(f)
	hasHeader := false
	fm := []mapper{}
	names := []string{}
	return func(v reflect.Value) (err error) {
		if v.Kind() == reflect.Bool {
			cwr.Flush()
			if cls != nil {
				err = cls.Close()
			}
			if !v.Bool() { // shit happens, remove dest
			}
			return
		}
		lr := v.Interface().(mlutil.Struct)
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
