package ast

import (
	gast "github.com/yuin/goldmark/ast"
)

// A TypographicText struct represents text that
// typographic text replaces certain punctuations.
type TypographicText struct {
	gast.BaseInline
	Value []byte
}

// Dump implements Node.Dump.
func (n *TypographicText) Dump(source []byte, level int) {
	gast.DumpHelper(n, source, level, nil, nil)
}

// KindTypographicText is a NodeKind of the TypographicText node.
var KindTypographicText = gast.NewNodeKind("TypographicText")

// Kind implements Node.Kind.
func (n *TypographicText) Kind() gast.NodeKind {
	return KindTypographicText
}

// NewTypographicText returns a new TypographicText node.
func NewTypographicText(value []byte) *TypographicText {
	return &TypographicText{
		Value: value,
	}
}
