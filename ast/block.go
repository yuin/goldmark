package ast

import (
	"maps"

	textm "github.com/yuin/goldmark/v2/text"
)

const flagBlankPreviousLines = 1 << 0
const flagSingle = 1 << 1

// A BaseBlock struct implements the Node interface partially.
type BaseBlock struct {
	BaseNode
	// source holds raw source text that will be parsed into inline nodes.
	source []textm.Segment
	// single holds a single source segment for optimization when there is only one segment.
	single [1]textm.Segment

	flags uint8
}

// Implements BlockNode marker interface.
func (b *BaseBlock) blockNode() {}

// HasBlankPreviousLines implements Node.HasBlankPreviousLines.
func (b *BaseBlock) HasBlankPreviousLines() bool {
	return b.flags&flagBlankPreviousLines != 0
}

// SetBlankPreviousLines implements Node.SetBlankPreviousLines.
func (b *BaseBlock) SetBlankPreviousLines(v bool) {
	if v {
		b.flags |= flagBlankPreviousLines
	} else {
		b.flags &^= flagBlankPreviousLines
	}
}

// Source implements BlockNode.Source.
func (b *BaseBlock) Source() []textm.Segment {
	if b.flags&flagSingle != 0 {
		return b.single[:]
	}
	return b.source
}

// SetSource implements BlockNode.SetSource.
func (b *BaseBlock) SetSource(v []textm.Segment) {
	b.source = v
	b.flags &^= flagSingle
}

// AppendSource implements BlockNode.AppendSource.
func (b *BaseBlock) AppendSource(seg textm.Segment) {
	if b.source == nil {
		if b.flags&flagSingle == 0 {
			b.single[0] = seg
			b.flags |= flagSingle
			return
		}
		b.source = make([]textm.Segment, 0, 8)
		b.source = append(b.source, b.single[0])
		b.flags &^= flagSingle
	}
	b.source = append(b.source, seg)
}

// A Document struct is a root node of Markdown text.
type Document struct {
	BaseBlock

	metadata map[string]any
}

// KindDocument is a NodeKind of the Document node.
var KindDocument = NewNodeKind("Document")

// Dump implements Node.Dump .
func (n *Document) Dump(_ []byte) *NodeDump {
	return NewNodeDump(n, nil)
}

// Pos implements Node.Pos.
func (n *Document) Pos() int {
	return 0
}

// Kind implements Node.Kind.
func (n *Document) Kind() NodeKind {
	return KindDocument
}

// OwnerDocument implements Node.OwnerDocument.
func (n *Document) OwnerDocument() *Document {
	return n
}

// Metadata returns metadata of this document.
func (n *Document) Metadata() map[string]any {
	if n.metadata == nil {
		n.metadata = map[string]any{}
	}
	return n.metadata
}

// SetMetadata sets given metadata to this document.
func (n *Document) SetMetadata(meta map[string]any) {
	if n.metadata == nil {
		n.metadata = map[string]any{}
	}
	maps.Copy(n.metadata, meta)
}

// AddMeta adds given metadata to this document.
func (n *Document) AddMeta(key string, value any) {
	if n.metadata == nil {
		n.metadata = map[string]any{}
	}
	n.metadata[key] = value
}

// NewDocument returns a new Document node.
func NewDocument() *Document {
	n := &Document{
		metadata: nil,
	}
	n.Init(n)
	n.SetBlankPreviousLines(true)
	return n
}

// A Paragraph struct represents a paragraph of Markdown text.
type Paragraph struct {
	BaseBlock
}

// Dump implements Node.Dump .
func (n *Paragraph) Dump(_ []byte) *NodeDump {
	return NewNodeDump(n, nil)
}

// Pos implements Node.Pos.
func (n *Paragraph) Pos() int {
	if len(n.source) == 0 {
		return -1
	}
	return n.source[0].Start
}

// KindParagraph is a NodeKind of the Paragraph node.
var KindParagraph = NewNodeKind("Paragraph")

// Kind implements Node.Kind.
func (n *Paragraph) Kind() NodeKind {
	return KindParagraph
}

// NewParagraph returns a new Paragraph node.
func NewParagraph() *Paragraph {
	n := &Paragraph{}
	n.Init(n)
	return n
}

// IsParagraph returns true if the given node is a Paragraph node, otherwise false.
func IsParagraph(node Node) bool {
	return node != nil && node.Kind() == KindParagraph
}

// HeadingKind indicates whether a Heading is ATX or Setext.
type HeadingKind int

