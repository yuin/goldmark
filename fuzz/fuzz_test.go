package fuzz

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

func Fuzz(f *testing.F) {
	bs, err := ioutil.ReadFile("../_test/spec.json")
	if err != nil {
		panic(err)
	}
	var testCases []map[string]interface{}
	if err := json.Unmarshal(bs, &testCases); err != nil {
		panic(err)
	}
	for _, c := range testCases {
		f.Add(c["markdown"])
	}

	f.Fuzz(func(t *testing.T, orig string) {
		markdown := goldmark.New(
			goldmark.WithParserOptions(
				parser.WithAutoHeadingID(),
				parser.WithAttribute(),
			),
			goldmark.WithRendererOptions(
				html.WithUnsafe(),
				html.WithXHTML(),
			),
			goldmark.WithExtensions(
				extension.DefinitionList,
				extension.Footnote,
				extension.GFM,
				extension.Typographer,
				extension.Linkify,
				extension.Table,
				extension.TaskList,
			),
		)
		var b bytes.Buffer
		if err := markdown.Convert(util.StringToReadOnlyBytes(orig), &b); err != nil {
			panic(err)
		}
	})
}
