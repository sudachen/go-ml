package mlutil

import (
	"github.com/sudachen/go-fp/fu"
	"golang.org/x/xerrors"
	"io"
)

type CopyProgress func(count int)
type CopyBufferSize int

func Copy(writer io.Writer, reader io.Reader, opts ...interface{}) (count int, err error) {
	cp := fu.IfsOption(CopyProgress(func(int) {}), opts).(CopyProgress)
	size := fu.IntOption(CopyBufferSize(32*1024), opts)
	buf := make([]byte, size)
	for {
		if nr, er := reader.Read(buf); nr > 0 {
			var nw int
			if nw, err = writer.Write(buf[0:nr]); err != nil {
				return
			}
			if nw > 0 {
				count += nw
			}
			if nr != nw {
				err = xerrors.Errorf("short write")
				return
			}
			cp(count)
		} else if er != nil {
			if er != io.EOF {
				err = er
			}
			return
		}
	}
}
