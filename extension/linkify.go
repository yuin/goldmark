package extension

import (
	"bytes"
	"regexp"

	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/text"
	"github.com/yuin/goldmark/v2/util"
)

type linkifyConfig struct {
	AllowedProtocols [][]byte
	URLRegexp        *regexp.Regexp
	WWWRegexp        *regexp.Regexp
	EmailRegexp      *regexp.Regexp
}

// LinkifyParserOption is a functional option for the Linkify parser.
type LinkifyParserOption func(*linkifyConfig)

// WithAllowedProtocols is a functional option that specify allowed
// protocols in autolinks. Each protocol must end with ':' like
// 'http:' .
func WithAllowedProtocols[T []byte | string](value []T) LinkifyParserOption {
	return func(p *linkifyConfig) {
		for _, v := range value {
			p.AllowedProtocols = append(p.AllowedProtocols, []byte(v))
		}
	}
}

// WithURLRegexp is a functional option that specify
// a pattern of the URL including a protocol.
func WithURLRegexp(value *regexp.Regexp) LinkifyParserOption {
	return func(p *linkifyConfig) {
		p.URLRegexp = value
	}
}

// WithWWWRegexp is a functional option that specify
// a pattern of the URL without a protocol.
// This pattern must start with 'www.' .
func WithWWWRegexp(value *regexp.Regexp) LinkifyParserOption {
	return func(p *linkifyConfig) {
		p.WWWRegexp = value
	}
}

// WithEmailRegexp is a functional option that specify
// a pattern of the email address.
func WithEmailRegexp(value *regexp.Regexp) LinkifyParserOption {
	return func(p *linkifyConfig) {
		p.EmailRegexp = value
	}
}

type linkifyParser struct {
	linkifyConfig
}

func newLinkifyParser(opts ...LinkifyParserOption) parser.InlineParser {
	p := &linkifyParser{
		linkifyConfig: linkifyConfig{},
	}
	for _, o := range opts {
		o(&p.linkifyConfig)
	}
	return p
}

type linkKind int

const (
	linkKindUnknown linkKind = iota
	linkKindURL
	linkKindWWW
	linkKindEmail
)

func (s *linkifyParser) scan(line []byte, fast bool) (int, linkKind) {
	isEmail := false
	isWWW := false
	stop := -1
	hasProtocol := false
	if s.AllowedProtocols == nil {
		hasProtocol = hasDefaultProtocols(line)
	} else {
		for _, prefix := range s.AllowedProtocols {
			if bytes.HasPrefix(line, prefix) {
				hasProtocol = true
				break
			}
		}
	}
	if hasProtocol {
		if fast {
			return 1, linkKindURL
		}

		if s.URLRegexp == nil {
			stop = scanURL(scanProtocolURL, line)
		} else {
			stop = scanRegexp(s.URLRegexp, line)
		}
	}

	if stop < 0 && bytes.HasPrefix(line, domainWWW) {
		if fast {
			return 1, linkKindWWW
		}
		isWWW = true
		if s.WWWRegexp == nil {
			stop = scanURL(scanWWW, line)
		} else {
			stop = scanRegexp(s.WWWRegexp, line)
		}
	}
	if stop < 0 {
		if len(line) > 0 && util.IsPunct(line[0]) {
			return -1, linkKindUnknown
		}
		isEmail = true
		if s.EmailRegexp == nil {
			stop = parser.ScanEmail(line)
		} else {
			stop = scanRegexp(s.EmailRegexp, line)
		}
		if stop < 0 {
			return -1, linkKindUnknown
		}
		if fast {
			return 1, linkKindEmail
		}
		at := bytes.IndexByte(line, '@')
		m := []int{0, stop, at, stop - 1}
		if bytes.IndexByte(line[m[2]:m[3]], '.') < 0 {
			return -1, linkKindUnknown
		}
		lastChar := line[m[1]-1]
		if lastChar == '.' {
			m[1]--
		}
		if m[1] < len(line) {
			nextChar := line[m[1]]
			if nextChar == '-' || nextChar == '_' {
				return -1, linkKindUnknown
			}
		}
		stop = m[1]
	}
	if stop < 0 {
		return -1, linkKindUnknown
	}
	switch {
	case isEmail:
		return stop, linkKindEmail
	case isWWW:
		return stop, linkKindWWW
	default:
		return stop, linkKindURL
	}
}

// UnsafeScan scans the given line and returns the index of the end of the link if found.
//
// This function is aimed to be used only in goldmark's internal code.
// It may be removed or changed in the future. DO NOT USE.
func (s *linkifyParser) UnsafeScan(line []byte) int {
	index, _ := s.scan(line, true)
	return index
}

