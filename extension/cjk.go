package extension

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

// A CJKOption sets options for CJK support mostly for HTML based renderers.
type CJKOption func(*cjk)

// A EastAsianLineBreaksStyle is a style of east asian line breaks.
type EastAsianLineBreaksStyle int

const (
	// EastAsianLineBreaksStyleSimple is a style where soft line breaks are ignored
	// if both sides of the break are east asian wide characters.
	EastAsianLineBreaksStyleSimple EastAsianLineBreaksStyle = iota
	// EastAsianLineBreaksCSS3Draft is a style where soft line breaks are ignored
	// even if only one side of the break is an east asian wide character.
	EastAsianLineBreaksCSS3Draft
)

// WithEastAsianLineBreaks is a functional option that indicates whether softline breaks
// between east asian wide characters should be ignored.
func WithEastAsianLineBreaks(style ...EastAsianLineBreaksStyle) CJKOption {
	return func(c *cjk) {
		e := &eastAsianLineBreaks{
			Enabled:                  true,
			EastAsianLineBreaksStyle: EastAsianLineBreaksStyleSimple,
		}
		for _, s := range style {
			e.EastAsianLineBreaksStyle = s
		}
		c.EastAsianLineBreaks = e
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
	Enabled                  bool
	EastAsianLineBreaksStyle EastAsianLineBreaksStyle
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
			style := html.EastAsianLineBreaksStyleSimple
			switch e.EastAsianLineBreaks.EastAsianLineBreaksStyle {
			case EastAsianLineBreaksCSS3Draft:
				style = html.EastAsianLineBreaksCSS3Draft
			}
			m.Renderer().AddOptions(html.WithEastAsianLineBreaks(style))
		}
	}
	if e.EscapedSpace {
		m.Renderer().AddOptions(html.WithWriter(html.NewWriter(html.WithEscapedSpace())))
		m.Parser().AddOptions(parser.WithEscapedSpace())
	}
}
