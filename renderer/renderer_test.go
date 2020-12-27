package renderer_test

import (
	"testing"

	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

type invalid struct{}

func TestInvalidNodeRenderer(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("did not panic")
			return
		}

		rerr := r.(error)
		if rerr.Error() != "*renderer_test.invalid is not a NodeRenderer" {
			t.Errorf("unexpected panic caught: %v", rerr)
		}
	}()

	r := renderer.NewRenderer(renderer.WithNodeRenderers(util.Prioritized(&invalid{}, 1000)))
	r.Render(nil, nil, nil)
}
