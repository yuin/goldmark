// Package html implements renderer that outputs HTMLs.
package html

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/renderer"
	"github.com/yuin/goldmark/v2/util"
)

// Type aliases {{{

// Renderer is a renderer that renders AST nodes as (X)HTML.
type Renderer = renderer.Renderer[io.Writer]

// NodeRenderer is a renderer that renders AST nodes as (X)HTML.
type NodeRenderer = renderer.NodeRenderer[io.Writer]

// NodeRendererFunc is a function that renders AST nodes as (X)HTML.
func NodeRendererFunc(f func(w io.Writer, source []byte,
	n ast.Node, entering bool, rctx renderer.Context) (ast.WalkStatus, error)) NodeRenderer {
	return renderer.NodeRendererFunc(f)
}

// Extension is an extension that extends HTML based renderers.
type Extension = renderer.Extension[Config]

// Option is a functional option for HTML based renderers.
type Option = renderer.Option[Config]

// Hook is a hook that hooks before rendering a node.
type Hook = renderer.Hook[io.Writer]

// EmptyHook is a Hook that does nothing.
type EmptyHook = renderer.EmptyHook[io.Writer]

// WithNodeRenderers sets a node renderer for the given node kind.
func WithNodeRenderers(nodeRenderers map[ast.NodeKind]NodeRenderer) Option {
	return renderer.WithNodeRenderers[io.Writer, Config](nodeRenderers)
}

// WithNodeRenderer sets a node renderer for the given node kind.
func WithNodeRenderer(kind ast.NodeKind, nodeRenderer NodeRenderer) Option {
	return renderer.WithNodeRenderer[io.Writer, Config](kind, nodeRenderer)
}

// WithExtensions adds extensions.
func WithExtensions(extensions ...Extension) Option {
	return renderer.WithExtensions[io.Writer](extensions...)
}

// WithHooks adds render hooks.
func WithHooks(hooks ...Hook) Option {
	return renderer.WithHooks[io.Writer, Config](hooks...)
}

// }}} Type aliases

var writerKey = renderer.NewContextKey()

// ContextWriter returns the Writer in the renderer.Context.
func ContextWriter(c renderer.Context) Writer {
	return c.Get(writerKey).(Writer)
}

// IsInTightBlockFunc determines whether a paragraph node is inside a tight block
// and should be rendered without <p> tags.
type IsInTightBlockFunc func(n ast.Node) bool

// IsInTightBlock returns true if the paragraph is a direct child of a
// ListItem whose parent List is tight.
func IsInTightBlock(n ast.Node) bool {
	parent := n.Parent()
	if parent == nil {
		return false
	}
	gp := parent.Parent()
	if gp == nil {
		return false
	}
	if list, ok := gp.(*ast.List); ok {
		return list.IsTight
	}
	return false
}

// A ParagraphConfig struct has configurations for paragraph rendering.
type ParagraphConfig struct {
	// IsInTightBlockFunc determines whether a paragraph should be rendered
	// without <p> tags because it is inside a tight block.
	IsInTightBlockFunc IsInTightBlockFunc
}

// A Config struct has configurations for the HTML based renderers.
type Config struct {
	renderer.Config[io.Writer, Config]
	EscapedSpace        bool
	HardWraps           bool
	EastAsianLineBreaks EastAsianLineBreaks
	XHTML               bool
	Unsafe              bool
	Paragraph           ParagraphConfig
}

// Default returns a Config with default values.
func (c Config) Default() Config {
	return Config{
		HardWraps:           false,
		EastAsianLineBreaks: EastAsianLineBreaksNone,
		XHTML:               false,
		Unsafe:              false,
		Paragraph: ParagraphConfig{
			IsInTightBlockFunc: IsInTightBlock,
		},
	}
}

// WithHardWraps is a functional option that indicates whether softline breaks
// should be rendered as '<br>'.
func WithHardWraps() Option {
	return renderer.NewOptionFunc(func(c *Config) {
		c.HardWraps = true
	})
}

// A EastAsianLineBreaks is a style of east asian line breaks.
type EastAsianLineBreaks int

const (
	//EastAsianLineBreaksNone renders line breaks as it is.
	EastAsianLineBreaksNone EastAsianLineBreaks = iota

	// EastAsianLineBreaksSimple follows east_asian_line_breaks in Pandoc.
	EastAsianLineBreaksSimple
	// EastAsianLineBreaksCSS3Draft follows CSS text level3 "Segment Break Transformation Rules" with some enhancements.
	EastAsianLineBreaksCSS3Draft
)

