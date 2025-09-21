package parser

import (
	"sync"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

type emphasisDelimiterProcessor struct {
	isCJKFriendly bool
}

// IsCJKFriendly implements DelimiterProcessor.
func (p *emphasisDelimiterProcessor) IsCJKFriendly() bool {
	return p.isCJKFriendly
}

func (p *emphasisDelimiterProcessor) IsDelimiter(b byte) bool {
	return b == '*' || b == '_'
}

func (p *emphasisDelimiterProcessor) CanOpenCloser(opener, closer *Delimiter) bool {
	return opener.Char == closer.Char
}

func (p *emphasisDelimiterProcessor) OnMatch(consumes int) ast.Node {
	return ast.NewEmphasis(consumes)
}

var defaultEmphasisDelimiterProcessor = &emphasisDelimiterProcessor{}

type emphasisParser struct {
	EmphasisDelimiterProcessor *emphasisDelimiterProcessor
}

var defaultEmphasisParser = &emphasisParser{
	EmphasisDelimiterProcessor: defaultEmphasisDelimiterProcessor,
}

var getDefaultCJKFriendlyEmphaisisParser = sync.OnceValue(func() *emphasisParser {
	return &emphasisParser{
		EmphasisDelimiterProcessor: &emphasisDelimiterProcessor{
			isCJKFriendly: true,
		},
	}
})

// NewEmphasisParser return a new InlineParser that parses emphasises.
func NewEmphasisParser() InlineParser {
	return defaultEmphasisParser
}

func (s *emphasisParser) Trigger() []byte {
	return []byte{'*', '_'}
}

func (s *emphasisParser) Parse(parent ast.Node, block text.Reader, pc Context) ast.Node {
	before := block.PrecendingCharacter()
	line, segment := block.PeekLine()
	node := ScanDelimiter(line, before, 1, s.EmphasisDelimiterProcessor, block.TwoPrecedingCharacter)
	if node == nil {
		return nil
	}
	node.Segment = segment.WithStop(segment.Start + node.OriginalLength)
	block.Advance(node.OriginalLength)
	pc.PushDelimiter(node)
	return node
}

func (s *emphasisParser) GetCJKFriendlyVariant() CJKFriendlinessAwareEmphasisParser {
	return getDefaultCJKFriendlyEmphaisisParser()
}
