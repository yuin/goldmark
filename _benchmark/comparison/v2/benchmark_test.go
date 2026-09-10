package benchmark

import (
	"bytes"
	"os"
	"testing"

	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer/html"
)

func BenchmarkGoldmarkV2(b *testing.B) {
	p := parser.New()
	r := html.New(html.WithXHTML(), html.WithUnsafe())
	doBenchmark(b, func(source []byte) ([]byte, error) {
		var out bytes.Buffer
		err := r.Render(&out, source, p.Parse(source))
		return out.Bytes(), err
	})
}

func doBenchmark(b *testing.B, render func(src []byte) ([]byte, error)) {
	b.StopTimer()
	source, err := os.ReadFile("../../go/_data.md")
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