func (b EastAsianLineBreaks) softLineBreak(thisLastRune rune, siblingFirstRune rune) bool {
	switch b {
	case EastAsianLineBreaksNone:
		return false
	case EastAsianLineBreaksSimple:
		return !util.IsEastAsianWideRune(thisLastRune) || !util.IsEastAsianWideRune(siblingFirstRune)
	case EastAsianLineBreaksCSS3Draft:
		return eastAsianLineBreaksCSS3DraftSoftLineBreak(thisLastRune, siblingFirstRune)
	}
	return false
}

func eastAsianLineBreaksCSS3DraftSoftLineBreak(thisLastRune rune, siblingFirstRune rune) bool {
	// Implements CSS text level3 Segment Break Transformation Rules with some enhancements.
	// References:
	//   - https://www.w3.org/TR/2020/WD-css-text-3-20200429/#line-break-transform
	//   - https://github.com/w3c/csswg-drafts/issues/5086

	// Rule1:
	//   If the character immediately before or immediately after the segment break is
	//   the zero-width space character (U+200B), then the break is removed, leaving behind the zero-width space.
	if thisLastRune == '\u200B' || siblingFirstRune == '\u200B' {
		return false
	}

	// Rule2:
	//   Otherwise, if the East Asian Width property of both the character before and after the segment break is
	//   F, W, or H (not A), and neither side is Hangul, then the segment break is removed.
	thisLastRuneEastAsianWidth := util.EastAsianWidth(thisLastRune)
	siblingFirstRuneEastAsianWidth := util.EastAsianWidth(siblingFirstRune)
	if (thisLastRuneEastAsianWidth == "F" ||
		thisLastRuneEastAsianWidth == "W" ||
		thisLastRuneEastAsianWidth == "H") &&
		(siblingFirstRuneEastAsianWidth == "F" ||
			siblingFirstRuneEastAsianWidth == "W" ||
			siblingFirstRuneEastAsianWidth == "H") {
		return unicode.Is(unicode.Hangul, thisLastRune) || unicode.Is(unicode.Hangul, siblingFirstRune)
	}

	// Rule3:
	//   Otherwise, if either the character before or after the segment break belongs to
	//   the space-discarding character set and it is a Unicode Punctuation (P*) or U+3000,
	//   then the segment break is removed.
	if util.IsSpaceDiscardingUnicodeRune(thisLastRune) ||
		unicode.IsPunct(thisLastRune) ||
		thisLastRune == '\u3000' ||
		util.IsSpaceDiscardingUnicodeRune(siblingFirstRune) ||
		unicode.IsPunct(siblingFirstRune) ||
		siblingFirstRune == '\u3000' {
		return false
	}

	// Rule4:
	//   Otherwise, the segment break is converted to a space (U+0020).
	return true
}

// WithEastAsianLineBreaks is a functional option that indicates whether softline breaks
// between east asian wide characters should be ignored.
func WithEastAsianLineBreaks(e EastAsianLineBreaks) Option {
	return renderer.NewOptionFunc(func(c *Config) {
		c.EastAsianLineBreaks = e
	})
}

// WithXHTML is a functional option indicates that nodes should be rendered in
// xhtml instead of HTML5.
func WithXHTML() Option {
	return renderer.NewOptionFunc(func(c *Config) {
		c.XHTML = true
	})
}

// WithUnsafe is a functional option that renders dangerous contents
// (raw htmls and potentially dangerous links) as it is.
func WithUnsafe() Option {
	return renderer.NewOptionFunc(func(c *Config) {
		c.Unsafe = true
	})
}

// WithIsInTightBlock is a functional option that sets a custom function to
// determine whether a paragraph node is inside a tight block and should be
// rendered without <p> tags.
func WithIsInTightBlock(fn IsInTightBlockFunc) Option {
	return renderer.NewOptionFunc(func(c *Config) {
		c.Paragraph.IsInTightBlockFunc = fn
	})
}

type htmlRenderer struct {
	*renderer.Helper[io.Writer, Config]
}

type writerHook struct {
	EmptyHook
	cm *commonMark
}

func (h *writerHook) PreRender(_ io.Writer, _ []byte, _ ast.Node, rctx renderer.Context) error {
	rctx.Set(writerKey, h.cm.writer)
	return nil
}

