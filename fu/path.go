package fu

import (
	"github.com/sudachen/go-iokit/iokit"
	"path/filepath"
)

func ModelPath(s string) string {
	if filepath.IsAbs(s) {
		return s
	}
	return iokit.CacheFile(filepath.Join("go-ml", "Models", s))
}
