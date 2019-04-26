package extension

import (
	"bytes"
	"fmt"
	"regexp"

	"github.com/yuin/goldmark"
	gast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

var tableDelimRegexp = regexp.MustCompile(`^[\s\-\|\:]+$`)
var tableDelimLeft = regexp.MustCompile(`^\s*\:\-+\s*$`)
var tableDelimRight = regexp.MustCompile(`^\s*\-+\:\s*$`)
var tableDelimCenter = regexp.MustCompile(`^\s*\:\-+\:\s*$`)
var tableDelimNone = regexp.MustCompile(`^\s*\-+\s*$`)

type tableParagraphTransformer struct {
}

var defaultTableParagraphTransformer = &tableParagraphTransformer{}

// NewTableParagraphTransformer returns  a new ParagraphTransformer
// that can transform pargraphs into tables.
func NewTableParagraphTransformer() parser.ParagraphTransformer {
	return defaultTableParagraphTransformer
}

func (b *tableParagraphTransformer) Transform(node *gast.Paragraph, pc parser.Context) {
	lines := node.Lines()
	if lines.Len() < 2 {
		return
	}
	alignments := b.parseDelimiter(lines.At(1), pc)
	if alignments == nil {
		return
	}
	header := b.parseRow(lines.At(0), alignments, pc)
	if header == nil || len(alignments) != header.ChildCount() {
		return
	}
	table := ast.NewTable()
	table.Alignments = alignments
	table.AppendChild(table, ast.NewTableHeader(header))
	if lines.Len() > 2 {
		for i := 2; i < lines.Len(); i++ {
			table.AppendChild(table, b.parseRow(lines.At(i), alignments, pc))
		}
	}
	node.Parent().InsertBefore(node.Parent(), node, table)
	node.Parent().RemoveChild(node.Parent(), node)
	return
}

func (b *tableParagraphTransformer) parseRow(segment text.Segment, alignments []ast.Alignment, pc parser.Context) *ast.TableRow {
	line := segment.Value(pc.Source())
	pos := 0
	pos += util.TrimLeftSpaceLength(line)
	limit := len(line)
	limit -= util.TrimRightSpaceLength(line)
	row := ast.NewTableRow(alignments)
	if len(line) > 0 && line[pos] == '|' {
		pos++
	}
	if len(line) > 0 && line[limit-1] == '|' {
		limit--
	}
	for i := 0; pos < limit; i++ {
		closure := util.FindClosure(line[pos:], byte(0), '|', true, false)
		if closure < 0 {
			closure = len(line[pos:])
		}
		node := ast.NewTableCell()
		segment := text.NewSegment(segment.Start+pos, segment.Start+pos+closure)
		segment = segment.TrimLeftSpace(pc.Source())
		segment = segment.TrimRightSpace(pc.Source())
		node.Lines().Append(segment)
		node.Alignment = alignments[i]
		row.AppendChild(row, node)
		pos += closure + 1
	}
	return row
}

func (b *tableParagraphTransformer) parseDelimiter(segment text.Segment, pc parser.Context) []ast.Alignment {
	line := segment.Value(pc.Source())
	if !tableDelimRegexp.Match(line) {
		return nil
	}
	cols := bytes.Split(line, []byte{'|'})
	if util.IsBlank(cols[0]) {
		cols = cols[1:]
	}
	if len(cols) > 0 && util.IsBlank(cols[len(cols)-1]) {
		cols = cols[:len(cols)-1]
	}

	var alignments []ast.Alignment
	for _, col := range cols {
		if tableDelimLeft.Match(col) {
			if alignments == nil {
				alignments = []ast.Alignment{}
			}
			alignments = append(alignments, ast.AlignLeft)
		} else if tableDelimRight.Match(col) {
			if alignments == nil {
				alignments = []ast.Alignment{}
			}
			alignments = append(alignments, ast.AlignRight)
		} else if tableDelimCenter.Match(col) {
			if alignments == nil {
				alignments = []ast.Alignment{}
			}
			alignments = append(alignments, ast.AlignCenter)
		} else if tableDelimNone.Match(col) {
			if alignments == nil {
				alignments = []ast.Alignment{}
			}
			alignments = append(alignments, ast.AlignNone)
		} else {
			return nil
		}
	}
	return alignments
}

// TableHTMLRenderer is a renderer.NodeRenderer implementation that
// renders Table nodes.
type TableHTMLRenderer struct {
	html.Config
}

// NewTableHTMLRenderer returns a new TableHTMLRenderer.
func NewTableHTMLRenderer(opts ...html.Option) renderer.NodeRenderer {
	r := &TableHTMLRenderer{
		Config: html.NewConfig(),
	}
	for _, opt := range opts {
		opt.SetHTMLOption(&r.Config)
	}
	return r
}

// Render implements renderer.Renderer.Render.
func (r *TableHTMLRenderer) Render(writer util.BufWriter, source []byte, n gast.Node, entering bool) (gast.WalkStatus, error) {
	switch node := n.(type) {
	case *ast.Table:
		return r.renderTable(writer, source, node, entering), nil
	case *ast.TableHeader:
		return r.renderTableHeader(writer, source, node, entering), nil
	case *ast.TableRow:
		return r.renderTableRow(writer, source, node, entering), nil
	case *ast.TableCell:
		return r.renderTableCell(writer, source, node, entering), nil
	}
	return gast.WalkContinue, renderer.NotSupported
}

func (r *TableHTMLRenderer) renderTable(w util.BufWriter, source []byte, n *ast.Table, entering bool) gast.WalkStatus {
	if entering {
		w.WriteString("<table>\n")
	} else {
		w.WriteString("</table>\n")
	}
	return gast.WalkContinue
}

func (r *TableHTMLRenderer) renderTableHeader(w util.BufWriter, source []byte, n *ast.TableHeader, entering bool) gast.WalkStatus {
	if entering {
		w.WriteString("<thead>\n")
		w.WriteString("<tr>\n")
	} else {
		w.WriteString("</tr>\n")
		w.WriteString("</thead>\n")
		if n.NextSibling() != nil {
			w.WriteString("<tbody>\n")
		}
		if n.Parent().LastChild() == n {
			w.WriteString("</tbody>\n")
		}
	}
	return gast.WalkContinue
}

func (r *TableHTMLRenderer) renderTableRow(w util.BufWriter, source []byte, n *ast.TableRow, entering bool) gast.WalkStatus {
	if entering {
		w.WriteString("<tr>\n")
	} else {
		w.WriteString("</tr>\n")
		if n.Parent().LastChild() == n {
			w.WriteString("</tbody>\n")
		}
	}
	return gast.WalkContinue
}

func (r *TableHTMLRenderer) renderTableCell(w util.BufWriter, source []byte, n *ast.TableCell, entering bool) gast.WalkStatus {
	tag := "td"
	if n.Parent().Parent().FirstChild() == n.Parent() {
		tag = "th"
	}
	if entering {
		align := ""
		if n.Alignment != ast.AlignNone {
			align = fmt.Sprintf(` align="%s"`, n.Alignment.String())
		}
		fmt.Fprintf(w, "<%s%s>", tag, align)
	} else {
		fmt.Fprintf(w, "</%s>\n", tag)
	}
	return gast.WalkContinue
}

type table struct {
}

// Table is an extension that allow you to use GFM tables .
var Table = &table{}

func (e *table) Extend(m goldmark.Markdown) {
	m.Parser().AddOption(parser.WithParagraphTransformers(
		util.Prioritized(NewTableParagraphTransformer(), 200),
	))
	m.Renderer().AddOption(renderer.WithNodeRenderers(
		util.Prioritized(NewTableHTMLRenderer(), 500),
	))
}