// New returns a new Renderer with given options.
func New(opts ...Option) Renderer {
	cm := NewCommonMark()
	opts = append([]Option{WithHooks(&writerHook{cm: cm.(*commonMark)}), WithExtensions(cm)}, opts...)
	r := &htmlRenderer{
		Helper: renderer.NewHelper[io.Writer](opts...),
	}
	return r
}

// Render renders the given AST node to the given writer.
func (r *htmlRenderer) Render(w io.Writer, source []byte, n ast.Node) error {
	switch v := w.(type) {
	case util.ErrorBufWriter:
		return r.Helper.Render(v, source, n)
	case *bytes.Buffer, *strings.Builder:
		return r.Helper.Render(v, source, n)
	}

	return r.Helper.Render(util.NewErrorBufWriterSize(w, len(source)*3), source, n)
}

type commonMark struct {
	opts   []Option
	config *Config
	writer Writer
}

// NewCommonMark returns a new Extension that renders CommonMark compliant HTML.
func NewCommonMark(opts ...Option) Extension {
	return &commonMark{opts: opts}
}

func (e *commonMark) RendererOptions(cfg *Config) []Option {
	if len(e.opts) != 0 {
		thisConfig := *cfg
		for _, opt := range e.opts {
			opt.SetFormatOption(&thisConfig)
		}
		cfg = &thisConfig
	}
	e.config = cfg
	noop := NodeRendererFunc(func(
		_ io.Writer, _ []byte, _ ast.Node, _ bool, _ renderer.Context) (ast.WalkStatus, error) {
		return ast.WalkSkipChildren, nil
	})
	{
		var opts []WriterOption
		if e.config.EscapedSpace {
			opts = append(opts, WithEscapedSpace())
		}
		e.writer = NewWriter(opts...)
	}

	return []Option{
		WithNodeRenderers(map[ast.NodeKind]NodeRenderer{
			ast.KindDocument:                NodeRendererFunc(e.renderDocument),
			ast.KindHeading:                 NodeRendererFunc(e.renderHeading),
			ast.KindBlockquote:              NodeRendererFunc(e.renderBlockquote),
			ast.KindCodeBlock:               NodeRendererFunc(e.renderCodeBlock),
			ast.KindHTMLBlock:               NodeRendererFunc(e.renderHTMLBlock),
			ast.KindList:                    NodeRendererFunc(e.renderList),
			ast.KindListItem:                NodeRendererFunc(e.renderListItem),
			ast.KindParagraph:               NodeRendererFunc(e.renderParagraph),
			ast.KindThematicBreak:           NodeRendererFunc(e.renderThematicBreak),
			ast.KindLinkReferenceDefinition: noop,
			ast.KindAutoLink:                NodeRendererFunc(e.renderAutoLink),
			ast.KindCodeSpan:                NodeRendererFunc(e.renderCodeSpan),
			ast.KindEmphasis:                NodeRendererFunc(e.renderEmphasis),
			ast.KindStrong:                  NodeRendererFunc(e.renderStrong),
			ast.KindImage:                   NodeRendererFunc(e.renderImage),
			ast.KindLink:                    NodeRendererFunc(e.renderLink),
			ast.KindRawHTML:                 NodeRendererFunc(e.renderRawHTML),
			ast.KindText:                    NodeRendererFunc(e.renderText),
		},
		),
	}
}

// GlobalAttributeFilter defines attribute names which any elements can have.
var GlobalAttributeFilter = util.NewBytesFilterString(`accesskey,autocapitalize,autofocus,class,contenteditable,dir,draggable,enterkeyhint,hidden,id,inert,inputmode,is,itemid,itemprop,itemref,itemscope,itemtype,lang,part,role,slot,spellcheck,style,tabindex,title,translate`) // nolint:lll

func (e *commonMark) renderDocument(
	_ io.Writer, _ []byte, _ ast.Node, _ bool, _ renderer.Context) (ast.WalkStatus, error) {
	return ast.WalkContinue, nil
}

// HeadingAttributeFilter defines attribute names which heading elements can have.
var HeadingAttributeFilter = GlobalAttributeFilter

func (e *commonMark) renderHeading(
	writer io.Writer, _ []byte, node ast.Node, entering bool, _ renderer.Context) (ast.WalkStatus, error) {
	w := writer.(util.BufWriter)
	n := node.(*ast.Heading)
	if entering {
		_, _ = w.WriteString("<h")
		_ = w.WriteByte("0123456"[n.Level])
		if n.Attributes() != nil {
			RenderAttributes(w, node, HeadingAttributeFilter)
		}
		_ = w.WriteByte('>')
	} else {
		_, _ = w.WriteString("</h")
		_ = w.WriteByte("0123456"[n.Level])
		_, _ = w.WriteString(">\n")
	}
	return ast.WalkContinue, nil
}

