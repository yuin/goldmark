package parser

import (
	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/text"
	"github.com/yuin/goldmark/v2/util"
)

type codeBlockParser struct {
}

var defaultCodeBlockParser = &codeBlockParser{}

// NewCodeBlockParser returns a new BlockParser that
// parses code blocks.
func NewCodeBlockParser() BlockParser {
	return defaultCodeBlockParser
}

func (b *codeBlockParser) Trigger() []byte {
	return nil
}

func (b *codeBlockParser) Open(_ ast.Node, reader text.Reader, _ Context) (ast.Node, State) {
	line, segment := reader.PeekLine()
	pos, padding := util.IndentPosition(line, reader.LineOffset(), 4)
	if pos < 0 || util.IsBlank(line) {
		return nil, NoChildren
	}
	node := ast.NewCodeBlock(ast.CodeBlockKindIndented, text.Lines{})
	reader.AdvanceAndSetPadding(pos, padding)
	_, segment = reader.PeekLine()
	// if code block line starts with a tab, keep a tab as it is.
	if segment.Padding != 0 {
		preserveLeadingTabInCodeBlock(&segment, reader, 0)
	}
	segment.ForceNewline = true
	node.Value.AppendSegment(segment)
	reader.AdvanceToEOL()
	return node, NoChildren

}

func (b *codeBlockParser) Continue(node ast.Node, reader text.Reader, _ Context) State {
	cb := node.(*ast.CodeBlock)
	line, segment := reader.PeekLine()
	if util.IsBlank(line) {
		cb.Value.AppendSegment(segment.TrimLeftSpaceWidth(4, reader.Source()))
		return Continue | NoChildren
	}
	pos, padding := util.IndentPosition(line, reader.LineOffset(), 4)
	if pos < 0 {
		return Close
	}
	reader.AdvanceAndSetPadding(pos, padding)
	_, segment = reader.PeekLine()

	// if code block line starts with a tab, keep a tab as it is.
	if segment.Padding != 0 {
		preserveLeadingTabInCodeBlock(&segment, reader, 0)
	}

	segment.ForceNewline = true
	cb.Value.AppendSegment(segment)
	reader.AdvanceToEOL()
	return Continue | NoChildren
}

func (b *codeBlockParser) Close(node ast.Node, reader text.Reader, _ Context) {
	// trim trailing blank lines
	cb := node.(*ast.CodeBlock)
	length := len(cb.Value.Segments()) - 1
	source := reader.Source()
	for length >= 0 {
		line := cb.Value.Segments()[length]
		if util.IsBlank(line.Bytes(source)) {
			length--
		} else {
			break
		}
	}
	// rebuild Lines with only [0, length+1) segments
	var segs []text.Segment
	for i := 0; i <= length; i++ {
		segs = append(segs, cb.Value.Segments()[i])
	}
	cb.Value = text.NewLinesFromSegments(segs)
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
