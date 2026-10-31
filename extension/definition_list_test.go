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
	testutil.DoTestCaseFile(markdown, "testdata/definition_list.txt", t, testutil.ParseCliCaseArg()...)
}

func TestDefinitionListWithGFMRendererOrder(t *testing.T) {
	for _, exts := range [][]html.Extension{
		{NewGFMHTMLRenderer(), NewDefinitionListHTMLRenderer()},
		{NewDefinitionListHTMLRenderer(), NewGFMHTMLRenderer()},
	} {
		markdown := testutil.NewMarkdownToStringFunc(
			parser.New(
				parser.WithExtensions(NewDefinitionListParser(), NewGFMParser()),
			),
			html.New(html.WithExtensions(exts...)),
		)
		got, err := markdown("Go\n:   A language.\n")
		if err != nil {
			t.Fatal(err)
		}
		expected := "<dl>\n<dt>Go</dt>\n<dd>A language.</dd>\n</dl>\n"
		if got != expected {
			t.Errorf("expected %q, got %q", expected, got)
		}
	}
}