// BlockquoteAttributeFilter defines attribute names which blockquote elements can have.
var BlockquoteAttributeFilter = GlobalAttributeFilter.ExtendString(`cite`)

func (e *commonMark) renderBlockquote(
	writer io.Writer, _ []byte, n ast.Node, entering bool, _ renderer.Context) (ast.WalkStatus, error) {
	w := writer.(util.BufWriter)
	if entering {
		if n.Attributes() != nil {
			_, _ = w.WriteString("<blockquote")
			RenderAttributes(w, n, BlockquoteAttributeFilter)
			_ = w.WriteByte('>')
		} else {
			_, _ = w.WriteString("<blockquote>\n")
		}
	} else {
		_, _ = w.WriteString("</blockquote>\n")
	}
	return ast.WalkContinue, nil
}

func (e *commonMark) renderCodeBlock(
	writer io.Writer, source []byte, node ast.Node, entering bool, _ renderer.Context) (ast.WalkStatus, error) {
	w := writer.(util.BufWriter)
	n := node.(*ast.CodeBlock)
	if entering {
		_, _ = w.WriteString("<pre><code")
		language, ok := n.Language(source)
		if ok {
			_, _ = w.WriteString(" class=\"language-")
			e.writer.WriteText(w, language.Bytes(source))
			_ = w.WriteByte('"')
		}
		_ = w.WriteByte('>')
		e.writer.RawWriteText(w, n.Value.Bytes(source))
	} else {
		_, _ = w.WriteString("</code></pre>\n")
	}
	return ast.WalkContinue, nil
}

func (e *commonMark) renderHTMLBlock(
	writer io.Writer, source []byte, node ast.Node, entering bool, _ renderer.Context) (ast.WalkStatus, error) {
	w := writer.(util.BufWriter)
	n := node.(*ast.HTMLBlock)
	if entering {
		if e.config.Unsafe {
			e.writer.WriteHTML(w, n.Value.Bytes(source))
		} else {
			_, _ = w.WriteString("<!-- raw HTML omitted -->\n")
		}
	}
	return ast.WalkContinue, nil
}

// ListAttributeFilter defines attribute names which list elements can have.
var ListAttributeFilter = GlobalAttributeFilter.ExtendString(`start,reversed,type`)

func (e *commonMark) renderList(
	writer io.Writer, _ []byte, node ast.Node, entering bool, _ renderer.Context) (ast.WalkStatus, error) {
	w := writer.(util.BufWriter)
	n := node.(*ast.List)
	tag := "ul"
	if n.IsOrdered() {
		tag = "ol"
	}
	if entering {
		_ = w.WriteByte('<')
		_, _ = w.WriteString(tag)
		if n.IsOrdered() && n.Start != 1 {
			_, _ = fmt.Fprintf(w, " start=\"%d\"", n.Start)
		}
		if n.Attributes() != nil {
			RenderAttributes(w, n, ListAttributeFilter)
		}
		_, _ = w.WriteString(">\n")
	} else {
		_, _ = w.WriteString("</")
		_, _ = w.WriteString(tag)
		_, _ = w.WriteString(">\n")
	}
	return ast.WalkContinue, nil
}

// ListItemAttributeFilter defines attribute names which list item elements can have.
var ListItemAttributeFilter = GlobalAttributeFilter.ExtendString(`value`)

func (e *commonMark) renderListItem(
	writer io.Writer, _ []byte, n ast.Node, entering bool, _ renderer.Context) (ast.WalkStatus, error) {
	w := writer.(util.BufWriter)
	if entering {
		if n.Attributes() != nil {
			_, _ = w.WriteString("<li")
			RenderAttributes(w, n, ListItemAttributeFilter)
			_ = w.WriteByte('>')
		} else {
			_, _ = w.WriteString("<li>")
		}
		fc := n.FirstChild()
		if fc != nil {
			if paragraph, ok := fc.(*ast.Paragraph); !ok || !e.config.Paragraph.IsInTightBlockFunc(paragraph) {
				_ = w.WriteByte('\n')
			}
		}
	} else {
		_, _ = w.WriteString("</li>\n")
	}
	return ast.WalkContinue, nil
}

