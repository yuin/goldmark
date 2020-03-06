package renderer

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

func TestRenderUnknownNode(t *testing.T) {
	src := []byte("# Foo\n\nHello world")
	node := parser.NewParser().Parse(text.NewReader(src))

	r := NewRenderer()
	err := r.Render(ioutil.Discard, src, node)
	if err == nil {
		t.Fatalf("Render() expected error")
	}

	if !strings.Contains(err.Error(), "unrecognized node kind Document") {
		t.Errorf("Render() failed with unexpected error: %v", err)
	}
}