func (s *linkifyParser) Trigger() []byte {
	// ' ' indicates any white spaces and a line head
	return []byte{' ', '*', '_', '~', '('}
}

var (
	domainWWW = []byte("www.")
)

func (s *linkifyParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	if pc.IsInLinkLabel() {
		return nil
	}
	line, segment := block.PeekLine()
	consumes := 0
	start := segment.Start
	c := line[0]
	// advance if current position is not a line head.
	if c == ' ' || c == '*' || c == '_' || c == '~' || c == '(' {
		consumes++
		start++
		line = line[1:]
	}

	stop, kind := s.scan(line, false)
	if stop < 0 {
		return nil
	}

	if consumes != 0 {
		s := segment.WithStop(segment.Start + 1)
		parent.AppendChild(ast.NewText(text.NewSingleLineValueFromSegment(s, block.Decoder())))
	}
	opening := bytes.Count(line[:stop], []byte{'('})
	closing := bytes.Count(line[:stop], []byte{')'})
	i := stop
	for i > 1 {
		switch line[i-1] {
		case '?', '!', '.', ',', ':', '*', '_', '~', '\'', '"':
			i--
		case ')':
			if closing <= opening {
				goto endfor
			}
			closing--
			i--
		case ';':
			j := i - 2
			for ; j > 0; j-- {
				if util.IsAlphaNumeric(line[j]) {
					continue
				}
				break
			}
			if j == i-2 || line[j] != '&' {
				goto endfor
			}
			i = j
		default:
			goto endfor
		}
	}
endfor:
	consumes += i
	block.Advance(consumes)
	rawVal := text.NewSingleLineValueFromIndex(text.NewIndex(start, start+i), text.IdentityDecoder)
	var dest text.SingleLineValue
	switch kind {
	case linkKindEmail:
		dest = text.NewSingleLineValueFromString("mailto:"+string(line[:i]), text.IdentityDecoder)
	case linkKindWWW:
		dest = text.NewSingleLineValueFromString("http://"+string(line[:i]), text.IdentityDecoder)
	default:
		dest = rawVal
	}
	link := ast.NewAutoLink(dest, rawVal, ast.WithAutoLinkText(rawVal))
	return link
}

func (s *linkifyParser) CloseBlock(_ ast.Node, _ parser.Context) {
	// nothing to do
}

type linkifyParserExtension struct {
	options []LinkifyParserOption
}

// NewLinkifyParser returns a new parser.Extension that parses text that seems like a URL.
func NewLinkifyParser(opts ...LinkifyParserOption) parser.Extension {
	return &linkifyParserExtension{
		options: opts,
	}
}

func (e *linkifyParserExtension) ParserOptions(_ *parser.Config) []parser.Option {
	return []parser.Option{
		parser.WithInlineParsers(
			util.Prioritized(newLinkifyParser(e.options...), 999),
		),
	}
}

// LinkifyParser is a default [parser.Extension] that parses text that seems like a URL.
var LinkifyParser = NewLinkifyParser()

func hasDefaultProtocols(bs []byte) bool {
	// http:, https:, ftp
	l := len(bs)
	if l > 0 {
		if bs[0] == 'h' && l > 3 && bs[1] == 't' && bs[2] == 't' && bs[3] == 'p' {
			if l > 4 {
				switch bs[4] {
				case ':':
					return true
				case 's':
					if l > 5 && bs[5] == ':' {
						return true
					}
				}
			}
		}
		if bs[0] == 'f' && l > 3 && bs[1] == 't' && bs[2] == 'p' && bs[3] == ':' {
			return true
		}
	}
	return false
}

func scanProtocolURL(line []byte) int {
	i := bytes.IndexByte(line, ':')
	if i < 0 || !bytes.HasPrefix(line[i:], []byte("://")) || parser.ScanDomain(line[i+3:]) == 0 {
		return -1
	}
	return parser.ScanURL(line)
}

func scanWWW(line []byte) int {
	n := parser.ScanDomain(line)
	if n <= len(domainWWW) {
		return -1
	}
	return n
}

func scanURL(scanner func([]byte) int, line []byte) int {
	stop := scanner(line)
	if stop > 0 {
		for i := stop; i < len(line); i++ {
			if util.IsSpace(line[i]) || line[i] == '<' {
				break
			}
			stop++
		}
	}
	return stop
}

func scanRegexp(re *regexp.Regexp, line []byte) int {
	m := re.FindSubmatchIndex(line)
	if m != nil && m[0] == 0 {
		return m[1]
	}
	return -1
}
