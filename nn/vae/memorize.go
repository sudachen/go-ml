package vae

import (
	"github.com/sudachen/go-iokit/iokit"
	"github.com/sudachen/go-ml/fu"
	"github.com/sudachen/go-ml/model"
	"github.com/sudachen/go-ml/nn"
	"github.com/sudachen/go-ml/nn/mx"
	"gopkg.in/yaml.v3"
	"io"
)

type mnemosyne struct {
	network  *nn.Network
	symbol   *mx.Symbol
	features []string
	predicts string
	inputdim mx.Dimension
	params   string
}

func (mm mnemosyne) Memorize(c *model.CollectionWriter) (err error) {
	if err = c.Add("info.yaml", func(wr io.Writer) error {
		en := yaml.NewEncoder(wr)
		return en.Encode(map[string]interface{}{
			"features": mm.features,
			"predicts": mm.predicts,
		})
	}); err != nil {
		return
	}
	if err = c.AddLzma2("params.bin.xz", func(wr io.Writer) (e error) {
		return mm.network.SaveParams(iokit.Writer(wr), mm.params)
	}); err != nil {
		return
	}
	if err = c.AddLzma2("symbol.yaml.xz", func(wr io.Writer) (e error) {
		return nn.SaveSymbol(mm.inputdim, mm.symbol, iokit.Writer(wr))
	}); err != nil {
		return
	}
	return
}

func (e Model) modelmap(network *nn.Network, features []string) model.MemorizeMap {
	return model.MemorizeMap{
		EncoderCollection: mnemosyne{
			network,
			e.encoder(),
			features,
			fu.Fnzs(e.Feature, LatentCol),
			mx.Dim(e.Width),
			"encoder_*",
		},
		DecoderCollection: mnemosyne{
			network,
			e.decoder(mx.Input()),
			[]string{fu.Fnzs(e.Feature, LatentCol)},
			fu.Fnzs(e.Predicted, model.PredictedCol),
			mx.Dim(e.Latent * 2),
			"decoder_*",
		},
		RecoderCollection: mnemosyne{
			network,
			e.recoder(mx.Input()),
			features,
			fu.Fnzs(e.Predicted, model.PredictedCol),
			mx.Dim(e.Width),
			"*coder_*",
		},
	}
}
