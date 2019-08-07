package parser

import (
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type ThematicBreakParser struct {
}

var defaultThematicBreakParser = &ThematicBreakParser{}

// NewThematicBreakParser returns a new BlockParser that
// parses thematic breaks.
func NewThematicBreakParser() BlockParser {
	return defaultThematicBreakParser
}

func isThematicBreak(line []byte) bool {
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

func (b *ThematicBreakParser) Open(parent ast.Node, reader text.Reader, pc Context) (ast.Node, State) {
	line, segment := reader.PeekLine()
	if isThematicBreak(line) {
		reader.Advance(segment.Len() - 1)
		return ast.NewThematicBreak(), NoChildren
	}
	return nil, NoChildren
}

func (b *ThematicBreakParser) Continue(node ast.Node, reader text.Reader, pc Context) State {
	return Close
}

func (b *ThematicBreakParser) Close(node ast.Node, reader text.Reader, pc Context) {
	// nothing to do
}

func (b *ThematicBreakParser) CanInterruptParagraph() bool {
	return true
}

func (b *ThematicBreakParser) CanAcceptIndentedLine() bool {
	return false
}
