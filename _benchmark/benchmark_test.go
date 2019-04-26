package main

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"

	"gitlab.com/golang-commonmark/markdown"

	"gopkg.in/russross/blackfriday.v2"
)

func BenchmarkGoldMark(b *testing.B) {
	b.ResetTimer()
	source, err := ioutil.ReadFile("_data.md")
	if err != nil {
		panic(err)
	}
	markdown := goldmark.New(goldmark.WithRendererOptions(html.WithXHTML()))
	var out bytes.Buffer
	markdown.Convert([]byte(""), &out)

	for i := 0; i < b.N; i++ {
		out.Reset()
		if err := markdown.Convert(source, &out); err != nil {
			panic(err)
		}
	}
}

func BenchmarkGolangCommonMark(b *testing.B) {
	b.ResetTimer()
	source, err := ioutil.ReadFile("_data.md")
	if err != nil {
		panic(err)
	}
	md := markdown.New(markdown.XHTMLOutput(true))
	for i := 0; i < b.N; i++ {
		md.RenderToString(source)
	}
}

func BenchmarkBlackFriday(b *testing.B) {
	b.ResetTimer()
	source, err := ioutil.ReadFile("_data.md")
	if err != nil {
		panic(err)
	}
	for i := 0; i < b.N; i++ {
		blackfriday.Run(source)
	}
}
