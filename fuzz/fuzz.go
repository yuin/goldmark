package fuzz

import (
	"bytes"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

func Fuzz(data []byte) int {
	markdown := goldmark.New(
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
		goldmark.WithExtensions(
			extension.DefinitionList,
			extension.Footnote,
			extension.GFM,
			extension.Typographer,
		),
	)
	var b bytes.Buffer
	if err := markdown.Convert(data, &b); err != nil {
		return 0
	}

	return 1
}
