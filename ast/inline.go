package ast

import (
	"fmt"
	"strings"

	textm "github.com/yuin/goldmark/v2/text"
	"github.com/yuin/goldmark/v2/util"
)

// A BaseInline struct implements the Node interface partially.
type BaseInline struct {
	BaseNode
}

// Implements InlineNode marker interface.
func (b *BaseInline) inlineNode() {}

// A Text struct represents a textual content of the Markdown text.
type Text struct {
	BaseInline
	// Value is the text value. It is either a source position or a literal string.
	Value textm.Value

	flags uint8
}

const (
	textSoftLineBreak = 1 << iota
	textHardLineBreak
)

func textFlagsString(flags uint8) string {
	buf := []string{}
	if flags&textSoftLineBreak != 0 {
		buf = append(buf, "SoftLineBreak")
	}
	if flags&textHardLineBreak != 0 {
		buf = append(buf, "HardLineBreak")
	}
	return strings.Join(buf, ", ")
}

// Pos implements Node.Pos.
func (n *Text) Pos() int {
	if n.Value.IsOwned() {
		return -1
	}
	return n.Value.Index().Start
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
		n.flags = n.flags &^ textSoftLineBreak
	}
}

// HardLineBreak returns true if this node ends with a hard line break.
// See https://spec.commonmark.org/0.30/#hard-line-breaks for details.
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
// Merge returns true if the given node has been merged, otherwise false.
// Only two adjacent index-based Text nodes can be merged.
func (n *Text) Merge(node Node, source []byte) bool {
	t, ok := node.(*Text)
	if !ok {
		return false
	}
	if n.Value.IsOwned() || t.Value.IsOwned() {
		return false
	}
	ni := n.Value.Index()
	ti := t.Value.Index()
	if ni.Stop != ti.Start || source[ni.Stop-1] == '\n' {
		return false
	}
	n.Value = textm.NewIndexValue(textm.NewIndex(ni.Start, ti.Stop))
	n.SetSoftLineBreak(t.SoftLineBreak())
	n.SetHardLineBreak(t.HardLineBreak())
	return true
}

// Dump implements Node.Dump.
func (n *Text) Dump(source []byte, level int) {
	m := map[string]string{
		"Value": "\"" + strings.TrimRight(string(n.Value.Bytes(source)), "\n") + "\"",
	}
	fs := textFlagsString(n.flags)
	if len(fs) != 0 {
		m["Flags"] = fs
	}
	DumpHelper(n, source, level, m, nil)
}

// KindText is a NodeKind of the Text node.
var KindText = NewNodeKind("Text")

// Kind implements Node.Kind.
func (n *Text) Kind() NodeKind {
	return KindText
}

// NewText returns a new Text node.
func NewText() *Text {
	n := &Text{}
	n.Init(n)
	return n
}

// NewSegmentText returns a new Text node with the given source position.
func NewSegmentText(v textm.Segment) *Text {
	n := &Text{
		Value: textm.NewIndexValue(textm.NewIndex(v.Start, v.Stop)),
	}
	n.Init(n)
	return n
}

// NewStringText returns a new Text node with a literal string value.
func NewStringText(s string) *Text {
	n := &Text{
		Value: textm.NewStringValue(s),
	}
	n.Init(n)
	return n
}

// MergeOrAppendTextSegment merges a given s into the last child of the parent if
// it can be merged, otherwise creates a new Text node and appends it to after current
// last child.
func MergeOrAppendTextSegment(parent Node, s textm.Segment) {
	last := parent.LastChild()
	t, ok := last.(*Text)
	if ok && !t.Value.IsOwned() && t.Value.Index().Stop == s.Start && !t.SoftLineBreak() {
		t.Value = textm.NewIndexValue(textm.NewIndex(t.Value.Index().Start, s.Stop))
	} else {
		parent.AppendChild(NewSegmentText(s))
	}
}

