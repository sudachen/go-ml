package nn

import (
	"bufio"
	"encoding/binary"
	"github.com/sudachen/go-iokit/iokit"
	"github.com/sudachen/go-ml/fu"
	"github.com/sudachen/go-ml/nn/mx"
	"github.com/sudachen/go-zorros/zorros"
	"golang.org/x/xerrors"
	"io"
	"math"
)

func nf(p func(string) bool, f func(string) bool) func(string) bool {
	return func(s string) bool {
		if p(s) {
			return true
		}
		return f(s)
	}
}
func (network *Network) SaveParams(output iokit.Output, only ...string) (err error) {
	patt := func(string) bool { return true }
	if len(only) > 0 {
		patt = func(string) bool { return false }
		for _, o := range only {
			patt = nf(fu.Pattern(o), patt)
		}
	}
	var wr iokit.Whole
	if wr, err = output.Create(); err != nil {
		return zorros.Trace(err)
	}
	defer wr.End()
	params := fu.SortedKeysOf(network.Params).([]string)
	b := []byte{0, 0, 0, 0}
	dil := []byte{0xa, '-', '-', 0xa}
	magic := []byte{'A', 'N', 'N', '1'}
	order := binary.ByteOrder(binary.LittleEndian)
	if _, err = wr.Write(magic); err != nil {
		return zorros.Trace(err)
	}
	count := 0
	for _, n := range params {
		if n[0] != '_' {
			count++
		}
	}
	order.PutUint32(b, uint32(count))
	if _, err = wr.Write(b); err != nil {
		return zorros.Trace(err)
	}
	if _, err = wr.Write(dil); err != nil {
		return zorros.Trace(err)
	}
	for _, n := range params {
		if !patt(n) {
			continue
		}
		d := network.Params[n]
		if err = binary.Write(wr, order, int32(len(n))); err != nil {
			return zorros.Trace(err)
		}
		if err = binary.Write(wr, order, []byte(n)); err != nil {
			return zorros.Trace(err)
		}
		dim := d.Dim()
		order.PutUint32(b, uint32(dim.Len))
		if _, err = wr.Write(b); err != nil {
			return zorros.Trace(err)
		}
		for i := 0; i < dim.Len; i++ {
			order.PutUint32(b, uint32(dim.Shape[i]))
			if _, err = wr.Write(b); err != nil {
				return zorros.Trace(err)
			}
		}
		total := dim.Total()
		order.PutUint32(b, uint32(total))
		if _, err = wr.Write(b); err != nil {
			return zorros.Trace(err)
		}
		v := d.ValuesF32()
		for i := 0; i < total; i++ {
			order.PutUint32(b, math.Float32bits(v[i]))
			if _, err = wr.Write(b); err != nil {
				return zorros.Trace(err)
			}
		}
		if _, err = wr.Write(dil); err != nil {
			return zorros.Trace(err)
		}
	}
	return wr.Commit()
}

func (network *Network) LoadParams(input iokit.Input, forced ...bool) (err error) {
	var rd io.ReadCloser
	if rd, err = input.Open(); err != nil {
		return zorros.Trace(err)
	}
	defer rd.Close()
	r := bufio.NewReader(rd)
	b := []byte{0, 0, 0, 0}
	equal4b := func(a []byte) bool { return a[0] == b[0] && a[1] == b[1] && a[2] == b[2] && a[3] == b[3] }
	dil := []byte{0xa, '-', '-', 0xa}
	magic := []byte{'A', 'N', 'N', '1'}
	order := binary.ByteOrder(binary.LittleEndian)
	if _, err = io.ReadFull(r, b); err != nil {
		return zorros.Trace(err)
	}
	if !equal4b(magic) {
		return xerrors.Errorf("bad magic")
	}
	if _, err = io.ReadFull(r, b); err != nil {
		return zorros.Trace(err)
	}
	count := int(order.Uint32(b))
	if _, err = io.ReadFull(r, b); err != nil {
		return zorros.Trace(err)
	}
	if !equal4b(dil) {
		return xerrors.Errorf("bad delimiter")
	}
	ready := map[string]bool{}
	v := []float32{}
	for j := 0; j < count; j++ {
		var ln int32
		if err = binary.Read(r, order, &ln); err != nil {
			return zorros.Trace(err)
		}
		ns := make([]byte, ln)
		if err = binary.Read(r, order, &ns); err != nil {
			return zorros.Trace(err)
		}
		n := string(ns)
		d, ok := network.Params[n]
		if !ok && fu.Fnzb(forced...) {
			return xerrors.Errorf("layer '%v' is not exists in the network", n)
		}
		dim := mx.Dimension{}
		if _, err = io.ReadFull(r, b); err != nil {
			return zorros.Trace(err)
		}
		dim.Len = int(order.Uint32(b))
		if dim.Len > mx.MaxDimensionCount {
			return xerrors.Errorf("bad deimension of '%v' layer params", n)
		}
		for i := 0; i < dim.Len; i++ {
			if _, err = io.ReadFull(r, b); err != nil {
				return zorros.Trace(err)
			}
			dim.Shape[i] = int(order.Uint32(b))
		}
		if _, err = io.ReadFull(r, b); err != nil {
			return zorros.Trace(err)
		}
		total := int(order.Uint32(b))
		if total != dim.Total() {
			return xerrors.Errorf("bad deimension of '%v' layer params or values total count is incorrect", n)
		}
		if ok {
			v = make([]float32, total)
		}
		for i := range v {
			if _, err = io.ReadFull(r, b); err != nil {
				return zorros.Trace(err)
			}
			if ok {
				v[i] = math.Float32frombits(order.Uint32(b))
			}
		}
		if _, err = io.ReadFull(r, b); err != nil {
			return zorros.Trace(err)
		}
		if !equal4b(dil) {
			return xerrors.Errorf("bad delimiter")
		}
		if ok {
			d.SetValues(v)
			ready[n] = true
		}
	}
	for k := range network.Params {
		if !ready[k] {
			if k[0] != '_' && fu.Fnzb(forced...) {
				return xerrors.Errorf("layer '%v' does not exists in params file", k)
			}
		}
	}
	network.Initialized = true
	return nil
}
