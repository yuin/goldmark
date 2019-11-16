package goldmark_test

import (
	"bytes"
	"testing"

	. "github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/testutil"
)

func TestExtras(t *testing.T) {
	markdown := New(WithRendererOptions(
		html.WithXHTML(),
		html.WithUnsafe(),
	))
	testutil.DoTestCaseFile(markdown, "_test/extra.txt", t)
}

func TestEndsWithNonSpaceCharacters(t *testing.T) {
	markdown := New(WithRendererOptions(
		html.WithXHTML(),
		html.WithUnsafe(),
	))
	source := []byte("```\na\n```")
	var b bytes.Buffer
	err := markdown.Convert(source, &b)
	if err != nil {
		t.Error(err.Error())
	}
	if b.String() != "<pre><code>a\n</code></pre>\n" {
		t.Errorf("%s \n---------\n %s", source, b.String())
	}
}
