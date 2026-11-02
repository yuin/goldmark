package html_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/extension"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer"
	"github.com/yuin/goldmark/v2/renderer/html"
)

func render(t *testing.T, r html.Renderer, source string) string {
	t.Helper()
	var buf bytes.Buffer
	src := []byte(source)
	if err := r.Render(&buf, src, parser.New().Parse(src)); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

func TestWithNodeRendererOverridesCommonMark(t *testing.T) {
	strong := html.NodeRendererFunc(func(
		w io.Writer, _ []byte, _ ast.Node, entering bool, _ renderer.Context,
	) (ast.WalkStatus, error) {
		if entering {
			_, _ = io.WriteString(w, `<b class="important">`)
		} else {
			_, _ = io.WriteString(w, `</b>`)
		}
		return ast.WalkContinue, nil
	})
	r := html.New(html.WithNodeRenderer(ast.KindStrong, strong))
	got := render(t, r, "This is **important**.\n")
	want := "<p>This is <b class=\"important\">important</b>.</p>\n"
	if got != want {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestWithNodeRenderersOverridesCommonMark(t *testing.T) {
	emphasis := html.NodeRendererFunc(func(
		w io.Writer, _ []byte, _ ast.Node, entering bool, _ renderer.Context,
	) (ast.WalkStatus, error) {
		if entering {
			_, _ = io.WriteString(w, `<i class="important">`)
		} else {
			_, _ = io.WriteString(w, `</i>`)
		}
		return ast.WalkContinue, nil
	})
	r := html.New(html.WithNodeRenderers(map[ast.NodeKind]html.NodeRenderer{
		ast.KindEmphasis: emphasis,
	}))
	got := render(t, r, "*important*\n")
	want := "<p><i class=\"important\">important</i></p>\n"
	if got != want {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestWithNodeRendererOverridesExtension(t *testing.T) {
	paragraph := html.NodeRendererFunc(func(
		w io.Writer, _ []byte, _ ast.Node, entering bool, _ renderer.Context,
	) (ast.WalkStatus, error) {
		if entering {
			_, _ = io.WriteString(w, `<div>`)
		} else {
			_, _ = io.WriteString(w, `</div>`)
		}
		return ast.WalkContinue, nil
	})
	r := html.New(
		html.WithExtensions(extension.NewTaskListItemHTMLRenderer()),
		html.WithNodeRenderer(ast.KindParagraph, paragraph),
	)
	got := render(t, r, "- [ ] todo\n")
	want := "<ul>\n<li><div>[ ] todo</div></li>\n</ul>\n"
	if got != want {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestWithNodeRendererDecorator(t *testing.T) {
	decorator := func(next html.NodeRenderer) html.NodeRenderer {
		return html.NodeRendererFunc(func(
			w io.Writer, source []byte, n ast.Node, entering bool, rc renderer.Context,
		) (ast.WalkStatus, error) {
			if entering {
				_, _ = io.WriteString(w, `<!-- decorated -->`)
			}
			return next.Render(w, source, n, entering, rc)
		})
	}
	r := html.New(html.WithNodeRendererDecorator(ast.KindStrong, decorator))
	got := render(t, r, "**important**\n")
	want := "<p><!-- decorated --><strong>important</strong></p>\n"
	if got != want {
		t.Errorf("want %q, got %q", want, got)
	}
}
