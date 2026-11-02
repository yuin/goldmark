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

var thematicBreakCacheKey = NewContextKey()

type thematicBreakCache struct {
	stop int
	mark byte
	brk  int
}

func scanThematicBreak(line []byte, offset int) (isBreak bool, mark byte, pos int, brk int) {
	w, pos := util.IndentWidth(line, offset)
	if w > 3 {
		return false, 0, pos, -1
	}
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
			return false, mark, pos, -1
		}
		if c != mark {
			return false, mark, pos, i
		}
		count++
	}
	return count > 2, mark, pos, -1
}

func isThematicBreak(line []byte, offset int) bool {
	ok, _, _, _ := scanThematicBreak(line, offset)
	return ok
}

func (b *thematicBreakParser) Trigger() []byte {
	return []byte{'-', '*', '_'}
}

func (b *thematicBreakParser) Open(_ ast.Node, reader text.Reader, pc Context) (ast.Node, State) {
	line, segment := reader.PeekLine()
	cache := pc.ComputeIfAbsent(thematicBreakCacheKey, func() any {
		return &thematicBreakCache{brk: -1}
	}).(*thematicBreakCache)
	if cache.stop == segment.Stop && cache.brk >= 0 {
		for i := range line {
			if line[i] == ' ' || line[i] == '\t' {
				continue
			}
			if line[i] == cache.mark && segment.Start+i < cache.brk {
				return nil, NoChildren
			}
			break
		}
	}
	ok, mark, _, brk := scanThematicBreak(line, reader.LineOffset())
	if ok {
		reader.AdvanceToEOL()
		return ast.NewThematicBreak(), NoChildren
	}
	if brk >= 0 {
		cache.stop = segment.Stop
		cache.mark = mark
		cache.brk = segment.Start + brk
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
