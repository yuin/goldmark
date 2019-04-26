package ast

import (
	"fmt"
	gast "github.com/yuin/goldmark/ast"
	"strings"
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

	// Alignments returns alignments of the columns.
	Alignments []Alignment
}

// Dump implements Node.Dump
func (n *Table) Dump(source []byte, level int) {
	gast.DumpHelper(n, source, level, "Table", nil, func(level int) {
		indent := strings.Repeat("    ", level)
		fmt.Printf("%sAlignments {\n", indent)
		for i, alignment := range n.Alignments {
			indent2 := strings.Repeat("    ", level+1)
			fmt.Printf("%s%s", indent2, alignment.String())
			if i != len(n.Alignments)-1 {
				fmt.Println("")
			}
		}
		fmt.Printf("\n%s}\n", indent)
	})
}

// NewTable returns a new Table node.
func NewTable() *Table {
	return &Table{
		Alignments: []Alignment{},
	}
}

// A TableRow struct represents a table row of Markdown(GFM) text.
type TableRow struct {
	gast.BaseBlock
	Alignments []Alignment
}

// Dump implements Node.Dump.
func (n *TableRow) Dump(source []byte, level int) {
	gast.DumpHelper(n, source, level, "TableRow", nil, nil)
}

// NewTableRow returns a new TableRow node.
func NewTableRow(alignments []Alignment) *TableRow {
	return &TableRow{}
}

// A TableHeader struct represents a table header of Markdown(GFM) text.
type TableHeader struct {
	*TableRow
}

// NewTableHeader returns a new TableHeader node.
func NewTableHeader(row *TableRow) *TableHeader {
	return &TableHeader{row}
}

// A TableCell struct represents a table cell of a Markdown(GFM) text.
type TableCell struct {
	gast.BaseBlock
	Alignment Alignment
}

// Dump implements Node.Dump.
func (n *TableCell) Dump(source []byte, level int) {
	gast.DumpHelper(n, source, level, "TableCell", nil, nil)
}

// NewTableCell returns a new TableCell node.
func NewTableCell() *TableCell {
	return &TableCell{
		Alignment: AlignNone,
	}
}
