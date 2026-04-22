package extension

import (
	"testing"

	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer/html"
	"github.com/yuin/goldmark/v2/testutil"
)

func TestStrikethrough(t *testing.T) {
	markdown := testutil.NewMarkdownToStringFunc(
		parser.New(
			parser.WithExtensions(NewStrikethroughParser()),
		),
		html.New(
			html.WithUnsafe(),
			html.WithExtensions(NewStrikethroughHTMLRenderer()),
		),
	)
	testutil.DoTestCaseFile(markdown, "_test/strikethrough.txt", t, testutil.ParseCliCaseArg()...)
}
