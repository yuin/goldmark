package ast

import (
	"fmt"
	"strings"

	textm "github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// A BaseInline struct implements the Node interface.
type BaseInline struct {
	BaseNode
}

// Type implements Node.Type
func (b *BaseInline) Type() NodeType {
	return InlineNode
}

// IsRaw implements Node.IsRaw
func (b *BaseInline) IsRaw() bool {
	return false
}

// HasBlankPreviousLines implements Node.HasBlankPreviousLines.
func (b *BaseInline) HasBlankPreviousLines() bool {
	panic("can not call with inline nodes.")
}

// SetBlankPreviousLines implements Node.SetBlankPreviousLines.
func (b *BaseInline) SetBlankPreviousLines(v bool) {
	panic("can not call with inline nodes.")
}

// Lines implements Node.Lines
func (b *BaseInline) Lines() *textm.Segments {
	panic("can not call with inline nodes.")
}

// SetLines implements Node.SetLines
func (b *BaseInline) SetLines(v *textm.Segments) {
	panic("can not call with inline nodes.")
}

// A Text struct represents a textual content of the Markdown text.
type Text struct {
	BaseInline
	// Segment is a position in a source text.
	Segment textm.Segment

	flags uint8
}

const (
	textSoftLineBreak = 1 << iota
	textHardLineBreak
	textRaw
)

// Inline implements Inline.Inline.
func (n *Text) Inline() {
}

// SoftLineBreak returns true if this node ends with a new line,
// otherwise false.
func (n *Text) SoftLineBreak() bool {
	return n.flags&textSoftLineBreak != 0
}

// SetSoftLineBreak sets whether this node ends with a new line.
func (n *Text) SetSoftLineBreak(v bool) {
	if v {
		n.flags |= textSoftLineBreak
	} else {
		n.flags = n.flags &^ textHardLineBreak
	}
}

// IsRaw returns true if this text should be rendered without unescaping
// back slash escapes and resolving references.
func (n *Text) IsRaw() bool {
	return n.flags&textRaw != 0
}

// SetRaw sets whether this text should be rendered as raw contents.
func (n *Text) SetRaw(v bool) {
	if v {
		n.flags |= textRaw
	} else {
		n.flags = n.flags &^ textRaw
	}
}

// HardLineBreak returns true if this node ends with a hard line break.
// See https://spec.commonmark.org/0.29/#hard-line-breaks for details.
func (n *Text) HardLineBreak() bool {
	return n.flags&textHardLineBreak != 0
}

// SetHardLineBreak sets whether this node ends with a hard line break.
func (n *Text) SetHardLineBreak(v bool) {
	if v {
		n.flags |= textHardLineBreak
	} else {
		n.flags = n.flags &^ textHardLineBreak
	}
}

// Merge merges a Node n into this node.
// Merge returns true if given node has been merged, otherwise false.
func (n *Text) Merge(node Node, source []byte) bool {
	t, ok := node.(*Text)
	if !ok {
		return false
	}
	if n.Segment.Stop != t.Segment.Start || t.Segment.Padding != 0 || source[n.Segment.Stop-1] == '\n' || t.IsRaw() != n.IsRaw() {
		return false
	}
	n.Segment.Stop = t.Segment.Stop
	n.SetSoftLineBreak(t.SoftLineBreak())
	n.SetHardLineBreak(t.HardLineBreak())
	return true
}

// Text implements Node.Text.
func (n *Text) Text(source []byte) []byte {
	return n.Segment.Value(source)
}

// Dump implements Node.Dump.
func (n *Text) Dump(source []byte, level int) {
	fmt.Printf("%sText: \"%s\"\n", strings.Repeat("    ", level), strings.TrimRight(string(n.Text(source)), "\n"))
}

// NewText returns a new Text node.
func NewText() *Text {
	return &Text{
		BaseInline: BaseInline{},
	}
}

// NewTextSegment returns a new Text node with given source potision.
func NewTextSegment(v textm.Segment) *Text {
	return &Text{
		BaseInline: BaseInline{},
		Segment:    v,
	}
}

// NewRawTextSegment returns a new Text node with given source position.
// The new node should be rendered as raw contents.
func NewRawTextSegment(v textm.Segment) *Text {
	t := &Text{
		BaseInline: BaseInline{},
		Segment:    v,
	}
	t.SetRaw(true)
	return t
}

// MergeOrAppendTextSegment merges a given s into the last child of the parent if
// it can be merged, otherwise creates a new Text node and appends it to after current
// last child.
func MergeOrAppendTextSegment(parent Node, s textm.Segment) {
	last := parent.LastChild()
	t, ok := last.(*Text)
	if ok && t.Segment.Stop == s.Start && !t.SoftLineBreak() {
		ts := t.Segment
		t.Segment = ts.WithStop(s.Stop)
	} else {
		parent.AppendChild(parent, NewTextSegment(s))
	}
}

// MergeOrReplaceTextSegment merges a given s into a previous sibling of the node n
// if a previous sibling of the node n is *Text, otherwise replaces Node n with s.
func MergeOrReplaceTextSegment(parent Node, n Node, s textm.Segment) {
	prev := n.PreviousSibling()
	if t, ok := prev.(*Text); ok && t.Segment.Stop == s.Start && !t.SoftLineBreak() {
		t.Segment = t.Segment.WithStop(s.Stop)
		parent.RemoveChild(parent, n)
	} else {
		parent.ReplaceChild(parent, n, NewTextSegment(s))
	}
}

// A CodeSpan struct represents a code span of Markdown text.
type CodeSpan struct {
	BaseInline
}

// Inline implements Inline.Inline .
func (n *CodeSpan) Inline() {
}

// IsBlank returns true if this node consists of spaces, otherwise false.
func (n *CodeSpan) IsBlank(source []byte) bool {
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		text := c.(*Text).Segment
		if !util.IsBlank(text.Value(source)) {
			return false
		}
	}
	return true
}

