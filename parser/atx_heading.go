package parser

import (
	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/text"
	"github.com/yuin/goldmark/v2/util"
)

// A HeadingConfig struct is a data structure that holds configuration of the renderers related to headings.
type HeadingConfig struct {
	autoHeadingID bool
	attribute     bool
}

// A HeadingOption interface sets options for heading parsers.
type HeadingOption interface {
	setHeadingOption(*HeadingConfig)
}

type withAutoHeadingID struct{}

func (o *withAutoHeadingID) SetParserOption(c *Config) {
	c.autoHeadingID = true
}

func (o *withAutoHeadingID) setHeadingOption(p *HeadingConfig) {
	p.autoHeadingID = true
}

// WithAutoHeadingID is a functional option that enables custom heading ids and
// auto generated heading ids.
// It can be used as a parser Option and a HeadingOption.
func WithAutoHeadingID() interface {
	Option
	HeadingOption
} {
	return &withAutoHeadingID{}
}

type atxHeadingParser struct {
	HeadingConfig
}

// NewATXHeadingParser return a new BlockParser that can parse ATX headings.
func NewATXHeadingParser(opts ...HeadingOption) BlockParser {
	p := &atxHeadingParser{}
	for _, o := range opts {
		o.setHeadingOption(&p.HeadingConfig)
	}
	return p
}

func (b *atxHeadingParser) Trigger() []byte {
	return []byte{'#'}
}

func (b *atxHeadingParser) Open(_ ast.Node, reader text.Reader, pc Context) (ast.Node, State) {
	line, segment := reader.PeekLine()
	pos := pc.BlockOffset()
	if pos < 0 {
		return nil, NoChildren
	}
	i := pos
	for i < len(line) && line[i] == '#' {
		i++
	}
	level := i - pos
	if i == pos || level > 6 {
		return nil, NoChildren
	}
	if i == len(line) { // alone '#' (without a new line character)
		return ast.NewHeading(level, ast.HeadingKindATX), NoChildren
	}
	l := util.TrimLeftSpaceLength(line[i:])
	if l == 0 {
		return nil, NoChildren
	}

	start := min(i+l, len(line)-1)
	node := ast.NewHeading(level, ast.HeadingKindATX)
	hl := text.NewSegment(
		segment.Start+start-segment.Padding,
		segment.Start+len(line)-segment.Padding)
	hl = hl.TrimRightSpace(reader.Source())
	if hl.Len() == 0 {
		reader.AdvanceToEOL()
		return node, NoChildren
	}

	if b.attribute {
		node.AppendSource(hl)
		parseLastLineAttributes(node, reader, pc)
		hl = node.Source()[0]
		node.SetSource(nil)
	}

	// handle closing sequence of '#' characters
	line = hl.Bytes(reader.Source())
	stop := len(line)
	if stop == 0 { // empty headings like '##[space]'
		stop = 0
	} else {
		i = stop - 1
		for line[i] == '#' && i > 0 {
			i--
		}
		if i == 0 && line[0] == '#' { // empty headings like '### ###'
			reader.AdvanceToEOL()
			return node, NoChildren
		}
		if i != stop-1 && util.IsSpace(line[i]) {
			stop = i
			stop -= util.TrimRightSpaceLength(line[0:stop])
		}
	}
	hl.Stop = hl.Start + stop
	node.AppendSource(hl)
	reader.AdvanceToEOL()

	return node, NoChildren
}

func (b *atxHeadingParser) Continue(_ ast.Node, _ text.Reader, _ Context) State {
	return Close
}

func (b *atxHeadingParser) Close(node ast.Node, reader text.Reader, pc Context) {
	if b.autoHeadingID {
		id, ok := node.Attribute("id")
		if !ok {
			generateAutoHeadingID(node.(*ast.Heading), reader, pc)
		} else {
			pc.IDs().Put(id.Bytes(reader.Source()))
		}
	}
}

func (b *atxHeadingParser) CanInterruptParagraph() bool {
	return true
}

func (b *atxHeadingParser) CanAcceptIndentedLine() bool {
	return false
}

func generateAutoHeadingID(node *ast.Heading, reader text.Reader, pc Context) {
	var line []byte
	lastIndex := len(node.Source()) - 1
	if lastIndex > -1 {
		lastLine := node.Source()[lastIndex]
		line = lastLine.Bytes(reader.Source())
	}
	headingID := pc.IDs().Generate(line, ast.KindHeading)
	node.SetAttribute("id", text.NewStringMultilineValue(string(headingID)))
}

func parseLastLineAttributes(node ast.BlockNode, reader text.Reader, _ Context) {
	lastIndex := len(node.Source()) - 1
	if lastIndex < 0 { // empty headings
		return
	}
	lastLine := node.Source()[lastIndex]
	line := lastLine.Bytes(reader.Source())
	lr := text.NewReader(line)
	var start text.Segment
	var sl int
	for {
		c := lr.Peek()
		if c == text.EOF || c == '\n' {
			break
		}
		if c == '\\' {
			lr.Advance(1)
			if util.IsPunct(lr.Peek()) {
				lr.Advance(1)
			}
			continue
		}
		if c == '{' {
			sl, start = lr.Position()
			attrs, ok := ParseAttributes(lr)
			if ok {
				if nl, _ := lr.PeekLine(); nl == nil || util.IsBlank(nl) {
					for _, attr := range attrs {
						node.SetAttribute(attr.Name, attr.Value)
					}
					lastLine.Stop = lastLine.Start + start.Start
					lastLine = lastLine.TrimRightSpace(reader.Source())
					node.Source()[lastIndex] = lastLine
					return
				}
			}
			lr.SetPosition(sl, start)
		}
		lr.Advance(1)
	}
}
