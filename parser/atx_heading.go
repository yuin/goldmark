package parser

import (
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	"regexp"
)

// A HeadingConfig struct is a data structure that holds configuration of the renderers related to headings.
type HeadingConfig struct {
	AutoHeadingID bool
}

// SetOption implements SetOptioner.
func (b *HeadingConfig) SetOption(name OptionName, value interface{}) {
	switch name {
	case AutoHeadingID:
		b.AutoHeadingID = true
	}
}

// A HeadingOption interface sets options for heading parsers.
type HeadingOption interface {
	SetHeadingOption(*HeadingConfig)
}

// AutoHeadingID is an option name that enables auto IDs for headings.
var AutoHeadingID OptionName = "AutoHeadingID"

type withAutoHeadingID struct {
}

func (o *withAutoHeadingID) SetConfig(c *Config) {
	c.Options[AutoHeadingID] = true
}

func (o *withAutoHeadingID) SetHeadingOption(p *HeadingConfig) {
	p.AutoHeadingID = true
}

// WithAutoHeadingID is a functional option that enables custom heading ids and
// auto generated heading ids.
func WithAutoHeadingID() interface {
	Option
	HeadingOption
} {
	return &withAutoHeadingID{}
}

var atxHeadingRegexp = regexp.MustCompile(`^[ ]{0,3}(#{1,6})(?:\s+(.*?)\s*([\s]#+\s*)?)?\n?$`)

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

	node := ast.NewHeading(level)
	if len(util.TrimRight(line[start:stop], []byte{'#'})) != 0 { // empty heading like '### ###'
		node.Lines().Append(text.NewSegment(segment.Start+start, segment.Start+stop))
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
	generateAutoHeadingID(node.(*ast.Heading), reader, pc)
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
