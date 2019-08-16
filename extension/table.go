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

func (b *tableParagraphTransformer) Transform(node *gast.Paragraph, reader text.Reader, pc parser.Context) {
	lines := node.Lines()
	if lines.Len() < 2 {
		return
	}
	alignments := b.parseDelimiter(lines.At(1), reader)
	if alignments == nil {
		return
	}
	header := b.parseRow(lines.At(0), alignments, true, reader)
	if header == nil || len(alignments) != header.ChildCount() {
		return
	}
	table := ast.NewTable()
	table.Alignments = alignments
	table.AppendChild(table, ast.NewTableHeader(header))
	for i := 2; i < lines.Len(); i++ {
		table.AppendChild(table, b.parseRow(lines.At(i), alignments, false, reader))
	}
	node.Parent().InsertBefore(node.Parent(), node, table)
	node.Parent().RemoveChild(node.Parent(), node)
}

func (b *tableParagraphTransformer) parseRow(segment text.Segment, alignments []ast.Alignment, isHeader bool, reader text.Reader) *ast.TableRow {
	source := reader.Source()
	line := segment.Value(source)
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
	i := 0
	for ; pos < limit; i++ {
		alignment := ast.AlignNone
		if i >= len(alignments) {
			if !isHeader {
				return row
			}
		} else {
			alignment = alignments[i]
		}
		closure := util.FindClosure(line[pos:], byte(0), '|', true, false)
		if closure < 0 {
			closure = len(line[pos:])
		}
		node := ast.NewTableCell()
		segment := text.NewSegment(segment.Start+pos, segment.Start+pos+closure)
		segment = segment.TrimLeftSpace(source)
		segment = segment.TrimRightSpace(source)
		node.Lines().Append(segment)
		node.Alignment = alignment
		row.AppendChild(row, node)
		pos += closure + 1
	}
	for ; i < len(alignments); i++ {
		row.AppendChild(row, ast.NewTableCell())
	}
	return row
}

func (b *tableParagraphTransformer) parseDelimiter(segment text.Segment, reader text.Reader) []ast.Alignment {
	line := segment.Value(reader.Source())
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
			alignments = append(alignments, ast.AlignLeft)
		} else if tableDelimRight.Match(col) {
			alignments = append(alignments, ast.AlignRight)
		} else if tableDelimCenter.Match(col) {
			alignments = append(alignments, ast.AlignCenter)
		} else if tableDelimNone.Match(col) {
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

// RegisterFuncs implements renderer.NodeRenderer.RegisterFuncs.
func (r *TableHTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindTable, r.renderTable)
	reg.Register(ast.KindTableHeader, r.renderTableHeader)
	reg.Register(ast.KindTableRow, r.renderTableRow)
	reg.Register(ast.KindTableCell, r.renderTableCell)
}

func (r *TableHTMLRenderer) renderTable(w util.BufWriter, source []byte, n gast.Node, entering bool) (gast.WalkStatus, error) {
	if entering {
		_, _ = w.WriteString("<table>\n")
	} else {
		_, _ = w.WriteString("</table>\n")
	}
	return gast.WalkContinue, nil
}

func (r *TableHTMLRenderer) renderTableHeader(w util.BufWriter, source []byte, n gast.Node, entering bool) (gast.WalkStatus, error) {
	if entering {
		_, _ = w.WriteString("<thead>\n")
		_, _ = w.WriteString("<tr>\n")
	} else {
		_, _ = w.WriteString("</tr>\n")
		_, _ = w.WriteString("</thead>\n")
		if n.NextSibling() != nil {
			_, _ = w.WriteString("<tbody>\n")
		}
	}
	return gast.WalkContinue, nil
}

func (r *TableHTMLRenderer) renderTableRow(w util.BufWriter, source []byte, n gast.Node, entering bool) (gast.WalkStatus, error) {
	if entering {
		_, _ = w.WriteString("<tr>\n")
	} else {
		_, _ = w.WriteString("</tr>\n")
		if n.Parent().LastChild() == n {
			_, _ = w.WriteString("</tbody>\n")
		}
	}
	return gast.WalkContinue, nil
}

func (r *TableHTMLRenderer) renderTableCell(w util.BufWriter, source []byte, node gast.Node, entering bool) (gast.WalkStatus, error) {
	n := node.(*ast.TableCell)
	tag := "td"
	if n.Parent().Kind() == ast.KindTableHeader {
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
	return gast.WalkContinue, nil
}

type table struct {
}

// Table is an extension that allow you to use GFM tables .
var Table = &table{}

func (e *table) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithParagraphTransformers(
		util.Prioritized(NewTableParagraphTransformer(), 200),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(NewTableHTMLRenderer(), 500),
	))
}
