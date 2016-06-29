package mdlinks

import (
	"bytes"

	"github.com/russross/blackfriday"
)

type Link struct {
	Name     string
	Location string
}

var links []Link

type LinkRenderer struct{}

func NewLinkRenderer(flags int) blackfriday.Renderer {
	return &LinkRenderer{}
}

func (l *LinkRenderer) GetFlags() int {
	return 0
}

func (l *LinkRenderer) Paragraph(out *bytes.Buffer, text func() bool) {
	if text() {
		out.WriteString("")
	}
}

func (l *LinkRenderer) List(out *bytes.Buffer, text func() bool, flags int) {
	if text() {
		out.WriteString("")
	}
}

func (l *LinkRenderer) Link(out *bytes.Buffer, link []byte, title []byte, content []byte) {
	newLink := Link{
		Name:     string(content),
		Location: string(link),
	}
	links = append(links, newLink)
}

func (l *LinkRenderer) NormalText(out *bytes.Buffer, text []byte) {
	out.Write(text)
}

// Unused renderers.
func (l *LinkRenderer) TitleBlock(out *bytes.Buffer, text []byte)                             {}
func (l *LinkRenderer) BlockCode(out *bytes.Buffer, text []byte, lang string)                 {}
func (l *LinkRenderer) BlockQuote(out *bytes.Buffer, text []byte)                             {}
func (l *LinkRenderer) BlockHtml(out *bytes.Buffer, text []byte)                              {}
func (l *LinkRenderer) Header(out *bytes.Buffer, text func() bool, level int, id string)      {}
func (l *LinkRenderer) HRule(out *bytes.Buffer)                                               {}
func (l *LinkRenderer) ListItem(out *bytes.Buffer, text []byte, flags int)                    {}
func (l *LinkRenderer) Table(out *bytes.Buffer, header []byte, body []byte, columnData []int) {}
func (l *LinkRenderer) TableRow(out *bytes.Buffer, text []byte)                               {}
func (l *LinkRenderer) TableHeaderCell(out *bytes.Buffer, text []byte, align int)             {}
func (l *LinkRenderer) TableCell(out *bytes.Buffer, text []byte, align int)                   {}
func (l *LinkRenderer) Footnotes(out *bytes.Buffer, text func() bool)                         {}
func (l *LinkRenderer) FootnoteItem(out *bytes.Buffer, name, text []byte, flags int)          {}
func (l *LinkRenderer) AutoLink(out *bytes.Buffer, link []byte, kind int)                     {}
func (l *LinkRenderer) CodeSpan(out *bytes.Buffer, text []byte)                               {}
func (l *LinkRenderer) DoubleEmphasis(out *bytes.Buffer, text []byte)                         {}
func (l *LinkRenderer) Emphasis(out *bytes.Buffer, text []byte)                               {}
func (l *LinkRenderer) Image(out *bytes.Buffer, link []byte, title []byte, alt []byte)        {}
func (l *LinkRenderer) LineBreak(out *bytes.Buffer)                                           {}
func (l *LinkRenderer) RawHtmlTag(out *bytes.Buffer, tag []byte)                              {}
func (l *LinkRenderer) TripleEmphasis(out *bytes.Buffer, text []byte)                         {}
func (l *LinkRenderer) StrikeThrough(out *bytes.Buffer, text []byte)                          {}
func (l *LinkRenderer) FootnoteRef(out *bytes.Buffer, ref []byte, id int)                     {}
func (l *LinkRenderer) Entity(out *bytes.Buffer, entity []byte)                               {}
func (l *LinkRenderer) DocumentHeader(out *bytes.Buffer)                                      {}
func (l *LinkRenderer) DocumentFooter(out *bytes.Buffer)                                      {}

func Links(markdown []byte) []Link {
	l := NewLinkRenderer(0)
	_ = blackfriday.Markdown(markdown, l, 0)

	return links
}
