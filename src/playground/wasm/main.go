package main

import (
	"bytes"
	"syscall/js"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
)

const (
	optTableExtension int = 1 << iota
	optStrikethroughExtension
	optLinkifyExtension
	optTaskListExtension
	optDefinitionListExtension
	optFootnoteExtension
	optTypographerExtension
	optCJKExtension
	optXHTML
	optUnsafe
)

func toHtml(_ js.Value, args []js.Value) any {
	source := args[0].String()
	opts := args[1].Int()
	out := convert(source, opts)
	return out
}

func main() {
	c := make(chan struct{}, 0)

	js.Global().Set("toHtml", js.FuncOf(toHtml))
	js.Global().Set("optTableExtension", js.ValueOf(optTableExtension))
	js.Global().Set("optStrikethroughExtension", js.ValueOf(optStrikethroughExtension))
	js.Global().Set("optLinkifyExtension", js.ValueOf(optLinkifyExtension))
	js.Global().Set("optTaskListExtension", js.ValueOf(optTaskListExtension))
	js.Global().Set("optDefinitionListExtension", js.ValueOf(optDefinitionListExtension))
	js.Global().Set("optFootnoteExtension", js.ValueOf(optFootnoteExtension))
	js.Global().Set("optTypographerExtension", js.ValueOf(optTypographerExtension))
	js.Global().Set("optCJKExtension", js.ValueOf(optCJKExtension))
	js.Global().Set("optXHTML", js.ValueOf(optXHTML))
	js.Global().Set("optUnsafe", js.ValueOf(optUnsafe))

	<-c
}

func convert(s string, opts int) string {
	source := []byte(s)
	var out bytes.Buffer

	var extensions []goldmark.Extender
	var renderer []renderer.Option

	if opts&optTableExtension == optTableExtension {
		extensions = append(extensions, extension.Table)
	}
	if opts&optStrikethroughExtension == optStrikethroughExtension {
		extensions = append(extensions, extension.Strikethrough)
	}
	if opts&optLinkifyExtension == optLinkifyExtension {
		extensions = append(extensions, extension.Linkify)
	}
	if opts&optTaskListExtension == optTaskListExtension {
		extensions = append(extensions, extension.TaskList)
	}
	if opts&optDefinitionListExtension == optDefinitionListExtension {
		extensions = append(extensions, extension.DefinitionList)
	}
	if opts&optFootnoteExtension == optFootnoteExtension {
		extensions = append(extensions, extension.Footnote)
	}
	if opts&optTypographerExtension == optTypographerExtension {
		extensions = append(extensions, extension.Typographer)
	}
	if opts&optCJKExtension == optCJKExtension {
		extensions = append(extensions, extension.CJK)
	}

	if opts&optXHTML == optXHTML {
		renderer = append(renderer, html.WithXHTML())
	}
	if opts&optUnsafe == optUnsafe {
		renderer = append(renderer, html.WithUnsafe())
	}

	md := goldmark.New(
		goldmark.WithExtensions(extensions...),
		goldmark.WithRendererOptions(renderer...),
	)

	_ = md.Convert(source, &out)
	return out.String()
}
