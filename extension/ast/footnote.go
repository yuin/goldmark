package ast

import (
	gast "github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/text"
)

// A FootnoteReference struct represents a reference to a footnote definition
// in Markdown (PHP Markdown Extra) text.
type FootnoteReference struct {
	gast.BaseInline

	// Label is the label of the footnote reference (e.g. "1" for [^1]).
	Label text.Value

	// Index is the display index of the referenced FootnoteDefinition.
	// This is set by the Footnotes manager when the reference is added.
	Index int

	// RefIndex is the position of this reference among all references
	// to the same FootnoteDefinition (0-based).
	RefIndex int
}

// Dump implements Node.Dump.
func (n *FootnoteReference) Dump(source []byte) *gast.NodeDump {
	return gast.NewNodeDump(n, map[string]any{
		"Label":    string(n.Label.Bytes(source)),
		"Index":    n.Index,
		"RefIndex": n.RefIndex,
	})
}

// KindFootnoteReference is a NodeKind of the FootnoteReference node.
var KindFootnoteReference = gast.NewNodeKind("FootnoteReference")

// Kind implements Node.Kind.
func (n *FootnoteReference) Kind() gast.NodeKind {
	return KindFootnoteReference
}

// NewFootnoteReference returns a new FootnoteReference node.
func NewFootnoteReference(label text.Value) *FootnoteReference {
	n := &FootnoteReference{
		Label:    label,
		Index:    -1,
		RefIndex: -1,
	}
	n.Init(n)
	return n
}

// A FootnoteDefinition struct represents a footnote definition
// in Markdown (PHP Markdown Extra) text.
type FootnoteDefinition struct {
	gast.BaseBlock

	// Label is the label of the footnote definition (e.g. "1" for [^1]: ...).
	Label text.Value
}

// Dump implements Node.Dump.
func (n *FootnoteDefinition) Dump(source []byte) *gast.NodeDump {
	return gast.NewNodeDump(n, map[string]any{
		"Label": string(n.Label.Bytes(source)),
	})
}

// KindFootnoteDefinition is a NodeKind of the FootnoteDefinition node.
var KindFootnoteDefinition = gast.NewNodeKind("FootnoteDefinition")

// Kind implements Node.Kind.
func (n *FootnoteDefinition) Kind() gast.NodeKind {
	return KindFootnoteDefinition
}

// NewFootnoteDefinition returns a new FootnoteDefinition node.
func NewFootnoteDefinition(label text.Value) *FootnoteDefinition {
	n := &FootnoteDefinition{
		Label: label,
	}
	n.Init(n)
	return n
}