// ParagraphAttributeFilter defines attribute names which paragraph elements can have.
var ParagraphAttributeFilter = GlobalAttributeFilter

func (e *commonMark) renderParagraph(
	writer io.Writer, _ []byte, n ast.Node, entering bool, _ renderer.Context) (ast.WalkStatus, error) {
	w := writer.(util.BufWriter)
	if e.config.Paragraph.IsInTightBlockFunc(n) {
		if !entering && n.NextSibling() != nil && n.FirstChild() != nil {
			_ = w.WriteByte('\n')
		}
		return ast.WalkContinue, nil
	}
	if entering {
		if n.Attributes() != nil {
			_, _ = w.WriteString("<p")
			RenderAttributes(w, n, ParagraphAttributeFilter)
			_ = w.WriteByte('>')
		} else {
			_, _ = w.WriteString("<p>")
		}
	} else {
		_, _ = w.WriteString("</p>\n")
	}
	return ast.WalkContinue, nil
}

// ThematicAttributeFilter defines attribute names which hr elements can have.
var ThematicAttributeFilter = GlobalAttributeFilter.ExtendString(`align,color,noshade,size,width`)

func (e *commonMark) renderThematicBreak(
	writer io.Writer, _ []byte, n ast.Node, entering bool, _ renderer.Context) (ast.WalkStatus, error) {
	w := writer.(util.BufWriter)
	if !entering {
		return ast.WalkContinue, nil
	}
	_, _ = w.WriteString("<hr")
	if n.Attributes() != nil {
		RenderAttributes(w, n, ThematicAttributeFilter)
	}
	if e.config.XHTML {
		_, _ = w.WriteString(" />\n")
	} else {
		_, _ = w.WriteString(">\n")
	}
	return ast.WalkContinue, nil
}

// LinkAttributeFilter defines attribute names which link elements can have.
var LinkAttributeFilter = GlobalAttributeFilter.ExtendString(`download,href,lang,media,ping,referrerpolicy,rel,shape,target`) // nolint:lll

func (e *commonMark) renderAutoLink(
	writer io.Writer, source []byte, node ast.Node, entering bool, _ renderer.Context) (ast.WalkStatus, error) {
	w := writer.(util.BufWriter)
	n := node.(*ast.AutoLink)
	if !entering {
		return ast.WalkContinue, nil
	}
	_, _ = w.WriteString(`<a href="`)
	url := util.URLEscape(n.Destination.Bytes(source), false)
	label := n.Label.Bytes(source)
	if e.config.Unsafe || !IsDangerousURL(url) {
		_, _ = w.Write(util.EscapeHTML(url))
	}
	if n.Attributes() != nil {
		_ = w.WriteByte('"')
		RenderAttributes(w, n, LinkAttributeFilter)
		_ = w.WriteByte('>')
	} else {
		_, _ = w.WriteString(`">`)
	}
	_, _ = w.Write(util.EscapeHTML(label))
	_, _ = w.WriteString(`</a>`)
	return ast.WalkContinue, nil
}

// CodeAttributeFilter defines attribute names which code elements can have.
var CodeAttributeFilter = GlobalAttributeFilter

func (e *commonMark) renderCodeSpan(
	writer io.Writer, source []byte, n ast.Node, entering bool, _ renderer.Context) (ast.WalkStatus, error) {
	w := writer.(util.BufWriter)
	if entering {
		if n.Attributes() != nil {
			_, _ = w.WriteString("<code")
			RenderAttributes(w, n, CodeAttributeFilter)
			_ = w.WriteByte('>')
		} else {
			_, _ = w.WriteString("<code>")
		}
		value := n.(*ast.CodeSpan).Value.Bytes(source)
		// CommonMark spec: line endings within code spans are treated as spaces.
		value = bytes.ReplaceAll(value, []byte{'\n'}, []byte{' '})
		e.writer.RawWriteText(w, value)
		return ast.WalkSkipChildren, nil
	}
	_, _ = w.WriteString("</code>")
	return ast.WalkContinue, nil
}

// EmphasisAttributeFilter defines attribute names which emphasis elements can have.
var EmphasisAttributeFilter = GlobalAttributeFilter

func (e *commonMark) renderEmphasis(
	writer io.Writer, _ []byte, node ast.Node, entering bool, _ renderer.Context) (ast.WalkStatus, error) {
	w := writer.(util.BufWriter)
	if entering {
		_ = w.WriteByte('<')
		_, _ = w.WriteString("em")
		if node.Attributes() != nil {
			RenderAttributes(w, node, EmphasisAttributeFilter)
		}
		_ = w.WriteByte('>')
	} else {
		_, _ = w.WriteString("</em>")
	}
	return ast.WalkContinue, nil
}

