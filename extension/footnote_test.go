package extension

import (
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/testutil"
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
	testutil.DoTestCaseFile(markdown, "_test/footnote.txt", t, testutil.ParseCliCaseArg()...)
}