const (
	// HeadingKindATX represents an ATX heading (e.g. ## heading).
	HeadingKindATX HeadingKind = iota + 1
	// HeadingKindSetext represents a Setext heading (underline style).
	HeadingKindSetext
)

// String returns a human-readable name of the HeadingKind.
func (k HeadingKind) String() string {
	switch k {
	case HeadingKindATX:
		return "ATX"
	case HeadingKindSetext:
		return "Setext"
	default:
		return "Unknown"
	}
}

// A Heading struct represents headings like SetextHeading and ATXHeading.
type Heading struct {
	BaseBlock
	// Level returns a level of this heading.
	// This value is between 1 and 6.
	Level int
	// HeadingKind indicates whether this is an ATX or Setext heading.
	HeadingKind HeadingKind
}

// Dump implements Node.Dump .
func (n *Heading) Dump(_ []byte) *NodeDump {
	return NewNodeDump(n, map[string]any{
		"Level":       n.Level,
		"HeadingKind": n.HeadingKind.String(),
	})
}

// KindHeading is a NodeKind of the Heading node.
var KindHeading = NewNodeKind("Heading")

// Kind implements Node.Kind.
func (n *Heading) Kind() NodeKind {
	return KindHeading
}

// NewHeading returns a new Heading node.
func NewHeading(level int, kind HeadingKind) *Heading {
	n := &Heading{
		Level:       level,
		HeadingKind: kind,
	}
	n.Init(n)
	return n
}

// A ThematicBreak struct represents a thematic break of Markdown text.
type ThematicBreak struct {
	BaseBlock
}

// Dump implements Node.Dump .
func (n *ThematicBreak) Dump(_ []byte) *NodeDump {
	return NewNodeDump(n, nil)
}

// KindThematicBreak is a NodeKind of the ThematicBreak node.
var KindThematicBreak = NewNodeKind("ThematicBreak")

// Kind implements Node.Kind.
func (n *ThematicBreak) Kind() NodeKind {
	return KindThematicBreak
}

// NewThematicBreak returns a new ThematicBreak node.
func NewThematicBreak() *ThematicBreak {
	n := &ThematicBreak{}
	n.Init(n)
	return n
}

// CodeBlockKind indicates whether a CodeBlock is indented or fenced.
type CodeBlockKind int

const (
	// CodeBlockKindIndented indicates an indented code block (4-space or tab indent).
	CodeBlockKindIndented CodeBlockKind = iota + 1
	// CodeBlockKindFenced indicates a fenced code block (``` or ~~~).
	CodeBlockKindFenced
)

// String returns a human-readable name of the CodeBlockKind.
func (k CodeBlockKind) String() string {
	switch k {
	case CodeBlockKindIndented:
		return "Indented"
	case CodeBlockKindFenced:
		return "Fenced"
	default:
		return "Unknown"
	}
}

// A CodeBlock struct represents a code block of Markdown text.
type CodeBlock struct {
	BaseBlock

	// CodeBlockKind indicates whether this is an indented or fenced code block.
	CodeBlockKind CodeBlockKind

	// Info is the info string of a fenced code block (e.g. language identifier).
	// It is empty for indented code blocks.
	Info textm.SingleLineValue

	// Value holds the raw content of this code block for rendering.
	Value textm.Lines

	language *textm.SingleLineValue
}

// Language returns the language extracted from the info string.
// Language returns false if there is no info string.
func (n *CodeBlock) Language(source []byte) (textm.SingleLineValue, bool) {
	if n.language == nil {
		info := n.Info.Bytes(source)
		if len(info) == 0 {
			return textm.SingleLineValue{}, false
		}
		i := 0
		for ; i < len(info); i++ {
			if info[i] == ' ' {
				break
			}
		}
		if n.Info.IsOwned() {
			v := textm.NewSingleLineValue(info[:i])
			n.language = &v
		} else {
			v := textm.NewIndexSingleLineValue(textm.NewIndex(n.Info.Index().Start, n.Info.Index().Start+i))
			n.language = &v
		}
	}
	if n.language == nil {
		return textm.SingleLineValue{}, false
	}
	return *n.language, true
}

// Dump implements Node.Dump.
func (n *CodeBlock) Dump(source []byte) *NodeDump {
	m := map[string]any{
		"CodeBlockKind": n.CodeBlockKind.String(),
		"Value":         n.Value.Str(source),
	}
	info := n.Info.Bytes(source)
	if len(info) > 0 {
		m["Info"] = string(info)
	}
	return NewNodeDump(n, m)
}

