# Technical Documentation: `extension/table.go`

## Overview

The `extension/table.go` file provides support for parsing and rendering GitHub Flavored Markdown (GFM) tables within the [Goldmark](github.com/yuin/goldmark) Markdown engine framework. It implements paragraph and AST transformation logic to convert raw table text into structured table Abstract Syntax Tree (AST) nodes, alongside an HTML renderer to convert table AST nodes into standard HTML `<table>` elements with configurable alignment formatting.

---

## Key Components

### 1. Types and Enums

#### `TableCellAlignMethod`
Defines how table cell text alignments are formatted when rendered into HTML.

```go
type TableCellAlignMethod int
```

* **`TableCellAlignDefault`** (0): Renders alignments based on HTML output mode. For XHTML, it uses the `align` attribute. For HTML5, it uses the CSS `style` attribute (`text-align:...`).
* **`TableCellAlignAttribute`** (1): Renders alignments using the HTML `align` attribute (e.g., `align="left"`).
* **`TableCellAlignStyle`** (2): Renders alignments using the inline CSS `style` attribute (e.g., `style="text-align:left"`).
* **`TableCellAlignNone`** (3): Omits explicit alignment attributes. Useful when alignments are handled via external classes or AST transformers.

#### Configuration & Functional Options

* **`tableHTMLRendererConfig`**: Holds renderer settings (`XHTML` bool, `TableCellAlignMethod`).
* **`TableHTMLRendererOption`**: Functional option interface for configuring the table HTML renderer.
  * **`WithTableCellAlignMethod(a TableCellAlignMethod)`**: Returns a `TableHTMLRendererOption` that configures how cell alignment is rendered in HTML.

---

### 2. Context Tracking & Escaped Pipes

When code spans inside table cells contain escaped pipe characters (`\|`), special handling is required so that the pipe is not parsed as a column delimiter.

* **`escapedPipeCellListKey`**: A context key (`parser.NewContextKey()`) stored in `parser.Context` to retain tracked escaped cells during parsing.
* **`escapedPipeCell`**: Struct tracking table cells that contain escaped pipes within backticks:
  ```go
  type escapedPipeCell struct {
      Cell        *ast.TableCell
      Pos         []int
      Transformed bool
  }
  ```

---

### 3. Delimiter Recognition & Parsing Logic

Table header delimiter rows are evaluated using utility functions and Regular Expressions:

* **`isTableDelim(bs []byte) bool`**: Checks whether a line qualifies as a delimiter row. Requires indentation to be 3 spaces or fewer, excludes lines composed solely of `-` characters, and verifies that content only contains space, `-`, `|`, and `:` characters.
* **Delimiter Regular Expressions**:
  * `tableDelimLeft`: Matches left-aligned columns (`:-+`).
  * `tableDelimRight`: Matches right-aligned columns (`-+:`).
  * `tableDelimCenter`: Matches center-aligned columns (`:-+:`).
  * `tableDelimNone`: Matches unaligned columns (`-+`).

---

### 4. Parser Extensions & Transformers

#### `tableParagraphTransformer`
Implements `parser.ParagraphTransformer`. It intercepts paragraph nodes in the parsing pipeline and detects whether lines conform to a Markdown table structure.

* **`Transform(node *gast.Paragraph, reader text.Reader, pc parser.Context)`**:
  * Evaluates paragraphs containing two or more lines.
  * Checks for a valid header delimiter line (`parseDelimiter`).
  * Parses the preceding line as the header row using `parseRow`.
  * Constructs an `ast.Table` node, attaching an `ast.TableHeader` and an `ast.TableBody`.
  * Appends parsed body rows (`ast.TableRow`) for subsequent lines.
  * Trims or removes processed table lines from the original paragraph node source.
