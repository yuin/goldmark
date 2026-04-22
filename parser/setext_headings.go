package parser

import (
	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/text"
	"github.com/yuin/goldmark/v2/util"
)

var temporaryParagraphKey = NewContextKey()

type setextHeadingParser struct {
	HeadingConfig
}

func matchesSetextHeadingBar(line []byte) (byte, bool) {
	start := 0
	end := len(line)
	space := util.TrimLeftLength(line, []byte{' '})
	if space > 3 {
		return 0, false
	}
	start += space
	level1 := util.TrimLeftLength(line[start:end], []byte{'='})
	c := byte('=')
	var level2 int
	if level1 == 0 {
		level2 = util.TrimLeftLength(line[start:end], []byte{'-'})
		c = '-'
	}
	if util.IsSpace(line[end-1]) {
		end -= util.TrimRightSpaceLength(line[start:end])
	}
	if (level1 <= 0 || start+level1 != end) && (level2 <= 0 || start+level2 != end) {
		return 0, false
	}
	return c, true
}

// NewSetextHeadingParser return a new BlockParser that can parse Setext headings.
func NewSetextHeadingParser(opts ...HeadingOption) BlockParser {
	p := &setextHeadingParser{}
	for _, o := range opts {
		o.setHeadingOption(&p.HeadingConfig)
	}
	return p
}

func (b *setextHeadingParser) Trigger() []byte {
	return []byte{'-', '='}
}

func (b *setextHeadingParser) Open(parent ast.Node, reader text.Reader, pc Context) (ast.Node, State) {
	last := pc.LastOpenedBlock().Node
	if last == nil {
		return nil, NoChildren
	}
	paragraph, ok := last.(*ast.Paragraph)
	if !ok || paragraph.Parent() != parent {
		return nil, NoChildren
	}
	line, segment := reader.PeekLine()
	c, ok := matchesSetextHeadingBar(line)
	if !ok {
		return nil, NoChildren
	}
	level := 1
	if c == '-' {
		level = 2
	}
	node := ast.NewHeading(level, ast.HeadingKindSetext)
	node.AppendSource(segment)
	pc.Set(temporaryParagraphKey, last)
	return node, NoChildren | RequireParagraph
}

func (b *setextHeadingParser) Continue(_ ast.Node, _ text.Reader, _ Context) State {
	return Close
}

func (b *setextHeadingParser) Close(node ast.Node, reader text.Reader, pc Context) {
	heading := node.(*ast.Heading)
	segment := heading.Source()[0]
	heading.SetSource(nil)
	tmp := pc.Get(temporaryParagraphKey).(*ast.Paragraph)
	pc.Set(temporaryParagraphKey, nil)
	if len(tmp.Source()) == 0 {
		next := heading.NextSibling()
		segment = segment.TrimLeftSpace(reader.Source())
		if next == nil || !ast.IsParagraph(next) {
			para := ast.NewParagraph()
			para.AppendSource(segment)
			heading.Parent().InsertAfter(heading, para)
		} else {
			p := next.(*ast.Paragraph)
			p.SetSource(append([]text.Segment{segment}, p.Source()...))
		}
		heading.Parent().RemoveChild(heading)
	} else {
		heading.SetPos(tmp.Source()[0].Start)
		heading.SetSource(tmp.Source())
		heading.SetBlankPreviousLines(tmp.HasBlankPreviousLines())
		tp := tmp.Parent()
		if tp != nil {
			tp.RemoveChild(tmp)
		}
	}

	if b.attribute {
		parseLastLineAttributes(heading, reader, pc)
	}

	if b.autoHeadingID {
		id, ok := node.AttributeString("id")
		if !ok {
			generateAutoHeadingID(heading, reader, pc)
		} else {
			pc.IDs().Put(id.Bytes(nil))
		}
	}
}

func (b *setextHeadingParser) CanInterruptParagraph() bool {
	return true
}

func (b *setextHeadingParser) CanAcceptIndentedLine() bool {
	return false
}
