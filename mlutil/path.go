package mlutil

import (
	"github.com/sudachen/go-foo/fu"
	"path/filepath"
)

func ModelPath(s string) string {
	if filepath.IsAbs(s) {
		return s
	}
	return fu.CacheFile(filepath.Join("go-ml", "Models", s))
}
