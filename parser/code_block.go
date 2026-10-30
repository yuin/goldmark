package parser

import (
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type codeBlockParser struct {
}

// CodeBlockParser is a BlockParser implementation that parses indented code blocks.
var defaultCodeBlockParser = &codeBlockParser{}

// NewCodeBlockParser returns a new BlockParser that
// parses code blocks.
func NewCodeBlockParser() BlockParser {
	return defaultCodeBlockParser
}

func (b *codeBlockParser) Trigger() []byte {
	return nil
}

func (b *codeBlockParser) Open(parent ast.Node, reader text.Reader, pc Context) (ast.Node, State) {
	line, segment := reader.PeekLine()
	pos, padding := util.IndentPositionPadding(line, reader.LineOffset(), segment.Padding, 4)
	if pos < 0 || util.IsBlank(line) {
		return nil, NoChildren
	}
	node := ast.NewCodeBlock()
	segment = text.NewSegmentPadding(segment.Start+pos, segment.Stop, padding)
	// if code block line starts with a tab, keep a tab as it is.
	if segment.Padding != 0 {
		preserveLeadingTabInCodeBlock(&segment, reader, 0)
	}
	segment.ForceNewline = true
	node.Lines().Append(segment)
	reader.AdvanceToEOL()
	return node, NoChildren

}

func (b *codeBlockParser) Continue(node ast.Node, reader text.Reader, pc Context) State {
	line, segment := reader.PeekLine()
	pos, padding := util.IndentPositionPadding(line, reader.LineOffset(), segment.Padding, 4)

	if util.IsBlank(line) && pos < 0 {
		segment = text.NewSegment(segment.Stop-1, segment.Stop)
	} else if pos < 0 {
		return Close
	} else {
		segment = text.NewSegmentPadding(segment.Start+pos, segment.Stop, padding)
		segment.ForceNewline = true
		// if code block line starts with a tab, keep a tab as it is.
		if segment.Padding != 0 {
			preserveLeadingTabInCodeBlock(&segment, reader, 0)
		}
	}
	node.Lines().Append(segment)
	reader.AdvanceToEOL()
	return Continue | NoChildren
}

func (b *codeBlockParser) Close(node ast.Node, reader text.Reader, pc Context) {
	// trim trailing blank lines
	lines := node.Lines()
	length := lines.Len() - 1
	source := reader.Source()
	for length >= 0 {
		line := lines.At(length)
		if util.IsBlank(line.Value(source)) {
			length--
		} else {
			break
		}
	}
	lines.SetSliced(0, length+1)
}

func (b *codeBlockParser) CanInterruptParagraph() bool {
	return false
}

func (b *codeBlockParser) CanAcceptIndentedLine() bool {
	return true
}

func preserveLeadingTabInCodeBlock(segment *text.Segment, reader text.Reader, indent int) {
	offsetWithPadding := reader.LineOffset() + indent
	sl, ss := reader.Position()
	reader.SetPosition(sl, text.NewSegment(ss.Start-1, ss.Stop))
	if offsetWithPadding == reader.LineOffset() {
		segment.Padding = 0
		segment.Start--
	}
	reader.SetPosition(sl, ss)
}
