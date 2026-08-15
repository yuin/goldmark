package extension

import (
	"io"

	gast "github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/extension/ast"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer"
	"github.com/yuin/goldmark/v2/renderer/html"
	"github.com/yuin/goldmark/v2/text"
	"github.com/yuin/goldmark/v2/util"
)

type definitionListParser struct {
}

var defaultDefinitionListParser = &definitionListParser{}

func newDefinitionListParser() parser.BlockParser {
	return defaultDefinitionListParser
}

func (b *definitionListParser) Trigger() []byte {
	return []byte{':'}
}

func (b *definitionListParser) Open(parent gast.Node, reader text.Reader, pc parser.Context) (gast.Node, parser.State) {
	if _, ok := parent.(*ast.DefinitionList); ok {
		return nil, parser.NoChildren
	}
	line, _ := reader.PeekLine()
	pos := pc.BlockOffset()
	indent := pc.BlockIndent()
	if pos < 0 || line[pos] != ':' || indent != 0 {
		return nil, parser.NoChildren
	}

	last := parent.LastChild()
	// need 1 or more spaces after ':'
	w, _ := util.IndentWidth(line[pos+1:], pos+1)
	if w < 1 {
		return nil, parser.NoChildren
	}
	if w >= 8 { // starts with indented code
		w = 5
	}
	w += pos + 1 /* 1 = ':' */

	para, lastIsParagraph := last.(*gast.Paragraph)
	var list *ast.DefinitionList
	status := parser.HasChildren
	var ok bool
	if lastIsParagraph {
		list, ok = last.PreviousSibling().(*ast.DefinitionList)
		if ok { // is not first item
			list.SetOffset(w)
			list.SetTemporaryParagraph(para)
		} else { // is first item
			list = ast.NewDefinitionList()
			list.SetOffset(w)
			list.SetTemporaryParagraph(para)
			status |= parser.RequireParagraph
		}
	} else if list, ok = last.(*ast.DefinitionList); ok { // multiple description
		list.SetOffset(w)
		list.SetTemporaryParagraph(nil)
	} else {
		return nil, parser.NoChildren
	}

	return list, status
}

func (b *definitionListParser) Continue(node gast.Node, reader text.Reader, _ parser.Context) parser.State {
	line, _ := reader.PeekLine()
	if util.IsBlank(line) {
		return parser.Continue | parser.HasChildren
	}
	list, _ := node.(*ast.DefinitionList)
	w, _ := util.IndentWidth(line, reader.LineOffset())
	if w < list.Offset() {
		return parser.Close
	}
	pos, padding := util.IndentPosition(line, reader.LineOffset(), list.Offset())
	reader.AdvanceAndSetPadding(pos, padding)
	return parser.Continue | parser.HasChildren
}

func (b *definitionListParser) Close(_ gast.Node, _ text.Reader, _ parser.Context) {
	// nothing to do
}

func (b *definitionListParser) CanInterruptParagraph() bool {
	return true
}

func (b *definitionListParser) CanAcceptIndentedLine() bool {
	return false
}

type definitionDescriptionParser struct {
}

var defaultDefinitionDescriptionParser = &definitionDescriptionParser{}

func newDefinitionDescriptionParser() parser.BlockParser {
	return defaultDefinitionDescriptionParser
}

func (b *definitionDescriptionParser) Trigger() []byte {
	return []byte{':'}
}

func (b *definitionDescriptionParser) Open(
	parent gast.Node, reader text.Reader, pc parser.Context) (gast.Node, parser.State) {
	line, _ := reader.PeekLine()
	pos := pc.BlockOffset()
	indent := pc.BlockIndent()
	if pos < 0 || line[pos] != ':' || indent != 0 {
		return nil, parser.NoChildren
	}
	list, _ := parent.(*ast.DefinitionList)
	if list == nil {
		return nil, parser.NoChildren
	}
	para := list.TemporaryParagraph()
	list.SetTemporaryParagraph(nil)
	if para != nil {
		lines := para.Source()
		l := len(lines)
		for i := range l {
			term := ast.NewDefinitionTerm()
			term.AppendSource(lines[i].TrimRightSpace(reader.Source()))
			list.AppendChild(term)
		}
		para.Parent().RemoveChild(para)
	}
	cpos, padding := util.IndentPosition(line[pos+1:], pos+1, list.Offset()-pos-1)
	reader.AdvanceAndSetPadding(cpos+1, padding)

	return ast.NewDefinitionDescription(), parser.HasChildren
}

func (b *definitionDescriptionParser) Continue(_ gast.Node, _ text.Reader, _ parser.Context) parser.State {
	// definitionListParser detects end of the description.
	// so this method will never be called.
	return parser.Continue | parser.HasChildren
}

