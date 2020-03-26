package notes

import (
	"github.com/sudachen/go-foo/fu"
	"github.com/sudachen/go-ml/tables"
	"io"
)

const MaxRowsDefault = 30
const MaxColumnsDefault = 30

type Page struct {
	Title      string
	Header     string
	Footer     string
	MaxRows    int /* MaxRowsDefault if not specified */
	MaxColumns int /* MaxColumnsDefault if not specified */

	cells []Cell
}

type Cell interface {
	// Render as HTML section
	RenderHTML(io.Writer) error
}

/*
SaveAs saves notes page as HTML file

	pg.SaveAs(fu.Lzma2(fu.External("s3://$/reports/last.html.xz")))
    pg.SaveAs(fu.ZipFile("report.html",fu.External("gc://$/reports/last.zip")))
*/
func (pg *Page) SaveAs(output fu.Output) (err error) {
	return
}

/*
Show opens html notes page in the default browser
*/
func (pg *Page) Show(tempfilePattern ...string) {
}

/*
Display renders any if can
*/
func (pg *Page) Display(title string, a interface{}) {
}

func (pg *Page) Plot(title string, a tables.AnyData, charts ...Chart) {
}

func (pg *Page) Head(title string, a tables.AnyData, n int) {
}

func (pg *Page) Tail(title string, a tables.AnyData, n int) {
}

func (pg *Page) Info(title string, a tables.AnyData) {
}

func (pg *Page) Markdown(text string) {
}

func (pg *Page) Markdownf(format string, a ...interface{}) {
}
