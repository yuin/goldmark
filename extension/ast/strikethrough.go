// Package ast defines AST nodes that represents extension's elements
package ast

import (
	gast "github.com/yuin/goldmark/ast"
)

// A Strikethrough struct represents a strikethrough of GFM text.
type Strikethrough struct {
	gast.BaseInline
}

func (n *Strikethrough) Inline() {
}

func (n *Strikethrough) Dump(source []byte, level int) {
	gast.DumpHelper(n, source, level, "Strikethrough", nil, nil)
}

// NewStrikethrough returns a new Strikethrough node.
func NewStrikethrough() *Strikethrough {
	return &Strikethrough{}
}
