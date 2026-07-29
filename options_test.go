//go:build !goldmark_v1_attribute

package goldmark_test

import (
	"testing"

	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer/html"
	"github.com/yuin/goldmark/v2/testutil"
)

func TestAttributeAndAutoHeadingID(t *testing.T) {
	markdown := testutil.NewMarkdownToStringFunc(
		parser.New(parser.WithAttribute(), parser.WithAutoHeadingID()),
		html.New(),
	)
	testutil.DoTestCaseFile(markdown, "_test/options.txt", t, testutil.ParseCliCaseArg()...)
}
