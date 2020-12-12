package renderer_test

import (
	"bytes"
	"testing"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
)

type customLink struct {
	ast.Link
}

var kindCustomLink = ast.NewNodeKind("customLink")

// Kind implements Node.Kind.
func (n *customLink) Kind() ast.NodeKind {
	return kindCustomLink
}

func TestMissingRendererFunc(t *testing.T) {
	r := renderer.NewRenderer()
	buf := bytes.NewBuffer([]byte{})

	err := r.Render(buf, []byte{}, &customLink{})
	if err.Error() != "RendererFunc not found for kind: customLink" {
		t.Log(err)
		t.FailNow()
	}
}
