package notes

import (
	"github.com/sudachen/go-iokit/iokit"
	"github.com/sudachen/go-ml/tables"
	"github.com/sudachen/go-zorros/zorros"
)

const MaxRowsDefault = 30
const MaxColumnsDefault = 30

type Page struct {
	Title      string
	Header     string
	Footer     string
	MaxRows    int /* MaxRowsDefault if not specified */
	MaxColumns int /* MaxColumnsDefault if not specified */
}

type PageWriter struct {
	w iokit.Whole
}

func (Page) Create(output iokit.Output) (pw *PageWriter, err error) {
	return
}

func (pw *PageWriter) Ensure() {
	// if panic write to page reason of panic
	_ = pw.Commit()
}

func (pw *PageWriter) End() {
}

func (pw *PageWriter) Commit() (err error) {
	return
}

func (pg Page) LuckyCreate(output iokit.Output) *PageWriter {
	pw, err := pg.Create(output)
	if err != nil {
		panic(zorros.Panic(err))
	}
	return pw
}

/*
SaveAs saves notes page as HTML file

	pg.SaveAs(iokit.Lzma2(fu.External("s3://$/reports/last.html.xz")))
    pg.SaveAs(iokit.ZipFile("report.html",iokit.External("gc://$/reports/last.zip")))
*/
func (pg *PageWriter) SaveAs(output iokit.Output) (err error) {
	return
}

/*
Show opens html notes page in the default browser
*/
func (pg *PageWriter) Show(brawser ...string) {
}

/*
Display renders any if can
*/
func (pg *PageWriter) Display(title string, a interface{}) {
}

func (pg *PageWriter) Plot(title string, a tables.AnyData, charts ...Chart) {
}

func (pg *PageWriter) Head(title string, a tables.AnyData, n int) {
}

func (pg *PageWriter) Tail(title string, a tables.AnyData, n int) {
}

func (pg *PageWriter) Info(title string, a tables.AnyData) {
}

func (pg *PageWriter) Markdown(text string) {
}

func (pg *PageWriter) Markdownf(format string, a ...interface{}) {
}
