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

type emphasisParser struct {
}

var defaultEmphasisParser = &emphasisParser{}

// NewEmphasisParser return a new InlineParser that parses emphasises.
func NewEmphasisParser() InlineParser {
	return defaultEmphasisParser
}

func (s *emphasisParser) Trigger() []byte {
	return []byte{'*', '_'}
}

func (s *emphasisParser) Parse(_ ast.Node, block text.Reader, pc Context) ast.Node {
	return ParseDelimiter(block, 1, defaultEmphasisDelimiterProcessor, pc)
}
