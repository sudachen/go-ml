package notes

import (
	"io"
	//"github.com/gomarkdown/markdown"
)

type MarkdownCell struct {
	text string
}

func (*MarkdownCell) RenderHTML(wr io.Writer) (err error) {
	return
}