// KindCodeBlock is a NodeKind of the CodeBlock node.
var KindCodeBlock = NewNodeKind("CodeBlock")

// Kind implements Node.Kind.
func (n *CodeBlock) Kind() NodeKind {
	return KindCodeBlock
}

// NewCodeBlock returns a new CodeBlock node with the given kind and value.
func NewCodeBlock(kind CodeBlockKind, value textm.Lines, opts ...CodeBlockOption) *CodeBlock {
	n := &CodeBlock{CodeBlockKind: kind, Value: value}
	n.Init(n)
	for _, opt := range opts {
		opt.setCodeBlockOption(n)
	}
	return n
}

// CodeBlockOption is an option for CodeBlock nodes.
type CodeBlockOption interface {
	setCodeBlockOption(*CodeBlock)
}

type codeBlockInfo struct {
	value textm.SingleLineValue
}

func (o *codeBlockInfo) setCodeBlockOption(n *CodeBlock) {
	n.Info = o.value
}

// WithCodeBlockInfo returns a CodeBlockOption that sets the info string of a fenced code block.
func WithCodeBlockInfo[T textm.SingleLineValueInput](info T) CodeBlockOption {
	return &codeBlockInfo{value: textm.NewSingleLineValue(info)}
}

// A Blockquote struct represents an blockquote block of Markdown text.
type Blockquote struct {
	BaseBlock
}

// Dump implements Node.Dump .
func (n *Blockquote) Dump(_ []byte) *NodeDump {
	return NewNodeDump(n, nil)
}

// KindBlockquote is a NodeKind of the Blockquote node.
var KindBlockquote = NewNodeKind("Blockquote")

// Kind implements Node.Kind.
func (n *Blockquote) Kind() NodeKind {
	return KindBlockquote
}

// NewBlockquote returns a new Blockquote node.
func NewBlockquote() *Blockquote {
	n := &Blockquote{}
	n.Init(n)
	return n
}

// A List struct represents a list of Markdown text.
type List struct {
	BaseBlock

	// Marker is a marker character like '-', '+', ')' and '.'.
	Marker byte

	// IsTight is a true if this list is a 'tight' list.
	// See https://spec.commonmark.org/0.30/#loose for details.
	IsTight bool

	// Start is an initial number of this ordered list.
	// If this list is not an ordered list, Start is 0.
	Start int
}

// IsOrdered returns true if this list is an ordered list, otherwise false.
func (l *List) IsOrdered() bool {
	return l.Marker == '.' || l.Marker == ')'
}

// CanContinue returns true if this list can continue with
// the given mark and a list type, otherwise false.
func (l *List) CanContinue(marker byte, isOrdered bool) bool {
	return marker == l.Marker && isOrdered == l.IsOrdered()
}

// Dump implements Node.Dump.
func (l *List) Dump(_ []byte) *NodeDump {
	m := map[string]any{
		"Ordered": l.IsOrdered(),
		"Marker":  string([]byte{l.Marker}),
		"Tight":   l.IsTight,
	}
	if l.IsOrdered() {
		m["Start"] = l.Start
	}
	return NewNodeDump(l, m)
}

// KindList is a NodeKind of the List node.
var KindList = NewNodeKind("List")

// Kind implements Node.Kind.
func (l *List) Kind() NodeKind {
	return KindList
}

// NewList returns a new List node.
func NewList(marker byte) *List {
	n := &List{
		Marker:  marker,
		IsTight: true,
	}
	n.Init(n)
	return n
}

// A ListItem struct represents a list item of Markdown text.
type ListItem struct {
	BaseBlock

	// offset is the content indentation offset used by the parser to continue
	// parsing subsequent lines of this list item. It is set by the list parser
	// and is not meaningful when constructing an AST programmatically.
	offset int
}

// Offset returns the content indentation offset used during parsing.
// This is set by the list parser and is not meaningful for programmatic construction.
func (n *ListItem) Offset() int { return n.offset }

// SetOffset sets the content indentation offset.
// This is called by the list parser during parsing.
func (n *ListItem) SetOffset(v int) { n.offset = v }

// Dump implements Node.Dump.
func (n *ListItem) Dump(_ []byte) *NodeDump {
	return NewNodeDump(n, nil)
}

// KindListItem is a NodeKind of the ListItem node.
var KindListItem = NewNodeKind("ListItem")

// Kind implements Node.Kind.
func (n *ListItem) Kind() NodeKind {
	return KindListItem
}

