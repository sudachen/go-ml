package nn

import (
	"github.com/sudachen/go-iokit/iokit"
	"github.com/sudachen/go-ml/nn/mx"
	"gopkg.in/yaml.v3"
	"io"
	"sync"
)

var symbolMu = sync.Mutex{}
var _symbolId = 0

func NextSymbolId() int {
	_symbolId++
	return _symbolId
}

func resetSymbolId(first int) {
	_symbolId = first
}

func (network *Network) SaveSymbol(output iokit.Output) (err error) {
	var wr iokit.Whole
	if wr, err = output.Create(); err != nil {
		return
	}
	defer wr.End()
	enc := yaml.NewEncoder(wr)
	x := struct {
		Input    mx.Dimension `yaml:"input"`
		Symbolic *mx.Symbol   `yaml:"symbolic"`
	}{network.inputdim, network.symbolic}
	if err = enc.Encode(x); err != nil {
		return
	}
	if err = enc.Close(); err != nil {
		return
	}
	return wr.Commit()
}

func LoadSymbol(input iokit.Input) (symbolic *mx.Symbol, inputdim mx.Dimension, err error) {
	var rd io.ReadCloser
	if rd, err = input.Open(); err != nil {
		return
	}
	defer rd.Close()
	dec := yaml.NewDecoder(rd)
	x := struct {
		Input    mx.Dimension `yaml:"input"`
		Symbolic *mx.Symbol   `yaml:"symbolic"`
	}{}
	if err = dec.Decode(&x); err != nil {
		return
	}
	return x.Symbolic, x.Input, nil
}