func (b *definitionDescriptionParser) Close(node gast.Node, _ text.Reader, _ parser.Context) {
	desc := node.(*ast.DefinitionDescription)
	desc.IsTight = !desc.HasBlankPreviousLines()
}

func (b *definitionDescriptionParser) CanInterruptParagraph() bool {
	return true
}

func (b *definitionDescriptionParser) CanAcceptIndentedLine() bool {
	return false
}

type definitionListHTMLRendererExtension struct {
}

// NewDefinitionListHTMLRenderer returns a new html.Extension for rendering DefinitionList nodes.
func NewDefinitionListHTMLRenderer() html.Extension {
	return &definitionListHTMLRendererExtension{}
}

func (r *definitionListHTMLRendererExtension) RendererOptions(_ *html.Config) []html.Option {
	return []html.Option{
		html.WithNodeRenderers(map[gast.NodeKind]html.NodeRenderer{
			ast.KindDefinitionList:        html.NodeRendererFunc(r.renderDefinitionList),
			ast.KindDefinitionTerm:        html.NodeRendererFunc(r.renderDefinitionTerm),
			ast.KindDefinitionDescription: html.NodeRendererFunc(r.renderDefinitionDescription),
		}),
		html.WithIsInTightBlockFunc(definitionListIsInTightBlock),
	}
}

// DefinitionListAttributeFilter defines attribute names which dl elements can have.
var DefinitionListAttributeFilter = html.GlobalAttributeFilter

func (r *definitionListHTMLRendererExtension) renderDefinitionList(
	writer io.Writer, source []byte, n gast.Node, entering bool, rc renderer.Context) (gast.WalkStatus, error) {
	w := writer.(util.BufWriter)
	if entering {
		if n.Attributes() != nil {
			_, _ = w.WriteString("<dl")
			html.RenderAttributes(w, source, n, DefinitionListAttributeFilter, rc)
			_, _ = w.WriteString(">\n")
		} else {
			_, _ = w.WriteString("<dl>\n")
		}
	} else {
		_, _ = w.WriteString("</dl>\n")
	}
	return gast.WalkContinue, nil
}

// DefinitionTermAttributeFilter defines attribute names which dt elements can have.
var DefinitionTermAttributeFilter = html.GlobalAttributeFilter

func (r *definitionListHTMLRendererExtension) renderDefinitionTerm(
	writer io.Writer, source []byte, n gast.Node, entering bool, rc renderer.Context) (gast.WalkStatus, error) {
	w := writer.(util.BufWriter)
	if entering {
		if n.Attributes() != nil {
			_, _ = w.WriteString("<dt")
			html.RenderAttributes(w, source, n, DefinitionTermAttributeFilter, rc)
			_ = w.WriteByte('>')
		} else {
			_, _ = w.WriteString("<dt>")
		}
	} else {
		_, _ = w.WriteString("</dt>\n")
	}
	return gast.WalkContinue, nil
}

// DefinitionDescriptionAttributeFilter defines attribute names which dd elements can have.
var DefinitionDescriptionAttributeFilter = html.GlobalAttributeFilter

func (r *definitionListHTMLRendererExtension) renderDefinitionDescription(
	writer io.Writer, source []byte, node gast.Node, entering bool, rc renderer.Context) (gast.WalkStatus, error) {
	w := writer.(util.BufWriter)
	if entering {
		n := node.(*ast.DefinitionDescription)
		_, _ = w.WriteString("<dd")
		if n.Attributes() != nil {
			html.RenderAttributes(w, source, n, DefinitionDescriptionAttributeFilter, rc)
		}
		if n.IsTight {
			_, _ = w.WriteString(">")
		} else {
			_, _ = w.WriteString(">\n")
		}
	} else {
		_, _ = w.WriteString("</dd>\n")
	}
	return gast.WalkContinue, nil
}

type definitionListParserExtension struct {
}

// NewDefinitionListParser returns a new parser.Extension for parsing PHP Markdown Extra Definition lists.
func NewDefinitionListParser() parser.Extension {
	return &definitionListParserExtension{}
}

func (e *definitionListParserExtension) ParserOptions(_ *parser.Config) []parser.Option {
	return []parser.Option{
		parser.WithBlockParsers(
			util.Prioritized(newDefinitionListParser(), 101),
			util.Prioritized(newDefinitionDescriptionParser(), 102),
		),
	}
}

func definitionListIsInTightBlock(n gast.Node) bool {
	parent := n.Parent()
	if parent == nil {
		return false
	}
	if desc, ok := parent.(*ast.DefinitionDescription); ok {
		return desc.IsTight
	}
	if gp := parent.Parent(); gp != nil {
		if list, ok := gp.(*gast.List); ok {
			return list.IsTight
		}
	}
	return false
}
