package benchmark

import (
	"bytes"
	"os"
	"testing"

	gomarkdown "github.com/gomarkdown/markdown"
	"github.com/yuin/goldmark"
	v1html "github.com/yuin/goldmark/renderer/html"

	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer/html"
	"github.com/yuin/goldmark/v2/util"
	"gitlab.com/golang-commonmark/markdown"

	"github.com/88250/lute"
)

func BenchmarkMarkdown(b *testing.B) {
	b.Run("goldmark/v2", func(b *testing.B) {
		gp := parser.New()
		gr := html.New(html.WithXHTML(), html.WithUnsafe())
		r := func(src []byte) ([]byte, error) {
			var out bytes.Buffer
			err := gr.Render(&out, src, gp.Parse(src))
			return out.Bytes(), err
		}
		doBenchmark(b, r)
	})

	b.Run("goldmark/v1", func(b *testing.B) {
		markdown := goldmark.New(
			goldmark.WithRendererOptions(v1html.WithXHTML(), v1html.WithUnsafe()),
		)
		r := func(src []byte) ([]byte, error) {
			var out bytes.Buffer
			err := markdown.Convert(src, &out)
			return out.Bytes(), err
		}
		doBenchmark(b, r)
	})

	b.Run("CommonMark", func(b *testing.B) {
		md := markdown.New(markdown.XHTMLOutput(true))
		r := func(src []byte) ([]byte, error) {
			var out bytes.Buffer
			err := md.Render(&out, src)
			return out.Bytes(), err
		}
		doBenchmark(b, r)
	})

	b.Run("Lute", func(b *testing.B) {
		luteEngine := lute.New()
		luteEngine.SetGFMAutoLink(false)
		luteEngine.SetGFMStrikethrough(false)
		luteEngine.SetGFMTable(false)
		luteEngine.SetGFMTaskListItem(false)
		luteEngine.SetCodeSyntaxHighlight(false)
		luteEngine.SetSoftBreak2HardBreak(false)
		luteEngine.SetAutoSpace(false)
		luteEngine.SetFixTermTypo(false)
		r := func(src []byte) ([]byte, error) {
			out := luteEngine.MarkdownStr("Benchmark", util.BytesToReadOnlyString(src))
			return util.StringToReadOnlyBytes(out), nil
		}
		doBenchmark(b, r)
	})

	b.Run("GoMarkdown", func(b *testing.B) {
		r := func(src []byte) ([]byte, error) {
			out := gomarkdown.ToHTML(src, nil, nil)
			return out, nil
		}
		doBenchmark(b, r)
	})

}

// The different frameworks have different APIs. Create an adapter that
// should behave the same in the memory department.
func doBenchmark(b *testing.B, render func(src []byte) ([]byte, error)) {
	b.StopTimer()
	source, err := os.ReadFile("_data.md")
	if err != nil {
		b.Fatal(err)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		out, err := render(source)
		if err != nil {
			b.Fatal(err)
		}
		if len(out) < 100 {
			b.Fatal("No result")
		}
	}
}
