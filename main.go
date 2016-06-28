package main

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/google/go-github/github"
	"github.com/russross/blackfriday"
)

var links []Link

type Link struct {
	Name     string
	Location string
}

type LinkRenderer struct{}

func NewLinkRenderer(flags int) blackfriday.Renderer {
	return &LinkRenderer{}
}

func (l *LinkRenderer) TitleBlock(out *bytes.Buffer, text []byte)                             {}
func (l *LinkRenderer) BlockCode(out *bytes.Buffer, text []byte, lang string)                 {}
func (l *LinkRenderer) BlockQuote(out *bytes.Buffer, text []byte)                             {}
func (l *LinkRenderer) BlockHtml(out *bytes.Buffer, text []byte)                              {}
func (l *LinkRenderer) Header(out *bytes.Buffer, text func() bool, level int, id string)      {}
func (l *LinkRenderer) HRule(out *bytes.Buffer)                                               {}
func (l *LinkRenderer) ListItem(out *bytes.Buffer, text []byte, flags int)                    {}
func (l *LinkRenderer) Paragraph(out *bytes.Buffer, text func() bool)                         {}
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

func (l *LinkRenderer) GetFlags() int {
	return 0
}

func (l *LinkRenderer) List(out *bytes.Buffer, text func() bool, flags int) {
	if text() {
		out.Write([]byte(""))
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

func SkipDir(name string) bool {
	prefixes := []string{
		".",
		"_",
	}

	for _, skip := range prefixes {
		if strings.HasPrefix(name, skip) {
			return true
		}
	}

	return false
}

func RandomInt(maxInt int) (int64, error) {
	max := big.NewInt(int64(maxInt))
	random, err := rand.Int(rand.Reader, max)
	if err != nil {
		return 0, err
	}

	return random.Int64(), nil
}

func RandomReadme(dir string) (string, error) {
	pwl := "papers-we-love"

	if SkipDir(dir) {
		return RandomReadme("/")
	}

	client := github.NewClient(nil)
	fc, dc, resp, err := client.Repositories.GetContents(pwl, pwl, dir, nil)
	if err != nil {
		return "", err
	}
	fmt.Println(resp)

	if fc == nil {
		randInt, err := RandomInt(len(dc))
		if err != nil {
			return "", err
		}
		randDir := dc[randInt]
		randDirName := randDir.Name

		readmePath := strings.Join([]string{*randDirName, "README.md"}, "/")
		return RandomReadme(readmePath)
	}

	readme, err := fc.GetContent()
	if err != nil {
		return "", err
	}

	return readme, nil
}

func main() {
	readme, err := RandomReadme("/")
	if err != nil {
		log.Println(err)
	}

	fmt.Println(readme)
	fmt.Println("---")

	l := NewLinkRenderer(0)
	extensions := 0
	_ = blackfriday.Markdown([]byte(readme), l, extensions)

	for _, link := range links {
		fmt.Println(link.Name)
		fmt.Println("\t", link.Location)
	}
}
