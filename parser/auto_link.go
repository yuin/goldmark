package parser

import (
	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/text"
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

func (s *autoLinkParser) Parse(_ ast.Node, block text.Reader, _ Context) ast.Node {
	line, segment := block.PeekLine()
	stop := ScanEmail(line[1:])
	isEmail := true
	if stop < 0 {
		stop = ScanURL(line[1:])
		isEmail = false
	}
	if stop < 0 {
		return nil
	}
	stop++
	if stop >= len(line) || line[stop] != '>' {
		return nil
	}
	textVal := text.NewSingleLineValueFromIndex(text.NewIndex(segment.Start, segment.Start+stop+1), text.IdentityDecoder)
	labelVal := text.NewSingleLineValueFromIndex(text.NewIndex(segment.Start+1, segment.Start+stop), text.IdentityDecoder)
	block.Advance(stop + 1)
	var dest text.SingleLineValue
	if isEmail {
		dest = text.NewSingleLineValueFromString("mailto:"+string(line[1:stop]), text.IdentityDecoder)
	} else {
		dest = text.NewSingleLineValueFromIndex(text.NewIndex(segment.Start+1, segment.Start+stop), text.IdentityDecoder)
	}
	n := ast.NewAutoLink(dest, labelVal, ast.WithAutoLinkText(textVal))
	return n
}