// MergeOrReplaceTextSegment merges a given s into a previous sibling of the node n
// if a previous sibling of the node n is *Text, otherwise replaces Node n with s.
func MergeOrReplaceTextSegment(parent Node, n Node, s textm.Segment) {
	prev := n.PreviousSibling()
	if t, ok := prev.(*Text); ok && !t.Value.IsOwned() && t.Value.Index().Stop == s.Start && !t.SoftLineBreak() {
		t.Value = textm.NewIndexValue(textm.NewIndex(t.Value.Index().Start, s.Stop))
		parent.RemoveChild(n)
	} else {
		parent.ReplaceChild(n, NewSegmentText(s))
	}
}

// A CodeSpan struct represents a code span of Markdown text.
type CodeSpan struct {
	BaseInline

	// Value holds the content of this code span.
	// The content is sourced from the raw Markdown text and may span multiple
	// source lines.
	Value textm.MultilineValue
}

// IsBlank returns true if this node consists of spaces, otherwise false.
func (n *CodeSpan) IsBlank(source []byte) bool {
	return util.IsBlank(n.Value.Bytes(source))
}

// Dump implements Node.Dump.
func (n *CodeSpan) Dump(source []byte, level int) {
	m := map[string]string{
		"Value": string(n.Value.Bytes(source)),
	}
	DumpHelper(n, source, level, m, nil)
}

// KindCodeSpan is a NodeKind of the CodeSpan node.
var KindCodeSpan = NewNodeKind("CodeSpan")

// Kind implements Node.Kind.
func (n *CodeSpan) Kind() NodeKind {
	return KindCodeSpan
}

// NewCodeSpan returns a new CodeSpan node with the given value.
func NewCodeSpan(value textm.MultilineValue) *CodeSpan {
	n := &CodeSpan{Value: value}
	n.Init(n)
	return n
}

// An Emphasis struct represents an emphasis of Markdown text.
// An Emphasis struct represents an emphasis of Markdown text (e.g. *text* or _text_).
type Emphasis struct {
	BaseInline
}

// Dump implements Node.Dump.
func (n *Emphasis) Dump(source []byte, level int) {
	DumpHelper(n, source, level, nil, nil)
}

// KindEmphasis is a NodeKind of the Emphasis node.
var KindEmphasis = NewNodeKind("Emphasis")

// Kind implements Node.Kind.
func (n *Emphasis) Kind() NodeKind {
	return KindEmphasis
}

// NewEmphasis returns a new Emphasis node.
func NewEmphasis() *Emphasis {
	n := &Emphasis{}
	n.Init(n)
	return n
}

// A Strong struct represents strong importance of Markdown text (e.g. **text** or __text__).
type Strong struct {
	BaseInline
}

// Dump implements Node.Dump.
func (n *Strong) Dump(source []byte, level int) {
	DumpHelper(n, source, level, nil, nil)
}

// KindStrong is a NodeKind of the Strong node.
var KindStrong = NewNodeKind("Strong")

// Kind implements Node.Kind.
func (n *Strong) Kind() NodeKind {
	return KindStrong
}

// NewStrong returns a new Strong node.
func NewStrong() *Strong {
	n := &Strong{}
	n.Init(n)
	return n
}

type baseLink struct {
	BaseInline

	// Destination is a destination(URL) of this link.
	Destination textm.Value

	// Title is a title of this link.
	Title textm.MultilineValue

	// Reference is a reference of this link. This field is used for reference links.
	// If this link is not a reference link, this field is nil.
	Reference *ReferenceLink
}

// LinkOption is an option for Link and Image nodes.
type LinkOption interface {
	SetLinkOption(*baseLink)
}

type linkTitle struct {
	value textm.MultilineValue
}

func (o *linkTitle) SetLinkOption(n *baseLink) {
	n.Title = o.value
}

// WithLinkTitle returns a LinkOption that sets the title of a link or image.
func WithLinkTitle[T textm.MultilineValueInput](title T) LinkOption {
	return &linkTitle{value: textm.NewMultilineValue(title)}
}

