package extension

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"
	"testing"
)

func TestFootnote(t *testing.T) {
	markdown := goldmark.New(
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
		goldmark.WithExtensions(
			Footnote,
		),
	)
	goldmark.DoTestCaseFile(markdown, "_test/footnote.txt", t)
}
