package extension_test

import (
	"testing"

	"github.com/yuin/goldmark/v2/extension"
	"github.com/yuin/goldmark/v2/parser"
)

func newExtensionParser() parser.Parser {
	return parser.New(
		parser.WithExtensions(
			extension.NewStrikethroughParser(),
			extension.NewTableParser(),
			extension.NewFootnoteParser(),
			extension.NewDefinitionListParser(),
		),
	)
}

func TestInlinePos(t *testing.T) {
	source := []byte(`~~strike~~

| a | b |
|---|---|
| c | d |

[^fn]: footnote

term
: definition

text [^fn] end
`)
	n := newExtensionParser().Parse(source)
	// Strikethrough starts at the first character of the source
	if n.FirstChild().FirstChild().Pos() != 0 {
		t.Error("unexpected position for strikethrough")
	}
	// FootnoteReference is the 2nd child of the last paragraph
	if n.LastChild().FirstChild().NextSibling().Pos() != 84 {
		t.Error("unexpected position for footnote reference")
	}
}

func TestBlockPos(t *testing.T) {
	source := []byte(`~~strike~~

| a | b |
|---|---|
| c | d |

[^fn]: footnote

term
: definition

text [^fn] end
`)
	n := newExtensionParser().Parse(source)
	table := n.FirstChild().NextSibling()
	if table.Pos() != 12 {
		t.Error("unexpected position for table")
	}
	tableHeader := table.FirstChild()
	if tableHeader.Pos() != 12 {
		t.Error("unexpected position for table header")
	}
	if tableHeader.FirstChild().Pos() != 13 {
		t.Error("unexpected position for 1st table header cell")
	}
	if tableHeader.FirstChild().NextSibling().Pos() != 17 {
		t.Error("unexpected position for 2nd table header cell")
	}
	tableBody := table.LastChild()
	if tableBody.Pos() != 32 {
		t.Error("unexpected position for table body")
	}
	tableRow := tableBody.FirstChild()
	if tableRow.Pos() != 32 {
		t.Error("unexpected position for table row")
	}
	if tableRow.FirstChild().Pos() != 33 {
		t.Error("unexpected position for 1st table body cell")
	}
	if tableRow.FirstChild().NextSibling().Pos() != 37 {
		t.Error("unexpected position for 2nd table body cell")
	}
	footnotedef := table.NextSibling()
	if footnotedef.Pos() != 43 {
		t.Error("unexpected position for footnote definition")
	}
	defList := footnotedef.NextSibling()
	if defList.Pos() != 60 {
		t.Error("unexpected position for definition list")
	}
	if defList.FirstChild().Pos() != 60 {
		t.Error("unexpected position for definition term")
	}
	if defList.LastChild().Pos() != 65 {
		t.Error("unexpected position for definition description")
	}
}