type linkReference struct {
	kind  ReferenceLinkKind
	value textm.MultilineValue
}

func (o *linkReference) SetLinkOption(n *baseLink) {
	n.Reference = &ReferenceLink{ReferenceLinkKind: o.kind, Value: o.value}
}

// WithLinkReference returns a LinkOption that sets the reference of a link or image.
func WithLinkReference(kind ReferenceLinkKind, value textm.MultilineValue) LinkOption {
	return &linkReference{kind: kind, value: value}
}

type autoLinkText struct {
	value textm.Value
}

func (o *autoLinkText) SetAutoLinkOption(n *AutoLink) {
	n.Text = o.value
}

// WithAutoLinkText returns an AutoLinkOption that sets the original source text of an autolink.
func WithAutoLinkText[T textm.ValueInput](text T) AutoLinkOption {
	return &autoLinkText{value: textm.NewValue(text)}
}

// ReferenceLinkKind defines a kind of reference link.
type ReferenceLinkKind int

const (
	// ReferenceLinkKindFull indicates that a reference link has a full reference like [foo][bar].
	ReferenceLinkKindFull ReferenceLinkKind = iota + 1
	// ReferenceLinkKindCollapsed indicates that a reference link has a collapsed reference like [foo][].
	ReferenceLinkKindCollapsed
	// ReferenceLinkKindShortcut indicates that a reference link has a shortcut reference like [foo].
	ReferenceLinkKindShortcut
)

// String returns a string representation of this reference link kind.
func (t ReferenceLinkKind) String() string {
	switch t {
	case ReferenceLinkKindFull:
		return "Full"
	case ReferenceLinkKindCollapsed:
		return "Collapsed"
	case ReferenceLinkKindShortcut:
		return "Shortcut"
	default:
		return fmt.Sprintf("Unknown(%d)", t)
	}
}

// ReferenceLink struct represents a reference link of the Markdown text.
type ReferenceLink struct {
	// ReferenceLinkKind is the kind of this reference link.
	ReferenceLinkKind ReferenceLinkKind

	// Value is a value of this reference link.
	Value textm.MultilineValue
}

// NewReferenceLink returns a new ReferenceLink with the given kind and value.
func NewReferenceLink(kind ReferenceLinkKind, value textm.MultilineValue) *ReferenceLink {
	return &ReferenceLink{
		ReferenceLinkKind: kind,
		Value:             value,
	}
}

// A Link struct represents a link of the Markdown text.
type Link struct {
	baseLink
}

// Dump implements Node.Dump.
func (n *Link) Dump(source []byte, level int) {
	m := map[string]string{}
	m["Destination"] = string(n.Destination.Bytes(source))
	title := n.Title.Bytes(source)
	if len(title) != 0 {
		m["Title"] = string(title)
	}
	cb := func(int) {}
	if n.Reference != nil {
		cb = func(level int) {
			indent := strings.Repeat("    ", level)
			fmt.Printf("%sReference {\n", indent)
			indent2 := strings.Repeat("    ", level+1)
			fmt.Printf("%sReferenceLinkKind: %s\n", indent2, n.Reference.ReferenceLinkKind.String())
			fmt.Printf("%sValue : %s\n", indent2, string(n.Reference.Value.Bytes(source)))
			fmt.Printf("%s}\n", indent)
		}
	}
	DumpHelper(n, source, level, m, cb)
}

// KindLink is a NodeKind of the Link node.
var KindLink = NewNodeKind("Link")

// Kind implements Node.Kind.
func (n *Link) Kind() NodeKind {
	return KindLink
}

// NewLink returns a new Link node with the given destination and options.
func NewLink(destination textm.Value, opts ...LinkOption) *Link {
	n := &Link{}
	n.Init(n)
	n.Destination = destination
	for _, opt := range opts {
		opt.SetLinkOption(&n.baseLink)
	}
	return n
}

// An Image struct represents an image of the Markdown text.
type Image struct {
	baseLink
}

