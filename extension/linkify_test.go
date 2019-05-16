package extension

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"
	"testing"
)

func TestLinkify(t *testing.T) {
	markdown := goldmark.New(
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
		goldmark.WithExtensions(
			Linkify,
		),
	)
	goldmark.DoTestCaseFile(markdown, "_test/linkify.txt", t)
}
