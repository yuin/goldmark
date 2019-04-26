package parser

import (
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
	"regexp"
)

type autoLinkParser struct {
}

var defaultAutoLinkParser = &autoLinkParser{}

// NewAutoLinkParser returns a new InlineParser that parses autolinks
// surrounded by '<' and '>' .
func NewAutoLinkParser() InlineParser {
	return defaultAutoLinkParser
}

func (s *autoLinkParser) Trigger() []byte {
	return []byte{'<'}
}

var emailAutoLinkRegexp = regexp.MustCompile(`^<([a-zA-Z0-9.!#$%&'*+\/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*)>`)

var autoLinkRegexp = regexp.MustCompile(`(?i)^<[A-Za-z][A-Za-z0-9.+-]{1,31}:[^<>\x00-\x20]*>`)

func (s *autoLinkParser) Parse(parent ast.Node, block text.Reader, pc Context) ast.Node {
	line, segment := block.PeekLine()
	match := emailAutoLinkRegexp.FindSubmatchIndex(line)
	typ := ast.AutoLinkType(ast.AutoLinkEmail)
	if match == nil {
		match = autoLinkRegexp.FindSubmatchIndex(line)
		typ = ast.AutoLinkURL
	}
	if match == nil {
		return nil
	}
	value := ast.NewTextSegment(text.NewSegment(segment.Start+1, segment.Start+match[1]-1))
	block.Advance(match[1])
	return ast.NewAutoLink(typ, value)
}

func (s *autoLinkParser) CloseBlock(parent ast.Node, pc Context) {
	// nothing to do
}
