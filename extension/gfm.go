package extension

import (
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer/html"
)

type gfmParserExtension struct {
}

// NewGFMParser returns a new parser.Extension that provides GitHub Flavored Markdown parser functionalities.
func NewGFMParser() parser.Extension {
	return &gfmParserExtension{}
}

func (e *gfmParserExtension) ParserOptions(c *parser.Config) []parser.Option {
	var opts []parser.Option
	opts = append(opts, NewLinkifyParser().ParserOptions(c)...)
	opts = append(opts, NewTableParser().ParserOptions(c)...)
	opts = append(opts, NewStrikethroughParser().ParserOptions(c)...)
	opts = append(opts, NewTaskListItemParser().ParserOptions(c)...)
	return opts
}

type gfmHTMLRendererExtension struct {
}

// NewGFMHTMLRenderer returns a new html.Extension that provides GitHub Flavored Markdown HTML renderer functionalities.
func NewGFMHTMLRenderer() html.Extension {
	return &gfmHTMLRendererExtension{}
}

func (e *gfmHTMLRendererExtension) RendererOptions(c *html.Config) []html.Option {
	var opts []html.Option
	opts = append(opts, NewTableHTMLRenderer().RendererOptions(c)...)
	opts = append(opts, NewStrikethroughHTMLRenderer().RendererOptions(c)...)
	opts = append(opts, NewTaskListItemHTMLRenderer().RendererOptions(c)...)
	return opts
}

// GFMParser is a default [parser.Extension] that provides GitHub Flavored Markdown parser functionalities.
var GFMParser = NewGFMParser()

// GFMHTMLRenderer is a default [html.Extension] that provides GitHub Flavored Markdown HTML renderer functionalities.
var GFMHTMLRenderer = NewGFMHTMLRenderer()
