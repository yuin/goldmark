package extension

import (
	"bytes"
	"regexp"

	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/text"
	"github.com/yuin/goldmark/v2/util"
)

var wwwURLRegxp = regexp.MustCompile(`^www\.[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-z]+(?:[/#?][-a-zA-Z0-9@:%_\+.~#!?&/=\(\);,'">\^{}\[\]` + "`" + `]*)?`) //nolint:lll

var urlRegexp = regexp.MustCompile(`^(?:http|https|ftp)://[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-z]+(?::\d+)?(?:[/#?][-a-zA-Z0-9@:%_+.~#$!?&/=\(\);,'">\^{}\[\]` + "`" + `]*)?`) //nolint:lll

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
		linkifyConfig: linkifyConfig{
			URLRegexp: urlRegexp,
			WWWRegexp: wwwURLRegxp,
		},
	}
	for _, o := range opts {
		o(&p.linkifyConfig)
	}
	return p
}

func (s *linkifyParser) Trigger() []byte {
	// ' ' indicates any white spaces and a line head
	return []byte{' ', '*', '_', '~', '('}
}

var (
	protoHTTP  = []byte("http:")
	protoHTTPS = []byte("https:")
	protoFTP   = []byte("ftp:")
	domainWWW  = []byte("www.")
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

	var m []int
	isEmail := false
	isWWW := false
	if s.AllowedProtocols == nil {
		if bytes.HasPrefix(line, protoHTTP) || bytes.HasPrefix(line, protoHTTPS) || bytes.HasPrefix(line, protoFTP) {
			m = s.URLRegexp.FindSubmatchIndex(line)
		}
	} else {
		for _, prefix := range s.AllowedProtocols {
			if bytes.HasPrefix(line, prefix) {
				m = s.URLRegexp.FindSubmatchIndex(line)
				break
			}
		}
	}
	if m == nil && bytes.HasPrefix(line, domainWWW) {
		m = s.WWWRegexp.FindSubmatchIndex(line)
		isWWW = true
	}
	if m != nil && m[0] != 0 {
		m = nil
	}
	if m != nil && m[0] == 0 {
		lastChar := line[m[1]-1]
		if lastChar == '.' {
			m[1]--
		} else if lastChar == ')' {
			closing := 0
			for i := m[1] - 1; i >= m[0]; i-- {
				switch line[i] {
				case ')':
					closing++
				case '(':
					closing--
				}
			}
			if closing > 0 {
				m[1] -= closing
			}
		} else if lastChar == ';' {
			i := m[1] - 2
			for ; i >= m[0]; i-- {
				if util.IsAlphaNumeric(line[i]) {
					continue
				}
				break
			}
			if i != m[1]-2 {
				if line[i] == '&' {
					m[1] -= m[1] - i
				}
			}
		}
	}
	if m == nil {
		if len(line) > 0 && util.IsPunct(line[0]) {
			return nil
		}
		isEmail = true
		stop := -1
		if s.EmailRegexp == nil {
			stop = findEmailIndex(line)
		} else {
			m := s.EmailRegexp.FindSubmatchIndex(line)
			if m != nil && m[0] == 0 {
				stop = m[1]
			}
		}
		if stop < 0 {
			return nil
		}
		at := bytes.IndexByte(line, '@')
		m = []int{0, stop, at, stop - 1}
		if bytes.IndexByte(line[m[2]:m[3]], '.') < 0 {
			return nil
		}
		lastChar := line[m[1]-1]
		if lastChar == '.' {
			m[1]--
		}
		if m[1] < len(line) {
			nextChar := line[m[1]]
			if nextChar == '-' || nextChar == '_' {
				return nil
			}
		}
	}
	if m == nil {
		return nil
	}
	if consumes != 0 {
		s := segment.WithStop(segment.Start + 1)
		parent.AppendChild(ast.NewText(text.NewSingleLineValueFromSegment(s, block.Decoder())))
	}
	i := m[1] - 1
	for ; i > 0; i-- {
		c := line[i]
		switch c {
		case '?', '!', '.', ',', ':', '*', '_', '~':
		default:
			goto endfor
		}
	}
endfor:
	i++
	consumes += i
	block.Advance(consumes)
	rawVal := text.NewSingleLineValueFromIndex(text.NewIndex(start, start+i), text.IdentityDecoder)
	var dest text.SingleLineValue
	switch {
	case isEmail:
		dest = text.NewSingleLineValueFromString("mailto:"+string(line[:i]), text.IdentityDecoder)
	case isWWW:
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

var emailTable = [256]uint8{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 1, 1, 1, 1, 1, 0, 0, 1, 1, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 1, 0, 1, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0} //nolint:lll

var emailDomainRegexp = regexp.MustCompile(`^[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*`) //nolint:lll

// findEmailIndex returns a stop index value if the given bytes seem an email address.
func findEmailIndex(b []byte) int {
	// TODO: eliminate regexps
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
	match := emailDomainRegexp.FindSubmatchIndex(b[i:])
	if match == nil {
		return -1
	}
	return i + match[1]
}
