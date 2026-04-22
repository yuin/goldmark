package extension

import (
	"testing"

	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer/html"
	"github.com/yuin/goldmark/v2/testutil"
)

func TestTypographer(t *testing.T) {
	markdown := testutil.NewMarkdownToStringFunc(
		parser.New(
			parser.WithExtensions(NewTypographerParser()),
		),
		html.New(
			html.WithUnsafe(),
		),
	)
	testutil.DoTestCaseFile(markdown, "_test/typographer.txt", t, testutil.ParseCliCaseArg()...)
}
