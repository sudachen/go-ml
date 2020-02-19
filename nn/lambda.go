package nn

import "github.com/sudachen/go-ml/nn/mx"

type Lambda struct {
	F func(*mx.Symbol) *mx.Symbol
}

func (nb *Lambda) Combine(input *mx.Symbol) *mx.Symbol {
	return nb.F(input)
}
