package parser

import (
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// A HeadingConfig struct is a data structure that holds configuration of the renderers related to headings.
type HeadingConfig struct {
	AutoHeadingID bool
	Attribute     bool
}

// SetOption implements SetOptioner.
func (b *HeadingConfig) SetOption(name OptionName, value interface{}) {
	switch name {
	case optAutoHeadingID:
		b.AutoHeadingID = true
	case optAttribute:
		b.Attribute = true
	}
}

// A HeadingOption interface sets options for heading parsers.
type HeadingOption interface {
	Option
	SetHeadingOption(*HeadingConfig)
}

// AutoHeadingID is an option name that enables auto IDs for headings.
const optAutoHeadingID OptionName = "AutoHeadingID"

type withAutoHeadingID struct {
}

func (o *withAutoHeadingID) SetParserOption(c *Config) {
	c.Options[optAutoHeadingID] = true
}

func (o *withAutoHeadingID) SetHeadingOption(p *HeadingConfig) {
	p.AutoHeadingID = true
}

// WithAutoHeadingID is a functional option that enables custom heading ids and
// auto generated heading ids.
func WithAutoHeadingID() HeadingOption {
	return &withAutoHeadingID{}
}

type withHeadingAttribute struct {
	Option
}

func (o *withHeadingAttribute) SetHeadingOption(p *HeadingConfig) {
	p.Attribute = true
}

// WithHeadingAttribute is a functional option that enables custom heading attributes.
func WithHeadingAttribute() HeadingOption {
	return &withHeadingAttribute{WithAttribute()}
}

type atxHeadingParser struct {
	HeadingConfig
}

// NewATXHeadingParser return a new BlockParser that can parse ATX headings.
func NewATXHeadingParser(opts ...HeadingOption) BlockParser {
	p := &atxHeadingParser{}
	for _, o := range opts {
		o.SetHeadingOption(&p.HeadingConfig)
	}
	return p
}

func (b *atxHeadingParser) Open(parent ast.Node, reader text.Reader, pc Context) (ast.Node, State) {
	line, segment := reader.PeekLine()
	pos := pc.BlockOffset()
	i := pos
	for ; i < len(line) && line[i] == '#'; i++ {
	}
	level := i - pos
	if i == pos || level > 6 {
		return nil, NoChildren
	}
	l := util.TrimLeftSpaceLength(line[i:])
	if l == 0 {
		return nil, NoChildren
	}
	start := i + l
	stop := len(line) - util.TrimRightSpaceLength(line)

	node := ast.NewHeading(level)
	parsed := false
	if b.Attribute { // handles special case like ### heading ### {#id}
		start--
		closureOpen := -1
		closureClose := -1
		for i := start; i < stop; {
			c := line[i]
			if util.IsEscapedPunctuation(line, i) {
				i += 2
			} else if util.IsSpace(c) && i < stop-1 && line[i+1] == '#' {
				closureOpen = i + 1
				j := i + 1
				for ; j < stop && line[j] == '#'; j++ {
				}
				closureClose = j
				break
			} else {
				i++
			}
		}
		if closureClose > 0 {
			i := closureClose
			for ; i < stop && util.IsSpace(line[i]); i++ {
			}
			if i < stop-1 || line[i] == '{' {
				as := i + 1
				for as < stop {
					ai := util.FindAttributeIndex(line[as:], true)
					if ai[0] < 0 {
						break
					}
					node.SetAttribute(line[as+ai[0]:as+ai[1]],
						line[as+ai[2]:as+ai[3]])
					as += ai[3]
				}
				if line[as] == '}' && (as > stop-2 || util.IsBlank(line[as:])) {
					parsed = true
					node.Lines().Append(text.NewSegment(segment.Start+start+1, segment.Start+closureOpen))
				} else {
					node.RemoveAttributes()
				}
			}
		}
	}
	if !parsed {
		stop := len(line) - util.TrimRightSpaceLength(line)
		if stop <= start { // empty headings like '##[space]'
			stop = start + 1
		} else {
			i = stop - 1
			for ; line[i] == '#' && i >= start; i-- {
			}
			if i != stop-1 && !util.IsSpace(line[i]) {
				i = stop - 1
			}
			i++
			stop = i
		}

		if len(util.TrimRight(line[start:stop], []byte{'#'})) != 0 { // empty heading like '### ###'
			node.Lines().Append(text.NewSegment(segment.Start+start, segment.Start+stop))
		}
	}
	return node, NoChildren
}

func (b *atxHeadingParser) Continue(node ast.Node, reader text.Reader, pc Context) State {
	return Close
}

func (b *atxHeadingParser) Close(node ast.Node, reader text.Reader, pc Context) {
	if !b.AutoHeadingID {
		return
	}
	if !b.Attribute {
		_, ok := node.AttributeString("id")
		if !ok {
			generateAutoHeadingID(node.(*ast.Heading), reader, pc)
		}
	}
}

func (b *atxHeadingParser) CanInterruptParagraph() bool {
	return true
}

func (b *atxHeadingParser) CanAcceptIndentedLine() bool {
	return false
}

var attrAutoHeadingIDPrefix = []byte("heading")
var attrNameID = []byte("#")

func generateAutoHeadingID(node *ast.Heading, reader text.Reader, pc Context) {
	lastIndex := node.Lines().Len() - 1
	lastLine := node.Lines().At(lastIndex)
	line := lastLine.Value(reader.Source())
	headingID := pc.IDs().Generate(line, attrAutoHeadingIDPrefix)
	node.SetAttribute(attrNameID, headingID)
}
