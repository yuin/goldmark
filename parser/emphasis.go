package parser

import (
	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/text"
)

type emphasisDelimiterProcessor struct {
}

func (p *emphasisDelimiterProcessor) IsDelimiter(b byte) bool {
	return b == '*' || b == '_'
}

func (p *emphasisDelimiterProcessor) CanOpenCloser(opener, closer *Delimiter) bool {
	return opener.Char == closer.Char
}

func (p *emphasisDelimiterProcessor) OnMatch(consumes int) ast.Node {
	if consumes == 1 {
		return ast.NewEmphasis()
	}
	return ast.NewStrong()
}

var defaultEmphasisDelimiterProcessor = &emphasisDelimiterProcessor{}

// EmphasisConfig struct is a data structure that holds configuration of the parsers related to emphasis.
type EmphasisConfig struct {
	f ParseDelimiterFunc
}

// A EmphasisOption interface sets options for emphasis parsers.
type EmphasisOption interface {
	setEmphasisOption(*EmphasisConfig)
}

type emphasisParser struct {
	f ParseDelimiterFunc
}

// NewEmphasisParser return a new InlineParser that parses emphasises.
func NewEmphasisParser(opts ...EmphasisOption) InlineParser {
	config := EmphasisConfig{
		f: ParseDelimiter,
	}
	for _, o := range opts {
		o.setEmphasisOption(&config)
	}
	return &emphasisParser{config.f}
}

func (s *emphasisParser) Trigger() []byte {
	return []byte{'*', '_'}
}

func (s *emphasisParser) Parse(_ ast.Node, block text.Reader, pc Context) ast.Node {
	return s.f(block, 1, defaultEmphasisDelimiterProcessor, pc)
}
