package nn

import (
	"github.com/sudachen/go-iokit/iokit"
	"github.com/sudachen/go-ml/model"
	"gopkg.in/yaml.v3"
	"io"
)

type mnemosyne struct {
	network  *Network
	features []string
	predicts string
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
		return mm.network.SaveParams(iokit.Writer(wr))
	}); err != nil {
		return
	}
	if err = c.AddLzma2("symbol.yaml.xz", func(wr io.Writer) (e error) {
		return mm.network.SaveSymbol(iokit.Writer(wr))
	}); err != nil {
		return
	}
	return
}