// Dump implements Node.Dump
func (n *CodeSpan) Dump(source []byte, level int) {
	DumpHelper(n, source, level, "CodeSpan", nil, nil)
}

// NewCodeSpan returns a new CodeSpan node.
func NewCodeSpan() *CodeSpan {
	return &CodeSpan{
		BaseInline: BaseInline{},
	}
}

// An Emphasis struct represents an emphasis of Markdown text.
type Emphasis struct {
	BaseInline

	// Level is a level of the emphasis.
	Level int
}

// Inline implements Inline.Inline.
func (n *Emphasis) Inline() {
}

// Dump implements Node.Dump.
func (n *Emphasis) Dump(source []byte, level int) {
	DumpHelper(n, source, level, fmt.Sprintf("Emphasis(%d)", n.Level), nil, nil)
}

// NewEmphasis returns a new Emphasis node with given level.
func NewEmphasis(level int) *Emphasis {
	return &Emphasis{
		BaseInline: BaseInline{},
		Level:      level,
	}
}

type baseLink struct {
	BaseInline

	// Destination is a destination(URL) of this link.
	Destination []byte

	// Title is a title of this link.
	Title []byte
}

// Inline implements Inline.Inline.
func (n *baseLink) Inline() {
}

func (n *baseLink) Dump(source []byte, level int) {
	m := map[string]string{}
	m["Destination"] = string(n.Destination)
	m["Title"] = string(n.Title)
	DumpHelper(n, source, level, "Link", m, nil)
}

// A Link struct represents a link of the Markdown text.
type Link struct {
	baseLink
}

// NewLink returns a new Link node.
func NewLink() *Link {
	c := &Link{
		baseLink: baseLink{
			BaseInline: BaseInline{},
		},
	}
	return c
}

// An Image struct represents an image of the Markdown text.
type Image struct {
	baseLink
}

// Dump implements Node.Dump.
func (n *Image) Dump(source []byte, level int) {
	m := map[string]string{}
	m["Destination"] = string(n.Destination)
	m["Title"] = string(n.Title)
	DumpHelper(n, source, level, "Image", m, nil)
}

// NewImage returns a new Image node.
func NewImage(link *Link) *Image {
	c := &Image{
		baseLink: baseLink{
			BaseInline: BaseInline{},
		},
	}
	c.Destination = link.Destination
	c.Title = link.Title
	for n := link.FirstChild(); n != nil; {
		next := n.NextSibling()
		link.RemoveChild(link, n)
		c.AppendChild(c, n)
		n = next
	}

	return c
}

// AutoLinkType defines kind of auto links.
type AutoLinkType int

const (
	// AutoLinkEmail indicates that an autolink is an email address.
	AutoLinkEmail = iota + 1
	// AutoLinkURL indicates that an autolink is a generic URL.
	AutoLinkURL
)

// An AutoLink struct represents an autolink of the Markdown text.
type AutoLink struct {
	BaseInline
	// Value is a link text of this node.
	Value *Text
	// Type is a type of this autolink.
	AutoLinkType AutoLinkType
}

// Inline implements Inline.Inline.
func (n *AutoLink) Inline() {}

// Dump implenets Node.Dump
func (n *AutoLink) Dump(source []byte, level int) {
	segment := n.Value.Segment
	m := map[string]string{
		"Value": string(segment.Value(source)),
	}
	DumpHelper(n, source, level, "AutoLink", m, nil)
}

// NewAutoLink returns a new AutoLink node.
func NewAutoLink(typ AutoLinkType, value *Text) *AutoLink {
	return &AutoLink{
		BaseInline:   BaseInline{},
		Value:        value,
		AutoLinkType: typ,
	}
}

// A RawHTML struct represents an inline raw HTML of the Markdown text.
type RawHTML struct {
	BaseInline
}

// Inline implements Inline.Inline.
func (n *RawHTML) Inline() {}

// Dump implements Node.Dump.
func (n *RawHTML) Dump(source []byte, level int) {
	DumpHelper(n, source, level, "RawHTML", nil, nil)
}

// NewRawHTML returns a new RawHTML node.
func NewRawHTML() *RawHTML {
	return &RawHTML{
		BaseInline: BaseInline{},
	}
}
