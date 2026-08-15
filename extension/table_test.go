package extension

import (
	"testing"

	"github.com/yuin/goldmark/v2/ast"
	east "github.com/yuin/goldmark/v2/extension/ast"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer/html"
	"github.com/yuin/goldmark/v2/testutil"
	"github.com/yuin/goldmark/v2/text"
	"github.com/yuin/goldmark/v2/util"
)

func TestTable(t *testing.T) {
	markdown := testutil.NewMarkdownToStringFunc(
		parser.New(
			parser.WithExtensions(NewTableParser()),
		),
		html.New(
			html.WithUnsafe(),
			html.WithXHTML(),
			html.WithExtensions(NewTableHTMLRenderer()),
		),
	)
	testutil.DoTestCaseFile(markdown, "_test/table.txt", t, testutil.ParseCliCaseArg()...)
}

func TestTableWithAlignDefault(t *testing.T) {
	markdown := testutil.NewMarkdownToStringFunc(
		parser.New(
			parser.WithExtensions(NewTableParser()),
		),
		html.New(
			html.WithXHTML(),
			html.WithUnsafe(),
			html.WithExtensions(NewTableHTMLRenderer(
				WithTableCellAlignMethod(TableCellAlignDefault),
			)),
		),
	)
	testutil.DoTestCase(
		markdown,
		testutil.MarkdownTestCase{
			No:          1,
			Description: "Cell with TableCellAlignDefault and XHTML should be rendered as an align attribute",
			Markdown: `
| abc | defghi |
:-: | -----------:
bar | baz
`,
			Expected: `<table>
<thead>
<tr>
<th align="center">abc</th>
<th align="right">defghi</th>
</tr>
</thead>
<tbody>
<tr>
<td align="center">bar</td>
<td align="right">baz</td>
</tr>
</tbody>
</table>`,
		},
		t,
	)

	markdown = testutil.NewMarkdownToStringFunc(
		parser.New(
			parser.WithExtensions(NewTableParser()),
		),
		html.New(
			html.WithUnsafe(),
			html.WithExtensions(NewTableHTMLRenderer(
				WithTableCellAlignMethod(TableCellAlignDefault),
			)),
		),
	)
	testutil.DoTestCase(
		markdown,
		testutil.MarkdownTestCase{
			No:          2,
			Description: "Cell with TableCellAlignDefault and HTML5 should be rendered as a style attribute",
			Markdown: `
| abc | defghi |
:-: | -----------:
bar | baz
`,
			Expected: `<table>
<thead>
<tr>
<th style="text-align:center">abc</th>
<th style="text-align:right">defghi</th>
</tr>
</thead>
<tbody>
<tr>
<td style="text-align:center">bar</td>
<td style="text-align:right">baz</td>
</tr>
</tbody>
</table>`,
		},
		t,
	)
}

func TestTableWithAlignAttribute(t *testing.T) {
	markdown := testutil.NewMarkdownToStringFunc(
		parser.New(
			parser.WithExtensions(NewTableParser()),
		),
		html.New(
			html.WithXHTML(),
			html.WithUnsafe(),
			html.WithExtensions(NewTableHTMLRenderer(
				WithTableCellAlignMethod(TableCellAlignAttribute),
			)),
		),
	)
	testutil.DoTestCase(
		markdown,
		testutil.MarkdownTestCase{
			No:          1,
			Description: "Cell with TableCellAlignAttribute and XHTML should be rendered as an align attribute",
			Markdown: `
| abc | defghi |
:-: | -----------:
bar | baz
`,
			Expected: `<table>
<thead>
<tr>
<th align="center">abc</th>
<th align="right">defghi</th>
</tr>
</thead>
<tbody>
<tr>
<td align="center">bar</td>
<td align="right">baz</td>
</tr>
</tbody>
</table>`,
		},
		t,
	)

	markdown = testutil.NewMarkdownToStringFunc(
		parser.New(
			parser.WithExtensions(NewTableParser()),
		),
		html.New(
			html.WithUnsafe(),
			html.WithExtensions(NewTableHTMLRenderer(
				WithTableCellAlignMethod(TableCellAlignAttribute),
			)),
		),
	)
	testutil.DoTestCase(
		markdown,
		testutil.MarkdownTestCase{
			No:          2,
			Description: "Cell with TableCellAlignAttribute and HTML5 should be rendered as an align attribute",
			Markdown: `
| abc | defghi |
:-: | -----------:
bar | baz
`,
			Expected: `<table>
<thead>
<tr>
<th align="center">abc</th>
<th align="right">defghi</th>
</tr>
</thead>
<tbody>
<tr>
<td align="center">bar</td>
<td align="right">baz</td>
</tr>
</tbody>
</table>`,
		},
		t,
	)
}

type tableStyleTransformer struct {
}

func (a *tableStyleTransformer) Transform(node *ast.Document, _ text.Reader, _ parser.Context) {
	cell := node.FirstChild().FirstChild().FirstChild().(*east.TableCell)
	cell.SetAttribute("style", text.NewMultiLineValueFromString("font-size:1em", text.IdentityDecoder))
}