// Dump implements Node.Dump.
func (n *Image) Dump(source []byte, level int) {
	m := map[string]string{}
	m["Destination"] = string(n.Destination.Bytes(source))
	title := n.Title.Bytes(source)
	if len(title) != 0 {
		m["Title"] = string(title)
	}
	cb := func(int) {}
	if n.Reference != nil {
		cb = func(level int) {
			indent := strings.Repeat("    ", level)
			fmt.Printf("%sReference {\n", indent)
			indent2 := strings.Repeat("    ", level+1)
			fmt.Printf("%sReferenceLinkKind: %s\n", indent2, n.Reference.ReferenceLinkKind.String())
			fmt.Printf("%sValue : %s\n", indent2, string(n.Reference.Value.Bytes(source)))
			fmt.Printf("%s}\n", indent)
		}
	}
	DumpHelper(n, source, level, m, cb)
}

// KindImage is a NodeKind of the Image node.
var KindImage = NewNodeKind("Image")

// Kind implements Node.Kind.
func (n *Image) Kind() NodeKind {
	return KindImage
}

// NewImage returns a new Image node with the given destination and options.
func NewImage(destination textm.Value, opts ...LinkOption) *Image {
	n := &Image{}
	n.Init(n)
	n.Destination = destination
	for _, opt := range opts {
		opt.SetLinkOption(&n.baseLink)
	}
	return n
}

// An AutoLink struct represents an autolink of the Markdown text.
type AutoLink struct {
	BaseInline

	// Destination is the URL used for the href attribute.
	// For email autolinks, it includes the "mailto:" prefix.
	Destination textm.Value

	// Label is the display text shown inside the link element.
	Label textm.Value

	// Text is the original text as parsed from source, including any
	// surrounding syntax characters (e.g. "<" and ">" for CommonMark autolinks).
	Text textm.Value
}

// AutoLinkOption is an option for AutoLink nodes.
type AutoLinkOption interface {
	SetAutoLinkOption(*AutoLink)
}

// Dump implements Node.Dump.
func (n *AutoLink) Dump(source []byte, level int) {
	m := map[string]string{
		"Destination": string(n.Destination.Bytes(source)),
		"Label":       string(n.Label.Bytes(source)),
		"Text":        string(n.Text.Bytes(source)),
	}
	DumpHelper(n, source, level, m, nil)
}

// KindAutoLink is a NodeKind of the AutoLink node.
var KindAutoLink = NewNodeKind("AutoLink")

// Kind implements Node.Kind.
func (n *AutoLink) Kind() NodeKind {
	return KindAutoLink
}

// NewAutoLink returns a new AutoLink node with the given destination, label, and options.
func NewAutoLink(destination, label textm.Value, opts ...AutoLinkOption) *AutoLink {
	n := &AutoLink{}
	n.Init(n)
	n.Destination = destination
	n.Label = label
	for _, opt := range opts {
		opt.SetAutoLinkOption(n)
	}
	return n
}

// A RawHTML struct represents an inline raw HTML of the Markdown text.
type RawHTML struct {
	BaseInline

	// Value holds the raw HTML content. Lines are concatenated with leading
	// whitespace of each continuation line trimmed.
	Value textm.MultilineValue
}

// Dump implements Node.Dump.
func (n *RawHTML) Dump(source []byte, level int) {
	m := map[string]string{
		"Value": string(n.Value.Bytes(source)),
	}
	DumpHelper(n, source, level, m, nil)
}

// KindRawHTML is a NodeKind of the RawHTML node.
var KindRawHTML = NewNodeKind("RawHTML")

// Kind implements Node.Kind.
func (n *RawHTML) Kind() NodeKind {
	return KindRawHTML
}

// NewRawHTML returns a new RawHTML node with the given value.
func NewRawHTML(value textm.MultilineValue) *RawHTML {
	n := &RawHTML{Value: value}
	n.Init(n)
	return n
}
