// Package html implements renderer that outputs HTMLs.
package html

import (
	"bytes"
	"fmt"
	"io"
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

// NodeRendererDecorator is a decorator that decorates NodeRenderer.
type NodeRendererDecorator = renderer.NodeRendererDecorator[io.Writer]

// NodeRendererFunc adapts f into a NodeRenderer that renders AST nodes as (X)HTML.
func NodeRendererFunc(f func(w io.Writer, source []byte,
	n ast.Node, entering bool, rc renderer.Context) (ast.WalkStatus, error)) NodeRenderer {
	return renderer.NodeRendererFunc(f)
}

// Extension is an extension that extends HTML based renderers.
type Extension = renderer.Extension[Config]

// Option is a functional option for HTML based renderers.
type Option = renderer.Option[Config]

// WithNodeRenderers sets a node renderer for the given node kind.
func WithNodeRenderers(nodeRenderers map[ast.NodeKind]NodeRenderer) Option {
	return renderer.WithNodeRenderers[io.Writer, Config](nodeRenderers)
}

// WithNodeRenderer sets a node renderer for the given node kind.
func WithNodeRenderer(kind ast.NodeKind, nodeRenderer NodeRenderer) Option {
	return renderer.WithNodeRenderer[io.Writer, Config](kind, nodeRenderer)
}

// WithNodeRendererDecorators sets a node renderer decorator for the given node kind.
//
// If a decorator is already set for a node kind, the new decorator will be applied to the existing one.
func WithNodeRendererDecorators(decorators map[ast.NodeKind]NodeRendererDecorator) Option {
	return renderer.WithNodeRendererDecorators[io.Writer, Config](decorators)
}

// WithNodeRendererDecorator sets a node renderer decorator for the given node kind.
//
// If a decorator is already set for a node kind, the new decorator will be applied to the existing one.
func WithNodeRendererDecorator(kind ast.NodeKind, decorator NodeRendererDecorator) Option {
	return renderer.WithNodeRendererDecorator[io.Writer, Config](kind, decorator)
}

// WithExtensions adds extensions.
func WithExtensions(extensions ...Extension) Option {
	return renderer.WithExtensions[io.Writer](extensions...)
}

// }}} Type aliases

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
	HardWraps         bool
	LineBreakStrategy LineBreakStrategy
	XHTML             bool
	Unsafe            bool
	Paragraph         ParagraphConfig
}

