package main

import (
	"bytes"
	"io/ioutil"
	"testing"

	gomarkdown "github.com/gomarkdown/markdown"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
	"gitlab.com/golang-commonmark/markdown"

	bf1 "github.com/russross/blackfriday"
	bf2 "gopkg.in/russross/blackfriday.v2"

	"github.com/b3log/lute"
)

func BenchmarkMarkdown(b *testing.B) {
	b.Run("Blackfriday-v1", func(b *testing.B) {
		r := func(src []byte) ([]byte, error) {
			out := bf1.MarkdownBasic(src)
			return out, nil
		}
		doBenchmark(b, r)
	})

	b.Run("Blackfriday-v2", func(b *testing.B) {
		r := func(src []byte) ([]byte, error) {
			out := bf2.Run(src)
			return out, nil
		}
		doBenchmark(b, r)
	})

	b.Run("GoldMark(workers=16)", func(b *testing.B) {
		markdown := goldmark.New(
			goldmark.WithRendererOptions(html.WithXHTML(), html.WithUnsafe()),
		)
		r := func(src []byte) ([]byte, error) {
			var out bytes.Buffer
			err := markdown.Convert(src, &out, parser.WithWorkers(16))
			return out.Bytes(), err
		}
		doBenchmark(b, r)
	})

	b.Run("GoldMark", func(b *testing.B) {
		markdown := goldmark.New(
			goldmark.WithRendererOptions(html.WithXHTML(), html.WithUnsafe()),
		)
		r := func(src []byte) ([]byte, error) {
			var out bytes.Buffer
			err := markdown.Convert(src, &out, parser.WithWorkers(0))
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
		luteEngine := lute.New(
			lute.GFM(false),
			lute.CodeSyntaxHighlight(false),
			lute.SoftBreak2HardBreak(false),
			lute.AutoSpace(false),
			lute.FixTermTypo(false))
		r := func(src []byte) ([]byte, error) {
			out, err := luteEngine.MarkdownStr("Benchmark", util.BytesToReadOnlyString(src))
			return util.StringToReadOnlyBytes(out), err
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
	source, err := ioutil.ReadFile("_data.md")
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
