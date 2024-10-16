package parser

import (
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type thematicBreakParser struct {
}

var defaultthematicBreakParser = &thematicBreakParser{}

// NewThematicBreakParser returns a new BlockParser that
// parses thematic breaks.
func NewThematicBreakParser() BlockParser {
	return defaultthematicBreakParser
}

// isThematicBreak returns true if a line indented
// with an offset contains only a set of three or more
// asterisks, dashes, or underscores mixed with characters
// deemed white space by [util.IsSpace].
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
			// note the mark character
			mark = c
			count = 1
			if mark == '*' || mark == '-' || mark == '_' {
				continue
			}
			// mark character was not recognized
			return false
		}
		if c != mark {
			// current character does not match
			// previously noted mark
			return false
		}
		count++
	}
	return count > 2
}

func (b *thematicBreakParser) Trigger() []byte {
	return []byte{'-', '*', '_'}
}

func (b *thematicBreakParser) Open(parent ast.Node, reader text.Reader, pc Context) (ast.Node, State) {
	line, segment := reader.PeekLine()
	if isThematicBreak(line, reader.LineOffset()) {
		reader.Advance(segment.Len() - 1)
		return ast.NewThematicBreak(), NoChildren
	}
	return nil, NoChildren
}

func (b *thematicBreakParser) Continue(node ast.Node, reader text.Reader, pc Context) State {
	return Close
}

func (b *thematicBreakParser) Close(node ast.Node, reader text.Reader, pc Context) {
	// nothing to do
}

func (b *thematicBreakParser) CanInterruptParagraph() bool {
	return true
}

func (b *thematicBreakParser) CanAcceptIndentedLine() bool {
	return false
}