* **`parseRow(...) *ast.TableRow`**:
  * Trims pipe characters at the outer boundaries of the row.
  * Iterates through segments bounded by `|`.
  * Detects code backticks (`` ` ``) and tracks escaped pipe characters (`\|`), populating `escapedPipeCell` structures into `parser.Context`.
  * Returns an `ast.TableRow` containing populated `ast.TableCell` nodes.
* **`parseDelimiter(...) []ast.Alignment`**:
  * Splits the delimiter row by `|` and evaluates each column segment against the delimiter regular expressions.
  * Returns a slice of column alignment enumerations (`ast.AlignLeft`, `ast.AlignRight`, `ast.AlignCenter`, `ast.AlignNone`).

#### `tableASTTransformer`
Implements `parser.ASTTransformer`.

* **`Transform(_ *gast.Document, reader text.Reader, pc parser.Context)`**:
  * Retrieves tracked `escapedPipeCell` entries stored in the context under `escapedPipeCellListKey`.
  * Walks code span nodes (`gast.KindCodeSpan`) within tracked cells and restores `\|` replacements back to literal `|` characters in the AST.

#### `tableParserExtension`
Implements `parser.Extension` to register the parser options.

* **`NewTableParser() parser.Extension`**: Constructs a table parser extension.
* **`ParserOptions(_ *parser.Config) []parser.Option`**: Registers:
  * `tableParagraphTransformer` with priority `200`.
  * `defaultTableASTTransformer` with priority `0`.

---

### 5. HTML Renderer Extension

#### `tableHTMLRendererExtension`
Implements `html.Extension` to convert table AST nodes to HTML output.

* **`NewTableHTMLRenderer(opts ...TableHTMLRendererOption)`**: Creates a table renderer extension configured with optional settings.
* **`RendererOptions(c *html.Config) []html.Option`**: Configures renderer function mappings for table AST nodes:
  * `ast.KindTable` $\rightarrow$ `renderTable`
  * `ast.KindTableHeader` $\rightarrow$ `renderTableHeader`
  * `ast.KindTableBody` $\rightarrow$ `renderTableBody`
  * `ast.KindTableRow` $\rightarrow$ `renderTableRow`
  * `ast.KindTableCell` $\rightarrow$ `renderTableCell`

#### Node Render Functions

1. **`renderTable`**: Outputs `<table>` and `</table>\n`. Applies `TableAttributeFilter`.
2. **`renderTableHeader`**: Outputs `<thead>\n<tr>\n` when entering and `</tr>\n</thead>\n` when leaving. Applies `TableHeaderAttributeFilter`.
3. **`renderTableBody`**: Outputs `<tbody>\n` and `</tbody>\n`.
4. **`renderTableRow`**: Outputs `<tr>` and `</tr>\n`. Applies `TableRowAttributeFilter`.
5. **`renderTableCell`**:
   * Renders `<th>` elements when the parent node is an `ast.TableHeader`, otherwise renders `<td>`.
   * Evaluates alignment settings:
     * Handles standard HTML attribute (`align="left|right|center"`) or CSS style (`style="text-align:left|right|center"`).
     * Automatically adjusts behavior based on the `XHTML` configuration flag if `TableCellAlignDefault` is selected.
   * Applies `TableThCellAttributeFilter` for header cells or `TableTdCellAttributeFilter` for body cells.

---

### 6. Attribute Filters

The file defines several HTML attribute filters using `html.GlobalAttributeFilter.ExtendString(...)` to filter attributes rendered onto table HTML elements:

| Filter Variable | Allowed Additional Attributes |
| :--- | :--- |
| **`TableAttributeFilter`** | `align`, `bgcolor`, `border`, `cellpadding`, `cellspacing`, `frame`, `rules`, `summary`, `width` |
| **`TableHeaderAttributeFilter`** | `align`, `bgcolor`, `char`, `charoff`, `valign` |
| **`TableBodyAttributeFilter`** | `align`, `bgcolor`, `char`, `charoff`, `valign` |
| **`TableRowAttributeFilter`** | `align`, `bgcolor`, `char`, `charoff`, `valign` |
| **`TableThCellAttributeFilter`** | `abbr`, `align`, `axis`, `bgcolor`, `char`, `charoff`, `colspan`, `headers`, `height`, `rowspan`, `scope`, `valign`, `width` |
| **`TableTdCellAttributeFilter`** | `abbr`, `align`, `axis`, `bgcolor`, `char`, `charoff`, `colspan`, `headers`, `height`, `rowspan`, `scope`, `valign`, `width` |

---

## Exported Package Variables

For convenience, ready-to-use instances of default parser and renderer extensions are exposed:

* **`TableParser`**: Pre-configured default parser extension instance (`NewTableParser()`).
* **`TableHTMLRenderer`**: Pre-configured default HTML renderer extension instance (`NewTableHTMLRenderer()`).