// NewListItem returns a new ListItem node.
func NewListItem() *ListItem {
	n := &ListItem{}
	n.Init(n)
	return n
}

// HTMLBlockKind represents kinds of an html block.
// See https://spec.commonmark.org/0.30/#html-blocks
type HTMLBlockKind int

const (
	// HTMLBlockKind1 represents type 1 html blocks.
	HTMLBlockKind1 HTMLBlockKind = iota + 1
	// HTMLBlockKind2 represents type 2 html blocks.
	HTMLBlockKind2
	// HTMLBlockKind3 represents type 3 html blocks.
	HTMLBlockKind3
	// HTMLBlockKind4 represents type 4 html blocks.
	HTMLBlockKind4
	// HTMLBlockKind5 represents type 5 html blocks.
	HTMLBlockKind5
	// HTMLBlockKind6 represents type 6 html blocks.
	HTMLBlockKind6
	// HTMLBlockKind7 represents type 7 html blocks.
	HTMLBlockKind7
)

// String returns a human-readable name of the HTMLBlockKind.
func (k HTMLBlockKind) String() string {
	switch k {
	case HTMLBlockKind1:
		return "Kind1"
	case HTMLBlockKind2:
		return "Kind2"
	case HTMLBlockKind3:
		return "Kind3"
	case HTMLBlockKind4:
		return "Kind4"
	case HTMLBlockKind5:
		return "Kind5"
	case HTMLBlockKind6:
		return "Kind6"
	case HTMLBlockKind7:
		return "Kind7"
	default:
		return "Unknown"
	}
}

// An HTMLBlock struct represents an html block of Markdown text.
type HTMLBlock struct {
	BaseBlock

	// HTMLBlockKind is the kind of this html block.
	HTMLBlockKind HTMLBlockKind

	// Value holds the raw HTML content of this block for rendering.
	Value textm.Lines
}

// Dump implements Node.Dump.
func (n *HTMLBlock) Dump(source []byte) *NodeDump {
	return NewNodeDump(n, map[string]any{
		"HTMLBlockKind": n.HTMLBlockKind.String(),
		"Value":         n.Value.Str(source),
	})
}

// KindHTMLBlock is a NodeKind of the HTMLBlock node.
var KindHTMLBlock = NewNodeKind("HTMLBlock")

// Kind implements Node.Kind.
func (n *HTMLBlock) Kind() NodeKind {
	return KindHTMLBlock
}

// NewHTMLBlock returns a new HTMLBlock node.
func NewHTMLBlock(kind HTMLBlockKind) *HTMLBlock {
	n := &HTMLBlock{
		HTMLBlockKind: kind,
	}
	n.Init(n)
	return n
}

// A LinkReferenceDefinition struct represents a list of Markdown text.
type LinkReferenceDefinition struct {
	BaseBlock

	// Label is a label of this link reference definition.
	Label textm.MultiLineValue

	// Destination is a destination of this link reference definition.
	Destination textm.SingleLineValue

	// Title is a title of this link reference definition.
	Title textm.MultiLineValue
}

// LinkReferenceDefinitionOption is an option for LinkReferenceDefinition nodes.
type LinkReferenceDefinitionOption interface {
	setLinkReferenceDefinitionOption(*LinkReferenceDefinition)
}

func (o *linkTitle) setLinkReferenceDefinitionOption(n *LinkReferenceDefinition) {
	n.Title = o.value
}

// Dump implements Node.Dump.
func (l *LinkReferenceDefinition) Dump(source []byte) *NodeDump {
	return NewNodeDump(l, map[string]any{
		"Label":       l.Label.Str(source),
		"Destination": l.Destination.Str(source),
		"Title":       l.Title.Str(source),
	})
}

// KindLinkReferenceDefinition is a NodeKind of the LinkReferenceDefinition node.
var KindLinkReferenceDefinition = NewNodeKind("LinkReferenceDefinition")

// Kind implements Node.Kind.
func (l *LinkReferenceDefinition) Kind() NodeKind {
	return KindLinkReferenceDefinition
}

// NewLinkReferenceDefinition returns a new LinkReferenceDefinition node.
func NewLinkReferenceDefinition(
	label textm.MultiLineValue, destination textm.SingleLineValue,
	opts ...LinkReferenceDefinitionOption) *LinkReferenceDefinition {
	n := &LinkReferenceDefinition{
		Label:       label,
		Destination: destination,
	}
	for _, opt := range opts {
		opt.setLinkReferenceDefinitionOption(n)
	}
	n.Init(n)
	return n
}
