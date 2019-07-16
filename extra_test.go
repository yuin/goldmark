package goldmark

import (
	"testing"

	"github.com/yuin/goldmark/renderer/html"
)

func TestDefinitionList(t *testing.T) {
	markdown := New(WithRendererOptions(
		html.WithXHTML(),
		html.WithUnsafe(),
	))
	DoTestCaseFile(markdown, "_test/extra.txt", t)
}