// Default returns a Config with default values.
func (c Config) Default() Config {
	return Config{
		HardWraps:         false,
		LineBreakStrategy: nil,
		XHTML:             false,
		Unsafe:            false,
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

// LineBreakStrategy is an interface that defines a strategy for determining whether a line break
// should be rendered as a new line.
type LineBreakStrategy interface {
	// SoftLineBreak returns true if a soft line break should be rendered as a new line.
	SoftLineBreak(thisLastRune rune, siblingFirstRune rune) bool
}

// WithLineBreakStrategy is a functional option that sets a custom line break strategy.
func WithLineBreakStrategy(strategy LineBreakStrategy) Option {
	return renderer.NewOptionFunc(func(c *Config) {
		c.LineBreakStrategy = strategy
	})
}

type simpleEastAsianLineBreakStrategy struct{}

// SimpleEastAsianLineBreakStrategy follows east_asian_line_breaks in Pandoc.
var SimpleEastAsianLineBreakStrategy LineBreakStrategy = &simpleEastAsianLineBreakStrategy{}

func (s *simpleEastAsianLineBreakStrategy) SoftLineBreak(thisLastRune rune, siblingFirstRune rune) bool {
	return !util.IsEastAsianWideRune(thisLastRune) || !util.IsEastAsianWideRune(siblingFirstRune)
}

type cssText3LineBreakStrategy struct{}

// CSSText3LineBreakStrategy follows CSS Text Module Level 3 with some enhancements for CJK.
var CSSText3LineBreakStrategy LineBreakStrategy = &cssText3LineBreakStrategy{}

func (s *cssText3LineBreakStrategy) SoftLineBreak(thisLastRune rune, siblingFirstRune rune) bool {
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

// WithIsInTightBlockFunc is a functional option that sets a custom function to
// determine whether a paragraph node is inside a tight block and should be
// rendered without <p> tags.
func WithIsInTightBlockFunc(fn IsInTightBlockFunc) Option {
	return renderer.NewOptionFunc(func(c *Config) {
		c.Paragraph.IsInTightBlockFunc = fn
	})
}

var htmlWriterKey = renderer.NewContextKey()
var textWriterKey = renderer.NewContextKey()
var linkURLWriterKey = renderer.NewContextKey()

type htmlRenderer struct {
	*renderer.Helper[io.Writer, Config]
}

// New returns a new Renderer with given options.
func New(opts ...Option) Renderer {
	cm := NewCommonMark()
	opts = append([]Option{WithExtensions(cm)}, opts...)
	var hb renderer.HelperBuilder[io.Writer, Config]
	helper := hb.Options(opts...).OnBeforeRender(
		func(w io.Writer, _ []byte, _ ast.Node, rc renderer.Context) error {
			var hw io.Writer = &htmlWriter{w.(util.BufWriter)}
			var tw io.Writer = &textWriter{w.(util.BufWriter)}
			var lw io.Writer = &linkURLWriter{w.(util.BufWriter)}
			rc.Set(htmlWriterKey, hw)
			rc.Set(textWriterKey, tw)
			rc.Set(linkURLWriterKey, lw)
			return nil
		}).Build()
	r := &htmlRenderer{
		Helper: helper,
	}
	return r
}

// Render renders the given AST node to the given writer.
func (r *htmlRenderer) Render(w io.Writer, source []byte, n ast.Node, opts ...renderer.RenderOption) error {
	if ew, ok := w.(util.ErrorBufWriter); ok {
		return r.Helper.Render(ew, source, n, opts...)
	}

	return r.Helper.Render(util.NewErrorBufWriterSize(w, len(source)*3), source, n, opts...)
}

// RenderStringSource renders the given AST node to the given writer.
func (r *htmlRenderer) RenderStringSource(w io.Writer, source string, n ast.Node, opts ...renderer.RenderOption) error {
	return r.Render(w, util.StringToReadOnlyBytes(source), n, opts...)
}

// ContextHTMLWriter returns a writer that writes HTML content.
func ContextHTMLWriter(rc renderer.Context) util.BufWriter {
	v := rc.Get(htmlWriterKey)
	if v == nil {
		panic("HTMLWriter not found in context")
	}
	return v.(util.BufWriter)
}

// ContextTextWriter returns a writer that writes text content.
func ContextTextWriter(rc renderer.Context) util.BufWriter {
	v := rc.Get(textWriterKey)
	if v == nil {
		panic("TextWriter not found in context")
	}
	return v.(util.BufWriter)
}

// ContextLinkURLWriter returns a writer that writes link URL content.
func ContextLinkURLWriter(rc renderer.Context) util.BufWriter {
	v := rc.Get(linkURLWriterKey)
	if v == nil {
		panic("LinkURLWriter not found in context")
	}
	return v.(util.BufWriter)
}

type commonMark struct {
	opts   []Option
	config *Config
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
	writer io.Writer, source []byte, node ast.Node, entering bool, rc renderer.Context) (ast.WalkStatus, error) {
	w := writer.(util.BufWriter)
	n := node.(*ast.Heading)
	if entering {
		_, _ = w.WriteString("<h")
		_ = w.WriteByte("0123456"[n.Level])
		if n.Attributes() != nil {
			RenderAttributes(w, source, node, HeadingAttributeFilter, rc)
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
	writer io.Writer, source []byte, n ast.Node, entering bool, rc renderer.Context) (ast.WalkStatus, error) {
	w := writer.(util.BufWriter)
	if entering {
		if n.Attributes() != nil {
			_, _ = w.WriteString("<blockquote")
			RenderAttributes(w, source, n, BlockquoteAttributeFilter, rc)
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
	writer io.Writer, source []byte, node ast.Node, entering bool, rc renderer.Context) (ast.WalkStatus, error) {
	w := writer.(util.BufWriter)
	n := node.(*ast.CodeBlock)
	if entering {
		tw := ContextTextWriter(rc)
		_, _ = w.WriteString("<pre><code")
		language, ok := n.Language(source)
		if ok {
			_, _ = w.WriteString(" class=\"language-")
			_, _ = tw.WriteString(language)
			_ = w.WriteByte('"')
		}
		_ = w.WriteByte('>')
		_, _ = n.Value.WriteTo(tw, source)
	} else {
		_, _ = w.WriteString("</code></pre>\n")
	}
	return ast.WalkContinue, nil
}

func (e *commonMark) renderHTMLBlock(
	writer io.Writer, source []byte, node ast.Node, entering bool, rc renderer.Context) (ast.WalkStatus, error) {
	w := writer.(util.BufWriter)
	hw := ContextHTMLWriter(rc)
	n := node.(*ast.HTMLBlock)
	if entering {
		if e.config.Unsafe {
			_, _ = n.Value.WriteTo(hw, source)
		} else {
			_, _ = w.WriteString("<!-- raw HTML omitted -->\n")
		}
	}
	return ast.WalkContinue, nil
}

// ListAttributeFilter defines attribute names which list elements can have.
var ListAttributeFilter = GlobalAttributeFilter.ExtendString(`start,reversed,type`)

func (e *commonMark) renderList(
	writer io.Writer, source []byte, node ast.Node, entering bool, rc renderer.Context) (ast.WalkStatus, error) {
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
			RenderAttributes(w, source, n, ListAttributeFilter, rc)
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
	writer io.Writer, source []byte, n ast.Node, entering bool, rc renderer.Context) (ast.WalkStatus, error) {
	w := writer.(util.BufWriter)
	if entering {
		if n.Attributes() != nil {
			_, _ = w.WriteString("<li")
			RenderAttributes(w, source, n, ListItemAttributeFilter, rc)
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
	writer io.Writer, source []byte, n ast.Node, entering bool, rc renderer.Context) (ast.WalkStatus, error) {
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
			RenderAttributes(w, source, n, ParagraphAttributeFilter, rc)
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
	writer io.Writer, source []byte, n ast.Node, entering bool, rc renderer.Context) (ast.WalkStatus, error) {
	w := writer.(util.BufWriter)
	if !entering {
		return ast.WalkContinue, nil
	}
	_, _ = w.WriteString("<hr")
	if n.Attributes() != nil {
		RenderAttributes(w, source, n, ThematicAttributeFilter, rc)
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
	writer io.Writer, source []byte, node ast.Node, entering bool, rc renderer.Context) (ast.WalkStatus, error) {
	w := writer.(util.BufWriter)
	n := node.(*ast.AutoLink)
	if !entering {
		return ast.WalkContinue, nil
	}
	_, _ = w.WriteString(`<a href="`)
	dest := n.Destination.Value(source)
	if e.config.Unsafe || !IsDangerousURL(dest) {
		_, _ = ContextLinkURLWriter(rc).WriteString(dest)
	}
	if n.Attributes() != nil {
		_ = w.WriteByte('"')
		RenderAttributes(w, source, n, LinkAttributeFilter, rc)
		_ = w.WriteByte('>')
	} else {
		_, _ = w.WriteString(`">`)
	}
	_, _ = n.Label.WriteTo(ContextTextWriter(rc), source)
	_, _ = w.WriteString(`</a>`)
	return ast.WalkContinue, nil
}

// CodeAttributeFilter defines attribute names which code elements can have.
var CodeAttributeFilter = GlobalAttributeFilter

func (e *commonMark) renderCodeSpan(
	writer io.Writer, source []byte, n ast.Node, entering bool, rc renderer.Context) (ast.WalkStatus, error) {
	w := writer.(util.BufWriter)
	tw := ContextTextWriter(rc)
	if entering {
		if n.Attributes() != nil {
			_, _ = w.WriteString("<code")
			RenderAttributes(w, source, n, CodeAttributeFilter, rc)
			_ = w.WriteByte('>')
		} else {
			_, _ = w.WriteString("<code>")
		}
		_, _ = n.(*ast.CodeSpan).Value.WriteTo(tw, source)
		return ast.WalkSkipChildren, nil
	}
	_, _ = w.WriteString("</code>")
	return ast.WalkContinue, nil
}

// EmphasisAttributeFilter defines attribute names which emphasis elements can have.
var EmphasisAttributeFilter = GlobalAttributeFilter

func (e *commonMark) renderEmphasis(
	writer io.Writer, source []byte, node ast.Node, entering bool, rc renderer.Context) (ast.WalkStatus, error) {
	w := writer.(util.BufWriter)
	if entering {
		_ = w.WriteByte('<')
		_, _ = w.WriteString("em")
		if node.Attributes() != nil {
			RenderAttributes(w, source, node, EmphasisAttributeFilter, rc)
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
	writer io.Writer, source []byte, node ast.Node, entering bool, rc renderer.Context) (ast.WalkStatus, error) {
	w := writer.(util.BufWriter)
	if entering {
		_ = w.WriteByte('<')
		_, _ = w.WriteString("strong")
		if node.Attributes() != nil {
			RenderAttributes(w, source, node, StrongAttributeFilter, rc)
		}
		_ = w.WriteByte('>')
	} else {
		_, _ = w.WriteString("</strong>")
	}
	return ast.WalkContinue, nil
}

func (e *commonMark) renderLink(
	writer io.Writer, source []byte, node ast.Node, entering bool, rc renderer.Context) (ast.WalkStatus, error) {
	w := writer.(util.BufWriter)
	tw := textWriter{w}
	n := node.(*ast.Link)
	if entering {
		_, _ = w.WriteString("<a href=\"")
		dest := n.Destination.Value(source)
		if e.config.Unsafe || !IsDangerousURL(dest) {
			_, _ = ContextLinkURLWriter(rc).WriteString(dest)
		}
		_ = w.WriteByte('"')
		if title := n.Title.Bytes(source); len(title) > 0 {
			_, _ = w.WriteString(` title="`)
			_, _ = n.Title.WriteTo(&tw, source)
			_ = w.WriteByte('"')
		}
		if n.Attributes() != nil {
			RenderAttributes(w, source, n, LinkAttributeFilter, rc)
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
	writer io.Writer, source []byte, node ast.Node, entering bool, rc renderer.Context) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	w := writer.(util.BufWriter)
	tw := textWriter{w}
	n := node.(*ast.Image)
	_, _ = w.WriteString("<img src=\"")
	dest := n.Destination.Value(source)
	if e.config.Unsafe || !IsDangerousURL(dest) {
		_, _ = ContextLinkURLWriter(rc).WriteString(dest)
	}
	_, _ = w.WriteString(`" alt="`)
	e.renderTexts(w, source, n, rc)
	_ = w.WriteByte('"')
	if title := n.Title.Bytes(source); len(title) > 0 {
		_, _ = w.WriteString(` title="`)
		_, _ = n.Title.WriteTo(&tw, source)
		_ = w.WriteByte('"')
	}
	if n.Attributes() != nil {
		RenderAttributes(w, source, n, ImageAttributeFilter, rc)
	}
	if e.config.XHTML {
		_, _ = w.WriteString(" />")
	} else {
		_, _ = w.WriteString(">")
	}
	return ast.WalkSkipChildren, nil
}

func (e *commonMark) renderRawHTML(
	writer io.Writer, source []byte, node ast.Node, entering bool, rc renderer.Context) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkSkipChildren, nil
	}
	w := writer.(util.BufWriter)
	if e.config.Unsafe {
		n := node.(*ast.RawHTML)
		hw := ContextHTMLWriter(rc)
		_, _ = n.Value.WriteTo(hw, source)
		return ast.WalkSkipChildren, nil
	}
	_, _ = w.WriteString("<!-- raw HTML omitted -->")
	return ast.WalkSkipChildren, nil
}

func (e *commonMark) renderText(
	writer io.Writer, source []byte, node ast.Node, entering bool, rc renderer.Context) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	w := writer.(util.BufWriter)
	tw := ContextTextWriter(rc)
	n := node.(*ast.Text)
	_, _ = n.Value.WriteTo(tw, source)
	if n.HardLineBreak() || (n.SoftLineBreak() && e.config.HardWraps) {
		if e.config.XHTML {
			_, _ = w.WriteString("<br />\n")
		} else {
			_, _ = w.WriteString("<br>\n")
		}
	} else if n.SoftLineBreak() {
		if e.config.LineBreakStrategy != nil && !n.Value.IsEmpty() {
			sibling := node.NextSibling()
			if sibling != nil && sibling.Kind() == ast.KindText {
				if siblingText := sibling.(*ast.Text).Value.Bytes(source); len(siblingText) != 0 {
					value := n.Value.Bytes(source)
					thisLastRune := util.ToRune(value, len(value)-1)
					siblingFirstRune, _ := utf8.DecodeRune(siblingText)
					if e.config.LineBreakStrategy.SoftLineBreak(thisLastRune, siblingFirstRune) {
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

func (e *commonMark) renderTexts(w util.BufWriter, source []byte, n ast.Node, rc renderer.Context) {
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if t, ok := c.(*ast.Text); ok {
			_, _ = e.renderText(w, source, t, true, rc)
		} else {
			e.renderTexts(w, source, c, rc)
		}
	}
}

// RenderAttributes renders the attributes of the given node to the given writer.
func RenderAttributes(writer io.Writer, source []byte, node ast.Node, filter util.BytesFilter, _ renderer.Context) {
	w, ok := writer.(util.BufWriter)
	if !ok {
		w = util.NewErrorBufWriter(writer)
	}
	tw := &textWriter{w}
	for _, attr := range node.Attributes() {
		if filter != nil && !filter.ContainsString(attr.Name) {
			if !strings.HasPrefix(attr.Name, "data-") {
				continue
			}
		}
		_, _ = w.WriteString(" ")
		_, _ = w.WriteString(attr.Name)
		_, _ = w.WriteString(`="`)
		_, _ = attr.Value.WriteTo(tw, source)
		_ = w.WriteByte('"')
	}
}

var _ io.Writer = (*htmlWriter)(nil)

var replacementCharacter = []byte("\ufffd")

type htmlWriter struct {
	w util.BufWriter
}

func (w *htmlWriter) Write(p []byte) (int, error) {
	written := 0
	for len(p) > 0 {
		i := bytes.IndexByte(p, '\u0000')
		if i < 0 {
			wr, err := w.w.Write(p)
			written += wr
			return written, err
		}
		if i > 0 {
			wr, err := w.w.Write(p[:i])
			written += wr
			if err != nil {
				return written, err
			}
		}
		wr, err := w.w.Write(replacementCharacter)
		written += wr
		if err != nil {
			return written, err
		}
		p = p[i+1:]
	}
	return written, nil
}

func (w *htmlWriter) WriteByte(c byte) error {
	b := [1]byte{c}
	_, err := w.Write(b[:])
	return err
}

func (w *htmlWriter) WriteRune(r rune) (int, error) {
	rbuf := [4]byte{}
	n := utf8.EncodeRune(rbuf[:], r)
	return w.Write(rbuf[:n])
}

func (w *htmlWriter) WriteString(s string) (int, error) {
	return w.Write(util.StringToReadOnlyBytes(s))
}

func (w *htmlWriter) Flush() error {
	return w.w.Flush()
}

var _ io.Writer = (*textWriter)(nil)

type textWriter struct {
	w util.BufWriter
}

func (w *textWriter) Write(p []byte) (int, error) {
	written := 0
	n := 0
	l := len(p)
	for i := range l {
		v := util.EscapeHTMLByte(p[i])
		if v != nil {
			wr, err := w.w.Write(p[i-n : i])
			written += wr
			if err != nil {
				return written, err
			}
			n = 0
			wr, err = w.w.Write(v)
			written += wr
			if err != nil {
				return written, err
			}
			continue
		}
		n++
	}
	if n != 0 {
		wr, err := w.w.Write(p[l-n:])
		written += wr
		if err != nil {
			return written, err
		}
	}
	return written, nil
}

func (w *textWriter) WriteByte(c byte) error {
	b := [1]byte{c}
	_, err := w.Write(b[:])
	return err
}

func (w *textWriter) WriteRune(r rune) (int, error) {
	rbuf := [4]byte{}
	n := utf8.EncodeRune(rbuf[:], r)
	return w.Write(rbuf[:n])
}

func (w *textWriter) WriteString(s string) (int, error) {
	return w.Write(util.StringToReadOnlyBytes(s))
}

func (w *textWriter) Flush() error {
	return w.w.Flush()
}

var _ util.BufWriter = (*linkURLWriter)(nil)

type linkURLWriter struct {
	w util.BufWriter
}

func (w *linkURLWriter) Write(p []byte) (int, error) {
	b := util.URLEscape(p)
	return w.w.Write(b)
}

func (w *linkURLWriter) WriteByte(c byte) error {
	b := [1]byte{c}
	_, err := w.Write(b[:])
	return err
}

func (w *linkURLWriter) WriteRune(r rune) (int, error) {
	rbuf := [4]byte{}
	n := utf8.EncodeRune(rbuf[:], r)
	return w.Write(rbuf[:n])
}

func (w *linkURLWriter) WriteString(s string) (int, error) {
	return w.Write(util.StringToReadOnlyBytes(s))
}

func (w *linkURLWriter) Flush() error {
	return w.w.Flush()
}

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
func IsDangerousURL(url string) bool {
	s := util.StringToReadOnlyBytes(url)
	if hasPrefix(s, bDataImage) && len(s) >= 11 {
		v := s[11:]
		if hasPrefix(v, bPng) || hasPrefix(v, bGif) ||
			hasPrefix(v, bJpeg) || hasPrefix(v, bWebp) {
			return false
		}
		return true
	}
	return hasPrefix(s, bJs) || hasPrefix(s, bVb) ||
		hasPrefix(s, bFile) || hasPrefix(s, bData)
}
