package fuzz

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/yuin/goldmark/v2/extension"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer/html"
	"github.com/yuin/goldmark/v2/text"
	"github.com/yuin/goldmark/v2/util"
)

func fuzz(f *testing.F) {
	f.Fuzz(func(_ *testing.T, orig string) {
		p := parser.New(
			parser.WithAutoHeadingID(),
			parser.WithAttribute(),
			parser.WithExtensions(
				extension.NewDefinitionListParser(),
				extension.NewFootnoteParser(),
				extension.NewGFMParser(),
				extension.NewTypographerParser(),
				extension.NewLinkifyParser(),
				extension.NewTableParser(),
				extension.NewTaskCheckBoxParser(),
			),
		)
		r := html.New(
			html.WithUnsafe(),
			html.WithXHTML(),
			html.WithExtensions(
				extension.NewDefinitionListHTMLRenderer(),
				extension.NewFootnoteHTMLRenderer(),
				extension.NewGFMHTMLRenderer(),
				extension.NewTableHTMLRenderer(),
				extension.NewTaskListItemHTMLRenderer(),
			),
		)
		src := util.StringToReadOnlyBytes(orig)
		var b bytes.Buffer
		if err := r.Render(&b, src, p.Parse(text.NewReader(src))); err != nil {
			panic(err)
		}
	})
}

func FuzzDefault(f *testing.F) {
	bs, err := os.ReadFile("../_test/spec.json")
	if err != nil {
		panic(err)
	}
	var testCases []map[string]any
	if err := json.Unmarshal(bs, &testCases); err != nil {
		panic(err)
	}
	for _, c := range testCases {
		f.Add(c["markdown"])
	}
	fuzz(f)
}