// StrongAttributeFilter defines attribute names which strong elements can have.
var StrongAttributeFilter = GlobalAttributeFilter

func (e *commonMark) renderStrong(
	writer io.Writer, _ []byte, node ast.Node, entering bool, _ renderer.Context) (ast.WalkStatus, error) {
	w := writer.(util.BufWriter)
	if entering {
		_ = w.WriteByte('<')
		_, _ = w.WriteString("strong")
		if node.Attributes() != nil {
			RenderAttributes(w, node, StrongAttributeFilter)
		}
		_ = w.WriteByte('>')
	} else {
		_, _ = w.WriteString("</strong>")
	}
	return ast.WalkContinue, nil
}

func (e *commonMark) renderLink(
	writer io.Writer, source []byte, node ast.Node, entering bool, _ renderer.Context) (ast.WalkStatus, error) {
	w := writer.(util.BufWriter)
	n := node.(*ast.Link)
	if entering {
		_, _ = w.WriteString("<a href=\"")
		dest := util.URLEscape(n.Destination.Bytes(source), true)
		if e.config.Unsafe || !IsDangerousURL(dest) {
			_, _ = w.Write(util.EscapeHTML(dest))
		}
		_ = w.WriteByte('"')
		if title := n.Title.Bytes(source); len(title) > 0 {
			_, _ = w.WriteString(` title="`)
			e.writer.WriteText(w, title)
			_ = w.WriteByte('"')
		}
		if n.Attributes() != nil {
			RenderAttributes(w, n, LinkAttributeFilter)
		}
		_ = w.WriteByte('>')
	} else {
		_, _ = w.WriteString("</a>")
	}
	return ast.WalkContinue, nil
}

// ImageAttributeFilter defines attribute names which image elements can have.
var ImageAttributeFilter = GlobalAttributeFilter.ExtendString(`align,border,crossorigin,decoding,height,importance,intrinsicsize,ismap,loading,referrerpolicy,sizes,srcset,usemap,width`) // nolint: lll

func (e *commonMark) renderImage(
	writer io.Writer, source []byte, node ast.Node, entering bool, rctx renderer.Context) (ast.WalkStatus, error) {
	w := writer.(util.BufWriter)
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.Image)
	_, _ = w.WriteString("<img src=\"")
	dest := util.URLEscape(n.Destination.Bytes(source), true)
	if e.config.Unsafe || !IsDangerousURL(dest) {
		_, _ = w.Write(util.EscapeHTML(dest))
	}
	_, _ = w.WriteString(`" alt="`)
	e.renderTexts(w, source, n, rctx)
	_ = w.WriteByte('"')
	if title := n.Title.Bytes(source); len(title) > 0 {
		_, _ = w.WriteString(` title="`)
		e.writer.WriteText(w, title)
		_ = w.WriteByte('"')
	}
	if n.Attributes() != nil {
		RenderAttributes(w, n, ImageAttributeFilter)
	}
	if e.config.XHTML {
		_, _ = w.WriteString(" />")
	} else {
		_, _ = w.WriteString(">")
	}
	return ast.WalkSkipChildren, nil
}

func (e *commonMark) renderRawHTML(
	writer io.Writer, source []byte, node ast.Node, entering bool, _ renderer.Context) (ast.WalkStatus, error) {
	w := writer.(util.BufWriter)
	if !entering {
		return ast.WalkSkipChildren, nil
	}
	if e.config.Unsafe {
		n := node.(*ast.RawHTML)
		_, _ = w.Write(n.Value.Bytes(source))
		return ast.WalkSkipChildren, nil
	}
	_, _ = w.WriteString("<!-- raw HTML omitted -->")
	return ast.WalkSkipChildren, nil
}

