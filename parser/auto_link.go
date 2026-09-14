package parser

import (
	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/text"
	"github.com/yuin/goldmark/v2/util"
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
	stop := findEmailIndex(line[1:])
	isEmail := true
	if stop < 0 {
		stop = findURLIndex(line[1:])
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

var urlTable = [256]uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 5, 1, 5, 5, 1, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 1, 1, 0, 1, 0, 1, 1, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 1, 1, 1, 1, 1, 1, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1} //nolint:lll

var emailTable = [256]uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 1, 1, 1, 1, 1, 0, 0, 1, 1, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 1, 0, 1, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0} //nolint:lll

// findURLIndex returns a stop index value if the given bytes seem an URL.
// This function is equivalent to [A-Za-z][A-Za-z0-9.+-]{1,31}:[^<>\x00-\x20]* .
func findURLIndex(b []byte) int {
	i := 0
	if len(b) == 0 || urlTable[b[i]]&7 != 7 {
		return -1
	}
	i++
	for ; i < len(b); i++ {
		c := b[i]
		if urlTable[c]&4 != 4 {
			break
		}
	}
	if i == 1 || i > 33 || i >= len(b) {
		return -1
	}
	if b[i] != ':' {
		return -1
	}
	i++
	for ; i < len(b); i++ {
		c := b[i]
		if urlTable[c]&1 != 1 {
			break
		}
	}
	return i
}

func scanDomainLabel(b []byte) int {
	if len(b) == 0 || !util.IsAlphaNumeric(b[0]) {
		return 0
	}
	limit := min(len(b), 63) // 1 + 61 + 1
	i := 1
	for i < limit && (util.IsAlphaNumeric(b[i]) || b[i] == '-') {
		i++
	}
	for i > 1 && b[i-1] == '-' {
		i--
	}
	return i
}

func scanEmailDomain(b []byte) int {
	n := scanDomainLabel(b)
	if n == 0 {
		return 0
	}
	for n < len(b) && b[n] == '.' {
		m := scanDomainLabel(b[n+1:])
		if m == 0 {
			break
		}
		n += 1 + m
	}
	return n
}

// findEmailIndex returns a stop index value if the given bytes seem an email address.
func findEmailIndex(b []byte) int {
	i := 0
	for ; i < len(b); i++ {
		c := b[i]
		if emailTable[c]&1 != 1 {
			break
		}
	}
	if i == 0 {
		return -1
	}
	if i >= len(b) || b[i] != '@' {
		return -1
	}
	i++
	if i >= len(b) {
		return -1
	}
	n := scanEmailDomain(b[i:])
	if n == 0 {
		return -1
	}
	return i + n
}
