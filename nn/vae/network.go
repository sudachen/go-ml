package vae

import (
	"github.com/sudachen/go-ml/fu"
	"github.com/sudachen/go-ml/nn"
	"github.com/sudachen/go-ml/nn/mx"
)

func (e Model) encoder() *mx.Symbol {
	w := mx.Var("encoder_1_weight")
	b := mx.Var("encoder_1_bias", &nn.Const{0})
	h := mx.FullyConnected(mx.Flatten(mx.Input()), w, b, fu.Fnzi(e.Hidden, e.Width/2), false)
	a := mx.Activation(h, mx.ActivTanh)
	if e.Hidden2 > 0 {
		w = mx.Var("encoder_2_weight")
		b = mx.Var("encoder_2_bias", &nn.Const{0})
		h = mx.FullyConnected(a, w, b, e.Hidden2, false)
		a = mx.Activation(h, mx.ActivTanh)
	}
	w = mx.Var("encoder_x_weight")
	b = mx.Var("encoder_x_bias", &nn.Const{0})
	return mx.FullyConnected(a, w, b, e.Latent*2, false)
}

func (e Model) decoder(in *mx.Symbol) *mx.Symbol {
	w := mx.Var("decoder_1_weight")
	b := mx.Var("decoder_1_bias", &nn.Const{0})
	h := mx.FullyConnected(in, w, b, fu.Fnzi(e.Hidden2, e.Hidden, e.Width/2), false)
	a := mx.Activation(h, mx.ActivTanh)
	if e.Hidden2 > 0 {
		w = mx.Var("decoder_2_weight")
		b = mx.Var("decoder_2_bias", &nn.Const{0})
		h = mx.FullyConnected(a, w, b, fu.Fnzi(e.Hidden, e.Width/2), false)
		a = mx.Activation(h, mx.ActivTanh)
	}
	w = mx.Var("decoder_x_weight")
	b = mx.Var("decoder_x_bias", &nn.Const{0})
	h = mx.FullyConnected(a, w, b, e.Width, false)
	return mx.Activation(h, mx.ActivSigmoid)
}

func (e Model) loss(x *mx.Symbol) *mx.Symbol {
	y := mx.Flatten(mx.Input())
	mu := mx.Ref("mu", x)
	logvar := mx.Ref("logvar", x)
	a := mx.Add(mx.Square(mu), mx.Exp(logvar))
	a = mx.Add(mx.Sub(logvar, a), 1)
	kl := mx.Mul(mx.Sum(a, 1), 0.5)
	a = mx.Add(x, 1e-12)
	b := mx.Mul(y, mx.Log(a))
	c := mx.Mul(mx.Sub(1, y), mx.Log(mx.Sub(1, a)))
	a = mx.Sum(mx.Add(b, c), 1)
	return mx.Mul(mx.Add(kl, a), -1)
}

func (e Model) autoencoder(sym *mx.Symbol) *mx.Symbol {
	h := e.encoder()
	mu := mx.Slice(h, 1, 0, e.Latent)
	mu.SetName("mu")
	logvar := mx.Slice(h, 1, e.Latent, e.Latent*2)
	logvar.SetName("logvar")
	epsilon := mx.Normal(0, 1, e.BatchSize, e.Latent)
	epsilon = mx.BcastMul(epsilon, mx.Var("_sampling", mx.Dim(1)))
	z := mx.Add(mu, mx.Mul(epsilon, mx.Exp(mx.Mul(logvar, 0.5))))
	return e.decoder(z)
}

func (e Model) recoder(sym *mx.Symbol) *mx.Symbol {
	return e.decoder(mx.Slice(e.encoder(), 1, 0, e.Latent))
}