func (e *commonMark) renderText(
	writer io.Writer, source []byte, node ast.Node, entering bool, _ renderer.Context) (ast.WalkStatus, error) {
	w := writer.(util.BufWriter)
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.Text)
	value := n.Value.Bytes(source)
	if n.Value.IsOwned() {
		// Literal text (e.g. typographer HTML entity substitutions) is written
		// directly to the output without any further escaping or processing.
		_, _ = w.Write(value)
		return ast.WalkContinue, nil
	}
	e.writer.WriteText(w, value)
	if n.HardLineBreak() || (n.SoftLineBreak() && e.config.HardWraps) {
		if e.config.XHTML {
			_, _ = w.WriteString("<br />\n")
		} else {
			_, _ = w.WriteString("<br>\n")
		}
	} else if n.SoftLineBreak() {
		if e.config.EastAsianLineBreaks != EastAsianLineBreaksNone && len(value) != 0 {
			sibling := node.NextSibling()
			if sibling != nil && sibling.Kind() == ast.KindText {
				if siblingText := sibling.(*ast.Text).Value.Bytes(source); len(siblingText) != 0 {
					thisLastRune := util.ToRune(value, len(value)-1)
					siblingFirstRune, _ := utf8.DecodeRune(siblingText)
					if e.config.EastAsianLineBreaks.softLineBreak(thisLastRune, siblingFirstRune) {
						_ = w.WriteByte('\n')
					}
				}
			}
		} else {
			_ = w.WriteByte('\n')
		}
	}
	return ast.WalkContinue, nil
}

func (e *commonMark) renderTexts(w util.BufWriter, source []byte, n ast.Node, rctx renderer.Context) {
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if t, ok := c.(*ast.Text); ok {
			_, _ = e.renderText(w, source, t, true, rctx)
		} else {
			e.renderTexts(w, source, c, rctx)
		}
	}
}

var dataPrefix = []byte("data-")

// RenderAttributes renders given node's attributes.
// You can specify attribute names to render by the filter.
// If filter is nil, RenderAttributes renders all attributes.
func RenderAttributes(writer io.Writer, node ast.Node, filter util.BytesFilter) {
	w, ok := writer.(util.BufWriter)
	if !ok {
		w = bufio.NewWriter(writer)
	}
	for _, attr := range node.Attributes() {
		name := attr.Name
		if filter != nil && !filter.Contains(name) {
			if !bytes.HasPrefix(name, dataPrefix) {
				continue
			}
		}
		_, _ = w.WriteString(" ")
		_, _ = w.Write(name)
		_, _ = w.WriteString(`="`)
		_, _ = w.Write(util.EscapeHTML(attr.Value.Bytes(nil)))
		_ = w.WriteByte('"')
	}
}

// A Writer interface writes textual contents to a writer.
type Writer interface {
	// WriteText writes the given source to writer with resolving references and unescaping
	// backslash escaped characters.
	WriteText(writer io.Writer, source []byte)

	// RawWriteText writes the given source to writer without resolving references and
	// unescaping backslash escaped characters.
	RawWriteText(writer io.Writer, source []byte)

	// WriteHTML writes the given html.
	// WriteHTML replaces null bytes with the replacement character and writes the result to writer.
	WriteHTML(writer io.Writer, html []byte)
}

var replacementCharacter = []byte("\ufffd")

type writerConfig struct {
	EscapedSpace bool
}

// A WriterOption interface sets options for HTML based writers.
type WriterOption interface {
	SetWriterOption(*writerConfig)
}

type withEscapedSpace struct {
	v bool
}

func (o *withEscapedSpace) SetWriterOption(c *writerConfig) {
	c.EscapedSpace = true
}

func (o *withEscapedSpace) SetFormatOption(c *Config) {
	c.EscapedSpace = true
}

// WithEscapedSpace is a WriterOption indicates that a '\' escaped half-space(0x20) should not be rendered.
func WithEscapedSpace() interface {
	WriterOption
	Option
} {
	return &withEscapedSpace{true}
}

type defaultWriter struct {
	writerConfig
}

// NewWriter returns a new Writer.
func NewWriter(opts ...WriterOption) Writer {
	w := &defaultWriter{}
	for _, opt := range opts {
		opt.SetWriterOption(&w.writerConfig)
	}
	return w
}

func escapeRune(writer io.Writer, r rune) {
	if r < 256 {
		v := util.EscapeHTMLByte(byte(r))
		if v != nil {
			_, _ = writer.Write(v)
			return
		}
	}

	r = util.ToValidRune(r)
	var buf [utf8.UTFMax]byte
	n := utf8.EncodeRune(buf[:], r)
	_, _ = writer.Write(buf[:n])
}

func (d *defaultWriter) WriteHTML(writer io.Writer, source []byte) {
	for len(source) > 0 {
		i := bytes.IndexByte(source, '\u0000')
		if i < 0 {
			_, _ = writer.Write(source)
			return
		}
		if i > 0 {
			_, _ = writer.Write(source[:i])
		}
		_, _ = writer.Write(replacementCharacter)
		source = source[i+1:]
	}
}

