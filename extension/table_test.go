package extension

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"
	"testing"
)

func TestTable(t *testing.T) {
	markdown := goldmark.New(
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
		goldmark.WithExtensions(
			Table,
		),
	)
	goldmark.DoTestCaseFile(markdown, "_test/table.txt", t)
}
