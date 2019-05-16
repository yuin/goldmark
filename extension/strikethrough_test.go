package extension

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"
	"testing"
)

func TestStrikethrough(t *testing.T) {
	markdown := goldmark.New(
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
		goldmark.WithExtensions(
			Strikethrough,
		),
	)
	goldmark.DoTestCaseFile(markdown, "_test/strikethrough.txt", t)
}
