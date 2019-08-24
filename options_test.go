package goldmark_test

import (
	"testing"

	. "github.com/yuin/goldmark"
	"github.com/yuin/goldmark/testutil"
	"github.com/yuin/goldmark/parser"
)

func TestAttributeAndAutoHeadingID(t *testing.T) {
	markdown := New(
		WithParserOptions(
			parser.WithAttribute(),
			parser.WithAutoHeadingID(),
		),
	)
	testutil.DoTestCaseFile(markdown, "_test/options.txt", t)
}
