package extension

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

// A CJKOption sets options for CJK support mostly for HTML based renderers.
type CJKOption func(*cjk)

// A EastAsianLineBreaksOption sets options for east asian line breaks.
type EastAsianLineBreaksOption func(*eastAsianLineBreaks)

// WithEastAsianLineBreaks is a functional option that indicates whether softline breaks
// between east asian wide characters should be ignored.
func WithEastAsianLineBreaks(opts ...EastAsianLineBreaksOption) CJKOption {
	return func(c *cjk) {
		e := &eastAsianLineBreaks{
			Enabled: true,
		}
		for _, opt := range opts {
			opt(e)
		}
		c.EastAsianLineBreaks = e
	}
}

// WithWorksEvenWithOneSide is a functional option that indicates that a softline break
// is ignored even if only one side of the break is east asian wide character.
func WithWorksEvenWithOneSide() EastAsianLineBreaksOption {
	return func(e *eastAsianLineBreaks) {
		e.WorksEvenWithOneSide = true
	}
}

// WithEscapedSpace is a functional option that indicates that a '\' escaped half-space(0x20) should not be rendered.
func WithEscapedSpace() CJKOption {
	return func(c *cjk) {
		c.EscapedSpace = true
	}
}

type cjk struct {
	EastAsianLineBreaks *eastAsianLineBreaks
	EscapedSpace        bool
}

type eastAsianLineBreaks struct {
	Enabled              bool
	WorksEvenWithOneSide bool
}

// CJK is a goldmark extension that provides functionalities for CJK languages.
var CJK = NewCJK(WithEastAsianLineBreaks(), WithEscapedSpace())

// NewCJK returns a new extension with given options.
func NewCJK(opts ...CJKOption) goldmark.Extender {
	e := &cjk{}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

func (e *cjk) Extend(m goldmark.Markdown) {
	if e.EastAsianLineBreaks != nil {
		if e.EastAsianLineBreaks.Enabled {
			opts := []html.EastAsianLineBreaksOption{}
			if e.EastAsianLineBreaks.WorksEvenWithOneSide {
				opts = append(opts, html.WithWorksEvenWithOneSide())
			}
			m.Renderer().AddOptions(html.WithEastAsianLineBreaks(opts...))
		}

	}
	if e.EscapedSpace {
		m.Renderer().AddOptions(html.WithWriter(html.NewWriter(html.WithEscapedSpace())))
		m.Parser().AddOptions(parser.WithEscapedSpace())
	}
}
