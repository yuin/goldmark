package extension

import (
	"testing"

	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer/html"
	"github.com/yuin/goldmark/v2/testutil"
)

func TestDefinitionList(t *testing.T) {
	markdown := testutil.NewMarkdownToStringFunc(
		parser.New(
			parser.WithExtensions(NewDefinitionListParser()),
		),
		html.New(
			html.WithUnsafe(),
			html.WithExtensions(NewDefinitionListHTMLRenderer()),
		),
	)
	testutil.DoTestCaseFile(markdown, "_test/definition_list.txt", t, testutil.ParseCliCaseArg()...)
}
