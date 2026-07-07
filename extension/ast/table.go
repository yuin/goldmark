package ast

import (
	gast "github.com/yuin/goldmark/v2/ast"
)

// Alignment is a text alignment of table cells.
type Alignment int

const (
	// AlignLeft indicates text should be left justified.
	AlignLeft Alignment = iota + 1

	// AlignRight indicates text should be right justified.
	AlignRight

	// AlignCenter indicates text should be centered.
	AlignCenter

	// AlignNone indicates text should be aligned by default manner.
	AlignNone
)

func (a Alignment) String() string {
	switch a {
	case AlignLeft:
		return "left"
	case AlignRight:
		return "right"
	case AlignCenter:
		return "center"
	case AlignNone:
		return "none"
	}
	return ""
}

// A Table struct represents a table of Markdown(GFM) text.
type Table struct {
	gast.BaseBlock
}

// Dump implements Node.Dump.
func (n *Table) Dump(_ []byte) *gast.NodeDump {
	return gast.NewNodeDump(n, nil)
}

// KindTable is a NodeKind of the Table node.
var KindTable = gast.NewNodeKind("Table")

// Kind implements Node.Kind.
func (n *Table) Kind() gast.NodeKind {
	return KindTable
}

// NewTable returns a new Table node.
func NewTable() *Table {
	n := &Table{}
	n.Init(n)
	return n
}

// A TableRow struct represents a table row of Markdown(GFM) text.
type TableRow struct {
	gast.BaseBlock
}

// Dump implements Node.Dump.
func (n *TableRow) Dump(_ []byte) *gast.NodeDump {
	return gast.NewNodeDump(n, nil)
}

// KindTableRow is a NodeKind of the TableRow node.
var KindTableRow = gast.NewNodeKind("TableRow")

// Kind implements Node.Kind.
func (n *TableRow) Kind() gast.NodeKind {
	return KindTableRow
}

// NewTableRow returns a new TableRow node.
func NewTableRow() *TableRow {
	n := &TableRow{}
	n.Init(n)
	return n
}

// A TableHeader struct represents a table header of Markdown(GFM) text.
type TableHeader struct {
	gast.BaseBlock
}

// KindTableHeader is a NodeKind of the TableHeader node.
var KindTableHeader = gast.NewNodeKind("TableHeader")

// Kind implements Node.Kind.
func (n *TableHeader) Kind() gast.NodeKind {
	return KindTableHeader
}

// Dump implements Node.Dump.
func (n *TableHeader) Dump(_ []byte) *gast.NodeDump {
	return gast.NewNodeDump(n, nil)
}

// NewTableHeader returns a new TableHeader node.
func NewTableHeader() *TableHeader {
	n := &TableHeader{}
	n.Init(n)
	return n
}

// A TableCell struct represents a table cell of a Markdown(GFM) text.
type TableCell struct {
	gast.BaseBlock
	Alignment Alignment
}

// Dump implements Node.Dump.
func (n *TableCell) Dump(_ []byte) *gast.NodeDump {
	return gast.NewNodeDump(n, map[string]any{
		"Alignment": n.Alignment.String(),
	})
}

// KindTableCell is a NodeKind of the TableCell node.
var KindTableCell = gast.NewNodeKind("TableCell")

// Kind implements Node.Kind.
func (n *TableCell) Kind() gast.NodeKind {
	return KindTableCell
}

// NewTableCell returns a new TableCell node with the given alignment.
func NewTableCell(alignment Alignment) *TableCell {
	n := &TableCell{Alignment: alignment}
	n.Init(n)
	return n
}

// A TableBody struct represents a table body of Markdown(GFM) text.
type TableBody struct {
	gast.BaseBlock
}

// Dump implements Node.Dump.
func (n *TableBody) Dump(_ []byte) *gast.NodeDump {
	return gast.NewNodeDump(n, nil)
}

// KindTableBody is a NodeKind of the TableBody node.
var KindTableBody = gast.NewNodeKind("TableBody")

// Kind implements Node.Kind.
func (n *TableBody) Kind() gast.NodeKind {
	return KindTableBody
}

// NewTableBody returns a new TableBody node.
func NewTableBody() *TableBody {
	n := &TableBody{}
	n.Init(n)
	return n
}
