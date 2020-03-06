package renderer

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

func TestRenderUnknownNode(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("Render() should panic")
			return
		}

		msg, ok := r.(string)
		if !ok {
			t.Errorf("Render() should panic with a string")
		}

		if !strings.Contains(msg, "unrecognized node kind Document") {
			t.Errorf("Render() panicked with unexpected message: %v", msg)
		}
	}()

	src := []byte("# Foo\n\nHello world")
	node := parser.NewParser().Parse(text.NewReader(src))

	r := NewRenderer()
	r.Render(ioutil.Discard, src, node)
}