func TestTableWithAlignStyle(t *testing.T) {
	markdown := testutil.NewMarkdownToStringFunc(
		parser.New(
			parser.WithExtensions(NewTableParser()),
		),
		html.New(
			html.WithXHTML(),
			html.WithUnsafe(),
			html.WithExtensions(NewTableHTMLRenderer(
				WithTableCellAlignMethod(TableCellAlignStyle),
			)),
		),
	)
	testutil.DoTestCase(
		markdown,
		testutil.MarkdownTestCase{
			No:          1,
			Description: "Cell with TableCellAlignStyle and XHTML should be rendered as a style attribute",
			Markdown: `
| abc | defghi |
:-: | -----------:
bar | baz
`,
			Expected: `<table>
<thead>
<tr>
<th style="text-align:center">abc</th>
<th style="text-align:right">defghi</th>
</tr>
</thead>
<tbody>
<tr>
<td style="text-align:center">bar</td>
<td style="text-align:right">baz</td>
</tr>
</tbody>
</table>`,
		},
		t,
	)

	markdown = testutil.NewMarkdownToStringFunc(
		parser.New(
			parser.WithExtensions(NewTableParser()),
		),
		html.New(
			html.WithUnsafe(),
			html.WithExtensions(NewTableHTMLRenderer(
				WithTableCellAlignMethod(TableCellAlignStyle),
			)),
		),
	)
	testutil.DoTestCase(
		markdown,
		testutil.MarkdownTestCase{
			No:          2,
			Description: "Cell with TableCellAlignStyle and HTML5 should be rendered as a style attribute",
			Markdown: `
| abc | defghi |
:-: | -----------:
bar | baz
`,
			Expected: `<table>
<thead>
<tr>
<th style="text-align:center">abc</th>
<th style="text-align:right">defghi</th>
</tr>
</thead>
<tbody>
<tr>
<td style="text-align:center">bar</td>
<td style="text-align:right">baz</td>
</tr>
</tbody>
</table>`,
		},
		t,
	)

	markdown = testutil.NewMarkdownToStringFunc(
		parser.New(
			parser.WithASTTransformers(
				util.Prioritized[parser.ASTTransformer](&tableStyleTransformer{}, 0),
			),
			parser.WithExtensions(NewTableParser()),
		),
		html.New(
			html.WithUnsafe(),
			html.WithExtensions(NewTableHTMLRenderer(
				WithTableCellAlignMethod(TableCellAlignStyle),
			)),
		),
	)

	testutil.DoTestCase(
		markdown,
		testutil.MarkdownTestCase{
			No:          3,
			Description: "Styled cell should not be broken the style by the alignments",
			Markdown: `
| abc | defghi |
:-: | -----------:
bar | baz
`,
			Expected: `<table>
<thead>
<tr>
<th style="font-size:1em;text-align:center">abc</th>
<th style="text-align:right">defghi</th>
</tr>
</thead>
<tbody>
<tr>
<td style="text-align:center">bar</td>
<td style="text-align:right">baz</td>
</tr>
</tbody>
</table>`,
		},
		t,
	)
}

func TestTableWithAlignNone(t *testing.T) {
	markdown := testutil.NewMarkdownToStringFunc(
		parser.New(
			parser.WithExtensions(NewTableParser()),
		),
		html.New(
			html.WithXHTML(),
			html.WithUnsafe(),
			html.WithExtensions(NewTableHTMLRenderer(
				WithTableCellAlignMethod(TableCellAlignNone),
			)),
		),
	)
	testutil.DoTestCase(
		markdown,
		testutil.MarkdownTestCase{
			No:          1,
			Description: "Cell with TableCellAlignStyle and XHTML should not be rendered",
			Markdown: `
| abc | defghi |
:-: | -----------:
bar | baz
`,
			Expected: `<table>
<thead>
<tr>
<th>abc</th>
<th>defghi</th>
</tr>
</thead>
<tbody>
<tr>
<td>bar</td>
<td>baz</td>
</tr>
</tbody>
</table>`,
		},
		t,
	)
}

func TestTableFuzzedPanics(t *testing.T) {
	markdown := testutil.NewMarkdownToStringFunc(
		parser.New(
			parser.WithExtensions(NewTableParser()),
		),
		html.New(
			html.WithXHTML(),
			html.WithUnsafe(),
			html.WithExtensions(NewTableHTMLRenderer()),
		),
	)
	testutil.DoTestCase(
		markdown,
		testutil.MarkdownTestCase{
			No:          1,
			Description: "This should not panic",
			Markdown:    "* 0\n-|\n\t0",
			Expected: `<ul>
<li>
<table>
<thead>
<tr>
<th>0</th>
</tr>
</thead>
<tbody>
<tr>
<td>0</td>
</tr>
</tbody>
</table>
</li>
</ul>`,
		},
		t,
	)
}
