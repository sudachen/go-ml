package ml

import "github.com/sudachen/go-foo/lazy"

/*
Dataset is an abstraction of some source of a data to feed hungry models
*/
type Dataset struct {
	Source   func() lazy.Stream // any stream of mlutil.Struct objects
	Label    string   // name of float32/Tensor field containing label to train
	Test     string   // name of boolean field containing to select test data
	Features []string // patterns of feature names to train or test model
}

