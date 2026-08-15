package parser

import (
	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/text"
	"github.com/yuin/goldmark/v2/util"
)

type thematicBreakParser struct {
}

var defaultThematicBreakParser = &thematicBreakParser{}

// NewThematicBreakParser returns a new BlockParser that
// parses thematic breaks.
func NewThematicBreakParser() BlockParser {
	return defaultThematicBreakParser
}

func isThematicBreak(line []byte, offset int) bool {
	w, pos := util.IndentWidth(line, offset)
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

func (b *thematicBreakParser) Trigger() []byte {
	return []byte{'-', '*', '_'}
}

func (b *thematicBreakParser) Open(_ ast.Node, reader text.Reader, _ Context) (ast.Node, State) {
	line, _ := reader.PeekLine()
	if isThematicBreak(line, reader.LineOffset()) {
		reader.AdvanceToEOL()
		return ast.NewThematicBreak(), NoChildren
	}
	return nil, NoChildren
}

func (b *thematicBreakParser) Continue(_ ast.Node, _ text.Reader, _ Context) State {
	return Close
}

func (b *thematicBreakParser) Close(_ ast.Node, _ text.Reader, _ Context) {
	// nothing to do
}

func (b *thematicBreakParser) CanInterruptParagraph() bool {
	return true
}

func (b *thematicBreakParser) CanAcceptIndentedLine() bool {
	return false
}
