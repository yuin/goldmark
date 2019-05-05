package extension

import (
	"github.com/yuin/goldmark"
	gast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// TypographicPunctuation is a key of the punctuations that can be replaced with
// typographic entities.
type TypographicPunctuation int

const (
	// LeftSingleQuote is '
	LeftSingleQuote TypographicPunctuation = iota + 1
	// RightSingleQuote is '
	RightSingleQuote
	// LeftDoubleQuote is "
	LeftDoubleQuote
	// RightDoubleQuote is "
	RightDoubleQuote
	// EnDash is --
	EnDash
	// EmDash is ---
	EmDash
	// Ellipsis is ...
	Ellipsis
	// LeftAngleQuote is <<
	LeftAngleQuote
	// RightAngleQuote is >>
	RightAngleQuote

	typographicPunctuationMax
)

// An TypographerConfig struct is a data structure that holds configuration of the
// Typographer extension.
type TypographerConfig struct {
	Substitutions [][]byte
}

func newDefaultSubstitutions() [][]byte {
	replacements := make([][]byte, typographicPunctuationMax)
	replacements[LeftSingleQuote] = []byte("&lsquo;")
	replacements[RightSingleQuote] = []byte("&rsquo;")
	replacements[LeftDoubleQuote] = []byte("&ldquo;")
	replacements[RightDoubleQuote] = []byte("&rdquo;")
	replacements[EnDash] = []byte("&ndash;")
	replacements[EmDash] = []byte("&mdash;")
	replacements[Ellipsis] = []byte("&hellip;")
	replacements[LeftAngleQuote] = []byte("&laquo;")
	replacements[RightAngleQuote] = []byte("&raquo;")

	return replacements
}

// SetOption implements SetOptioner.
func (b *TypographerConfig) SetOption(name parser.OptionName, value interface{}) {
	switch name {
	case optTypographicSubstitutions:
		b.Substitutions = value.([][]byte)
	}
}

// A TypographerOption interface sets options for the TypographerParser.
type TypographerOption interface {
	parser.Option
	SetTypographerOption(*TypographerConfig)
}

const optTypographicSubstitutions parser.OptionName = "TypographicSubstitutions"

// TypographicSubstitutions is a list of the substitutions for the Typographer extension.
type TypographicSubstitutions map[TypographicPunctuation][]byte

type withTypographicSubstitutions struct {
	value [][]byte
}

func (o *withTypographicSubstitutions) SetParserOption(c *parser.Config) {
	c.Options[optTypographicSubstitutions] = o.value
}

func (o *withTypographicSubstitutions) SetTypographerOption(p *TypographerConfig) {
	p.Substitutions = o.value
}

// WithTypographicSubstitutions is a functional otpion that specify replacement text
// for punctuations.
func WithTypographicSubstitutions(values map[TypographicPunctuation][]byte) TypographerOption {
	replacements := newDefaultSubstitutions()
	for k, v := range values {
		replacements[k] = v
	}

	return &withTypographicSubstitutions{replacements}
}

type typographerDelimiterProcessor struct {
}

func (p *typographerDelimiterProcessor) IsDelimiter(b byte) bool {
	return b == '\'' || b == '"'
}

func (p *typographerDelimiterProcessor) CanOpenCloser(opener, closer *parser.Delimiter) bool {
	return opener.Char == closer.Char
}

func (p *typographerDelimiterProcessor) OnMatch(consumes int) gast.Node {
	return nil
}

var defaultTypographerDelimiterProcessor = &typographerDelimiterProcessor{}

type typographerParser struct {
	TypographerConfig
}

// NewTypographerParser return a new InlineParser that parses
// typographer expressions.
func NewTypographerParser(opts ...TypographerOption) parser.InlineParser {
	p := &typographerParser{
		TypographerConfig: TypographerConfig{
			Substitutions: newDefaultSubstitutions(),
		},
	}
	for _, o := range opts {
		o.SetTypographerOption(&p.TypographerConfig)
	}
	return p
}

func (s *typographerParser) Trigger() []byte {
	return []byte{'\'', '"', '-', '.', '<', '>'}
}

func (s *typographerParser) Parse(parent gast.Node, block text.Reader, pc parser.Context) gast.Node {
	before := block.PrecendingCharacter()
	line, _ := block.PeekLine()
	c := line[0]
	if len(line) > 2 {
		if c == '-' {
			if s.Substitutions[EmDash] != nil && line[1] == '-' && line[2] == '-' { // ---
				node := ast.NewTypographicText(s.Substitutions[EmDash])
				block.Advance(3)
				return node
			}
		} else if c == '.' {
			if s.Substitutions[Ellipsis] != nil && line[1] == '.' && line[2] == '.' { // ...
				node := ast.NewTypographicText(s.Substitutions[Ellipsis])
				block.Advance(3)
				return node
			}
			return nil
		}
	}
	if len(line) > 1 {
		if c == '<' {
			if s.Substitutions[LeftAngleQuote] != nil && line[1] == '<' { // <<
				node := ast.NewTypographicText(s.Substitutions[LeftAngleQuote])
				block.Advance(2)
				return node
			}
			return nil
		} else if c == '>' {
			if s.Substitutions[RightAngleQuote] != nil && line[1] == '>' { // >>
				node := ast.NewTypographicText(s.Substitutions[RightAngleQuote])
				block.Advance(2)
				return node
			}
			return nil
		} else if s.Substitutions[EnDash] != nil && c == '-' && line[1] == '-' { // --
			node := ast.NewTypographicText(s.Substitutions[EnDash])
			block.Advance(2)
			return node
		}
	}
	if c == '\'' || c == '"' {
		d := parser.ScanDelimiter(line, before, 1, defaultTypographerDelimiterProcessor)
		if d == nil {
			return nil
		}
		if c == '\'' {
			if s.Substitutions[LeftSingleQuote] != nil && d.CanOpen && !d.CanClose {
				node := ast.NewTypographicText(s.Substitutions[LeftSingleQuote])
				block.Advance(1)
				return node
			}
			if s.Substitutions[RightSingleQuote] != nil && d.CanClose && !d.CanOpen {
				node := ast.NewTypographicText(s.Substitutions[RightSingleQuote])
				block.Advance(1)
				return node
			}
		}
		if c == '"' {
			if s.Substitutions[LeftDoubleQuote] != nil && d.CanOpen && !d.CanClose {
				node := ast.NewTypographicText(s.Substitutions[LeftDoubleQuote])
				block.Advance(1)
				return node
			}
			if s.Substitutions[RightDoubleQuote] != nil && d.CanClose && !d.CanOpen {
				node := ast.NewTypographicText(s.Substitutions[RightDoubleQuote])
				block.Advance(1)
				return node
			}
		}
	}
	return nil
}

func (s *typographerParser) CloseBlock(parent gast.Node, pc parser.Context) {
	// nothing to do
}

// TypographerHTMLRenderer is a renderer.NodeRenderer implementation that
// renders Typographer nodes.
type TypographerHTMLRenderer struct {
	html.Config
}

// NewTypographerHTMLRenderer returns a new TypographerHTMLRenderer.
func NewTypographerHTMLRenderer(opts ...html.Option) renderer.NodeRenderer {
	r := &TypographerHTMLRenderer{
		Config: html.NewConfig(),
	}
	for _, opt := range opts {
		opt.SetHTMLOption(&r.Config)
	}
	return r
}

// RegisterFuncs implements renderer.NodeRenderer.RegisterFuncs.
func (r *TypographerHTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindTypographicText, r.renderTypographicText)
}

func (r *TypographerHTMLRenderer) renderTypographicText(w util.BufWriter, source []byte, n gast.Node, entering bool) (gast.WalkStatus, error) {
	if entering {
		w.Write(n.(*ast.TypographicText).Value)
	}
	return gast.WalkContinue, nil
}

type typographer struct {
	options []TypographerOption
}

// Typographer is an extension that repalace punctuations with typographic entities.
var Typographer = &typographer{}

// NewTypographer returns a new Entender that repalace punctuations with typographic entities.
func NewTypographer(opts ...TypographerOption) goldmark.Extender {
	return &typographer{
		options: opts,
	}
}

func (e *typographer) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithInlineParsers(
		util.Prioritized(NewTypographerParser(e.options...), 9999),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(NewTypographerHTMLRenderer(), 500),
	))
}
