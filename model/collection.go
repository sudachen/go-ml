package model

import (
	"archive/zip"
	"github.com/sudachen/go-iokit/iokit"
	"github.com/sudachen/go-zorros/zorros"
	"github.com/ulikunitz/xz"
	"io"
	"path/filepath"
)

/*
Mnemosyne is a Serialization interface for an ML model parts
*/
type Mnemosyne interface {
	Memorize(*CollectionWriter) error
}

/*
MemorizeMap maps names of models in directory to Mnemosyne abstraction
*/
type MemorizeMap map[string]Mnemosyne

/*
ObjectifyMap mpas names of models in directory to objectification functions
*/
type ObjectifyMap map[string]func(map[string]iokit.Input) (PredictionModel, error)

/*
Memorize writes models directory to single output
*/
func Memorize(output iokit.Output, m MemorizeMap) error {
	if output == nil {
		return nil
	}
	f, err := output.Create()
	if err != nil {
		return zorros.Trace(err)
	}
	defer f.End()
	wz := zip.NewWriter(f)
	for k, w := range m {
		if err = w.Memorize(&CollectionWriter{wz, k}); err != nil {
			return zorros.Trace(err)
		}
	}
	err = wz.Close()
	if err != nil {
		return zorros.Trace(err)
	}
	err = f.Commit()
	if err != nil {
		return zorros.Trace(err)
	}
	return nil
}

/*
CollectionWriter is an abstraction of a collection creator
*/
type CollectionWriter struct {
	wz *zip.Writer
	k  string
}

/*
Add an element to collection
*/
func (c *CollectionWriter) Add(name string, write func(io.Writer) error) error {
	return c.add(name, false, write)
}

/*
Add an Lzma2 compressed element to collection
*/
func (c *CollectionWriter) AddLzma2(name string, write func(io.Writer) error) error {
	return c.add(name, true, write)
}

func (c *CollectionWriter) add(name string, lzma2 bool, write func(io.Writer) error) error {
	fname := c.k + "/" + name
	fh := &zip.FileHeader{Name: fname, Method: zip.Deflate}
	if lzma2 {
		fh.Method = zip.Store
	}
	wr, err := c.wz.CreateHeader(fh)
	if err != nil {
		return zorros.Trace(err)
	}
	if lzma2 {
		xw, err := xz.NewWriter(wr)
		if err != nil {
			return zorros.Trace(err)
		}
		if err = write(xw); err != nil {
			return zorros.Trace(err)
		}
		if err = xw.Close(); err != nil {
			return zorros.Trace(err)
		}
	} else {
		if err = write(wr); err != nil {
			return zorros.Trace(err)
		}
	}
	return nil
}

/*
Objectify reads and reconstructs prediction models from a directory
*/
func Objectify(input iokit.Input, m ObjectifyMap) (pm map[string]PredictionModel, err error) {
	var r *zip.Reader
	f, err := input.Open()
	if err != nil {
		return
	}
	defer f.Close()
	if r, err = zip.NewReader(f.(io.ReaderAt), iokit.FileSize(f)); err != nil {
		return nil, zorros.Trace(err)
	}
	dict := map[string]map[string]iokit.Input{}
	order := []string{}
	for _, j := range r.File {
		dir := filepath.Dir(j.Name)
		if dir != "" && m[dir] != nil {
			d, ok := dict[dir]
			if !ok {
				d = map[string]iokit.Input{}
				dict[dir] = d
				order = append(order, dir)
			}
			if j.Method == zip.Store {
				d[filepath.Base(j.Name)] = iokit.Compressed(iokit.ZipFile(j.Name, input))
			} else {
				d[filepath.Base(j.Name)] = iokit.ZipFile(j.Name, input)
			}
		}
	}
	pm = map[string]PredictionModel{}
	for _, n := range order {
		var v PredictionModel
		f := m[n]
		if v, err = f(dict[n]); err != nil {
			return
		}
		pm[n] = v
	}
	return
}
