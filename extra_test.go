package goldmark_test

import (
	"testing"

	. "github.com/yuin/goldmark"
	"github.com/yuin/goldmark/testutil"
	"github.com/yuin/goldmark/renderer/html"
)

func TestDefinitionList(t *testing.T) {
	markdown := New(WithRendererOptions(
		html.WithXHTML(),
		html.WithUnsafe(),
	))
	testutil.DoTestCaseFile(markdown, "_test/extra.txt", t)
}
