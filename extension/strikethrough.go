package extension

import (
	"io"

	gast "github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/extension/ast"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer"
	"github.com/yuin/goldmark/v2/renderer/html"
	"github.com/yuin/goldmark/v2/text"
	"github.com/yuin/goldmark/v2/util"
)

type strikethroughDelimiterProcessor struct {
}

func (p *strikethroughDelimiterProcessor) IsDelimiter(b byte) bool {
	return b == '~'
}

func (p *strikethroughDelimiterProcessor) CanOpenCloser(opener, closer *parser.Delimiter) bool {
	return opener.Char == closer.Char
}

func (p *strikethroughDelimiterProcessor) OnMatch(_ int) gast.Node {
	return ast.NewStrikethrough()
}

var defaultStrikethroughDelimiterProcessor = &strikethroughDelimiterProcessor{}

type strikethroughParser struct {
}

var defaultStrikethroughParser = &strikethroughParser{}

func newStrikethroughParser() parser.InlineParser {
	return defaultStrikethroughParser
}

func (s *strikethroughParser) Trigger() []byte {
	return []byte{'~'}
}

func (s *strikethroughParser) Parse(_ gast.Node, block text.Reader, pc parser.Context) gast.Node {
	before := block.PrecedingCharacter()
	if before == '~' {
		return nil
	}
	line, _ := block.PeekLine()
	n := 0
	for n < len(line) && line[n] == '~' {
		n++
	}
	if n > 2 {
		return nil
	}
	return parser.ParseDelimiter(block, 1, defaultStrikethroughDelimiterProcessor, pc)
}

func (s *strikethroughParser) CloseBlock(_ gast.Node, _ parser.Context) {
	// nothing to do
}

type strikethroughHTMLRendererExtension struct {
}

// NewStrikethroughHTMLRenderer returns a new html.Extension for rendering Strikethrough nodes.
func NewStrikethroughHTMLRenderer() html.Extension {
	return &strikethroughHTMLRendererExtension{}
}

func (r *strikethroughHTMLRendererExtension) RendererOptions(_ *html.Config) []html.Option {
	return []html.Option{
		html.WithNodeRenderers(map[gast.NodeKind]html.NodeRenderer{
			ast.KindStrikethrough: html.NodeRendererFunc(r.renderStrikethrough),
		}),
	}
}

// StrikethroughAttributeFilter defines attribute names which del elements can have.
var StrikethroughAttributeFilter = html.GlobalAttributeFilter

func (r *strikethroughHTMLRendererExtension) renderStrikethrough(
	writer io.Writer, source []byte, n gast.Node, entering bool, rc renderer.Context) (gast.WalkStatus, error) {
	w := writer.(util.BufWriter)
	if entering {
		if n.Attributes() != nil {
			_, _ = w.WriteString("<del")
			html.RenderAttributes(w, source, n, StrikethroughAttributeFilter, rc)
			_ = w.WriteByte('>')
		} else {
			_, _ = w.WriteString("<del>")
		}
	} else {
		_, _ = w.WriteString("</del>")
	}
	return gast.WalkContinue, nil
}

type strikethroughParserExtension struct {
}

// NewStrikethroughParser returns a new parser.Extension for parsing strikethrough expressions.
func NewStrikethroughParser() parser.Extension {
	return &strikethroughParserExtension{}
}

func (e *strikethroughParserExtension) ParserOptions(_ *parser.Config) []parser.Option {
	return []parser.Option{
		parser.WithInlineParsers(
			util.Prioritized(newStrikethroughParser(), 500),
		),
	}
}

// StrikethroughParser is a default [parser.Extension] for parsing strikethrough expressions.
var StrikethroughParser = NewStrikethroughParser()

// StrikethroughHTMLRenderer is a default [html.Extension] for rendering Strikethrough nodes.
var StrikethroughHTMLRenderer = NewStrikethroughHTMLRenderer()
