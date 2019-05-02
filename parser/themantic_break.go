package parser

import (
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type themanticBreakParser struct {
}

var defaultThemanticBreakParser = &themanticBreakParser{}

// NewThemanticBreakParser returns a new BlockParser that
// parses themantic breaks.
func NewThemanticBreakParser() BlockParser {
	return defaultThemanticBreakParser
}

func isThemanticBreak(line []byte) bool {
	w, pos := util.IndentWidth(line, 0)
	if w > 3 {
		return false
	}
	mark := byte(0)
	count := 0
	for i := pos; i < len(line); i++ {
		c := line[i]
		if util.IsSpace(c) {
			continue
		}
		if mark == 0 {
			mark = c
			count = 1
			if mark == '*' || mark == '-' || mark == '_' {
				continue
			}
			return false
		}
		if c != mark {
			return false
		}
		count++
	}
	return count > 2
}

func (b *themanticBreakParser) Open(parent ast.Node, reader text.Reader, pc Context) (ast.Node, State) {
	line, segment := reader.PeekLine()
	if isThemanticBreak(line) {
		reader.Advance(segment.Len() - 1)
		return ast.NewThemanticBreak(), NoChildren
	}
	return nil, NoChildren
}

func (b *themanticBreakParser) Continue(node ast.Node, reader text.Reader, pc Context) State {
	return Close
}

func (b *themanticBreakParser) Close(node ast.Node, reader text.Reader, pc Context) {
	// nothing to do
}

func (b *themanticBreakParser) CanInterruptParagraph() bool {
	return true
}

func (b *themanticBreakParser) CanAcceptIndentedLine() bool {
	return false
}