func (d *defaultWriter) RawWriteText(writer io.Writer, source []byte) {
	n := 0
	l := len(source)
	for i := range l {
		v := util.EscapeHTMLByte(source[i])
		if v != nil {
			_, _ = writer.Write(source[i-n : i])
			n = 0
			_, _ = writer.Write(v)
			continue
		}
		n++
	}
	if n != 0 {
		_, _ = writer.Write(source[l-n:])
	}
}

func (d *defaultWriter) WriteText(writer io.Writer, source []byte) {
	escaped := false
	var ok bool
	limit := len(source)
	n := 0
	for i := 0; i < limit; i++ {
		c := source[i]
		if escaped {
			if util.IsPunct(c) {
				d.RawWriteText(writer, source[n:i-1])
				n = i
				escaped = false
				continue
			}
			if d.EscapedSpace && c == ' ' {
				d.RawWriteText(writer, source[n:i-1])
				n = i + 1
				escaped = false
				continue
			}
		}
		if c == '\x00' {
			d.RawWriteText(writer, source[n:i])
			d.RawWriteText(writer, replacementCharacter)
			n = i + 1
			escaped = false
			continue
		}
		if c == '&' {
			pos := i
			next := i + 1
			if next < limit && source[next] == '#' {
				nnext := next + 1
				if nnext < limit {
					nc := source[nnext]
					// code point like #x22;
					if nnext < limit && nc == 'x' || nc == 'X' {
						start := nnext + 1
						i, ok = util.ReadWhile(source, [2]int{start, limit}, util.IsHexDecimal)
						if ok && i < limit && source[i] == ';' && i-start < 7 {
							v, _ := strconv.ParseUint(util.BytesToReadOnlyString(source[start:i]), 16, 32)
							d.RawWriteText(writer, source[n:pos])
							n = i + 1
							escapeRune(writer, rune(v))
							continue
						}
						// code point like #1234;
					} else if nc >= '0' && nc <= '9' {
						start := nnext
						i, ok = util.ReadWhile(source, [2]int{start, limit}, util.IsNumeric)
						if ok && i < limit && i-start < 8 && source[i] == ';' {
							v, _ := strconv.ParseUint(util.BytesToReadOnlyString(source[start:i]), 10, 32)
							d.RawWriteText(writer, source[n:pos])
							n = i + 1
							escapeRune(writer, rune(v))
							continue
						}
					}
				}
			} else {
				start := next
				i, ok = util.ReadWhile(source, [2]int{start, limit}, util.IsAlphaNumeric)
				// entity reference
				if ok && i < limit && source[i] == ';' {
					name := util.BytesToReadOnlyString(source[start:i])
					entity, ok := util.LookUpHTML5EntityByName(name)
					if ok {
						d.RawWriteText(writer, source[n:pos])
						n = i + 1
						d.RawWriteText(writer, entity.Characters)
						continue
					}
				}
			}
			i = next - 1
		}
		if c == '\\' {
			escaped = true
			continue
		}
		escaped = false
	}
	d.RawWriteText(writer, source[n:])
}

// DefaultWriter is a default instance of the Writer.
var DefaultWriter = NewWriter()

var bDataImage = []byte("data:image/")
var bPng = []byte("png;")
var bGif = []byte("gif;")
var bJpeg = []byte("jpeg;")
var bWebp = []byte("webp;")
var bJs = []byte("javascript:")
var bVb = []byte("vbscript:")
var bFile = []byte("file:")
var bData = []byte("data:")

func hasPrefix(s, prefix []byte) bool {
	return len(s) >= len(prefix) && bytes.Equal(bytes.ToLower(s[0:len(prefix)]), bytes.ToLower(prefix))
}

// IsDangerousURL returns true if the given url seems a potentially dangerous url,
// otherwise false.
func IsDangerousURL(url []byte) bool {
	if hasPrefix(url, bDataImage) && len(url) >= 11 {
		v := url[11:]
		if hasPrefix(v, bPng) || hasPrefix(v, bGif) ||
			hasPrefix(v, bJpeg) || hasPrefix(v, bWebp) {
			return false
		}
		return true
	}
	return hasPrefix(url, bJs) || hasPrefix(url, bVb) ||
		hasPrefix(url, bFile) || hasPrefix(url, bData)
}
