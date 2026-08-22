package main

import (
	"bytes"
	"os"
	"syscall/js"

	"github.com/yuin/goldmark/v2/extension"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer/html"
)

const (
	optTableExtension int = 1 << iota
	optStrikethroughExtension
	optLinkifyExtension
	optTaskListExtension
	optDefinitionListExtension
	optFootnoteExtension
	optTypographerExtension
	optXHTML
	optUnsafe
)

func toHtml(_ js.Value, args []js.Value) any {
	source := args[0].String()
	opts := args[1].Int()
	out := convert(source, opts)
	return out
}

func dumpAST(_ js.Value, args []js.Value) any {
	source := args[0].String()
	opts := args[1].Int()
	dump(source, opts)
	return nil
}

func main() {
	c := make(chan struct{})

	js.Global().Set("toHtml", js.FuncOf(toHtml))
	js.Global().Set("dumpAst", js.FuncOf(dumpAST))
	js.Global().Set("optTableExtension", js.ValueOf(optTableExtension))
	js.Global().Set("optStrikethroughExtension", js.ValueOf(optStrikethroughExtension))
	js.Global().Set("optLinkifyExtension", js.ValueOf(optLinkifyExtension))
	js.Global().Set("optTaskListExtension", js.ValueOf(optTaskListExtension))
	js.Global().Set("optDefinitionListExtension", js.ValueOf(optDefinitionListExtension))
	js.Global().Set("optFootnoteExtension", js.ValueOf(optFootnoteExtension))
	js.Global().Set("optTypographerExtension", js.ValueOf(optTypographerExtension))
	js.Global().Set("optXHTML", js.ValueOf(optXHTML))
	js.Global().Set("optUnsafe", js.ValueOf(optUnsafe))

	<-c
}

func parseOptions(opts int) ([]parser.Option, []html.Option) {
	var popts []parser.Option
	var ropts []html.Option

	if opts&optTableExtension == optTableExtension {
		popts = append(popts, parser.WithExtensions(extension.NewTableParser()))
		ropts = append(ropts, html.WithExtensions(extension.NewTableHTMLRenderer()))
	}
	if opts&optStrikethroughExtension == optStrikethroughExtension {
		popts = append(popts, parser.WithExtensions(extension.NewStrikethroughParser()))
		ropts = append(ropts, html.WithExtensions(extension.NewStrikethroughHTMLRenderer()))
	}
	if opts&optLinkifyExtension == optLinkifyExtension {
		popts = append(popts, parser.WithExtensions(extension.NewLinkifyParser()))
	}
	if opts&optTaskListExtension == optTaskListExtension {
		popts = append(popts, parser.WithExtensions(extension.NewTaskListItemParser()))
		ropts = append(ropts, html.WithExtensions(extension.NewTaskListItemHTMLRenderer()))
	}
	if opts&optDefinitionListExtension == optDefinitionListExtension {
		popts = append(popts, parser.WithExtensions(extension.NewDefinitionListParser()))
		ropts = append(ropts, html.WithExtensions(extension.NewDefinitionListHTMLRenderer()))
	}
	if opts&optFootnoteExtension == optFootnoteExtension {
		popts = append(popts, parser.WithExtensions(extension.NewFootnoteParser()))
		ropts = append(ropts, html.WithExtensions(extension.NewFootnoteHTMLRenderer()))
	}
	if opts&optTypographerExtension == optTypographerExtension {
		popts = append(popts, parser.WithExtensions(extension.NewTypographerParser()))
	}

	if opts&optXHTML == optXHTML {
		ropts = append(ropts, html.WithXHTML())
	}
	if opts&optUnsafe == optUnsafe {
		ropts = append(ropts, html.WithUnsafe())
	}

	return popts, ropts
}

func dump(s string, opts int) {
	source := []byte(s)

	popts, _ := parseOptions(opts)

	p := parser.New(popts...)

	node := p.Parse(source)
	d := node.Dump(source)
	d.PrettyPrint(os.Stdout, source)
}

func convert(s string, opts int) string {
	source := []byte(s)
	var out bytes.Buffer

	popts, ropts := parseOptions(opts)

	p := parser.New(popts...)
	r := html.New(ropts...)

	node := p.Parse(source)
	_ = r.Render(&out, source, node)

	return out.String()
}
