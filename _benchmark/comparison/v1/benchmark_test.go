package benchmark

import (
	"bytes"
	"os"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"
)

func BenchmarkGoldmarkV1(b *testing.B) {
	markdown := goldmark.New(
		goldmark.WithRendererOptions(html.WithXHTML(), html.WithUnsafe()),
	)
	doBenchmark(b, func(source []byte) ([]byte, error) {
		var out bytes.Buffer
		err := markdown.Convert(source, &out)
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
