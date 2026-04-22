package ast

import (
	gast "github.com/yuin/goldmark/v2/ast"
)

// A DefinitionList struct represents a definition list of Markdown
// (PHPMarkdownExtra) text.
type DefinitionList struct {
	gast.BaseBlock
	offset             int
	temporaryParagraph *gast.Paragraph
}

// Offset returns the indentation offset of this definition list (parser-internal).
func (n *DefinitionList) Offset() int { return n.offset }

// SetOffset sets the indentation offset of this definition list (parser-internal).
func (n *DefinitionList) SetOffset(v int) { n.offset = v }

// TemporaryParagraph returns the temporary paragraph associated with this definition list (parser-internal).
func (n *DefinitionList) TemporaryParagraph() *gast.Paragraph { return n.temporaryParagraph }

// SetTemporaryParagraph sets the temporary paragraph for this definition list (parser-internal).
func (n *DefinitionList) SetTemporaryParagraph(p *gast.Paragraph) { n.temporaryParagraph = p }

// Dump implements Node.Dump.
func (n *DefinitionList) Dump(source []byte, level int) {
	gast.DumpHelper(n, source, level, nil, nil)
}

// Pos implements Node.Pos.
func (n *DefinitionList) Pos() int {
	if n.FirstChild() != nil {
		return n.FirstChild().Pos()
	}
	return -1
}

// KindDefinitionList is a NodeKind of the DefinitionList node.
var KindDefinitionList = gast.NewNodeKind("DefinitionList")

// Kind implements Node.Kind.
func (n *DefinitionList) Kind() gast.NodeKind {
	return KindDefinitionList
}

// NewDefinitionList returns a new DefinitionList node.
func NewDefinitionList() *DefinitionList {
	n := &DefinitionList{}
	n.Init(n)
	return n
}

// A DefinitionTerm struct represents a definition list term of Markdown
// (PHPMarkdownExtra) text.
type DefinitionTerm struct {
	gast.BaseBlock
}

// Dump implements Node.Dump.
func (n *DefinitionTerm) Dump(source []byte, level int) {
	gast.DumpHelper(n, source, level, nil, nil)
}

// Pos implements Node.Pos.
func (n *DefinitionTerm) Pos() int {
	if len(n.Source()) == 0 {
		return -1
	}
	return n.Source()[0].Start
}

// KindDefinitionTerm is a NodeKind of the DefinitionTerm node.
var KindDefinitionTerm = gast.NewNodeKind("DefinitionTerm")

// Kind implements Node.Kind.
func (n *DefinitionTerm) Kind() gast.NodeKind {
	return KindDefinitionTerm
}

// NewDefinitionTerm returns a new DefinitionTerm node.
func NewDefinitionTerm() *DefinitionTerm {
	n := &DefinitionTerm{}
	n.Init(n)
	return n
}

// A DefinitionDescription struct represents a definition list description of Markdown
// (PHPMarkdownExtra) text.
type DefinitionDescription struct {
	gast.BaseBlock
	IsTight bool
}

// Dump implements Node.Dump.
func (n *DefinitionDescription) Dump(source []byte, level int) {
	gast.DumpHelper(n, source, level, nil, nil)
}

// KindDefinitionDescription is a NodeKind of the DefinitionDescription node.
var KindDefinitionDescription = gast.NewNodeKind("DefinitionDescription")

// Kind implements Node.Kind.
func (n *DefinitionDescription) Kind() gast.NodeKind {
	return KindDefinitionDescription
}

// NewDefinitionDescription returns a new DefinitionDescription node.
func NewDefinitionDescription() *DefinitionDescription {
	n := &DefinitionDescription{}
	n.Init(n)
	return n
}
