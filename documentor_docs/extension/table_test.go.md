# Documentation Guide: `extension/table_test.go`

## Overview

The `extension/table_test.go` file contains unit tests for Goldmark's markdown table extension. Its primary purpose is to verify the parsing, rendering, alignment options, and robustness of markdown table parsing and HTML generation. 

It tests:
- Table parsing and rendering from external test files.
- Table cell text alignment strategies (`TableCellAlignDefault`, `TableCellAlignAttribute`, `TableCellAlignStyle`, `TableCellAlignNone`).
- Standard HTML5 vs. XHTML output formatting for table alignments.
- Merging existing AST attributes (e.g., custom `style` attributes) with table alignment styles.
- Edge cases discovered through fuzz testing to prevent parser panics.

---

## Package and Dependencies

**Package Name:** `extension`

### Imported Packages
- `testing`: Standard Go testing library.
- `github.com/yuin/goldmark/v2/ast`: Provides base AST node definitions.
- `east` (`github.com/yuin/goldmark/v2/extension/ast`): Provides table-specific AST extensions (such as `east.TableCell`).
- `github.com/yuin/goldmark/v2/parser`: Interfaces and options for markdown parsing.
- `github.com/yuin/goldmark/v2/renderer/html`: Options for HTML rendering (e.g., `WithUnsafe`, `WithXHTML`).
- `github.com/yuin/goldmark/v2/testutil`: Utility functions for running markdown-to-string test cases.
- `github.com/yuin/goldmark/v2/text`: Text manipulation abstractions (e.g., `text.Reader`, `text.NewMultiLineValueFromString`).
- `github.com/yuin/goldmark/v2/util`: Utility wrappers (e.g., `util.Prioritized`).

---

## Helper Types and Structures

### `tableStyleTransformer`

`tableStyleTransformer` is a custom, private AST transformer struct used exclusively within `TestTableWithAlignStyle` to test how alignment styles interact with pre-existing element attributes.

```go
type tableStyleTransformer struct {
}
```

#### Method: `Transform`
```go
func (a *tableStyleTransformer) Transform(node *ast.Document, _ text.Reader, _ parser.Context)
```
- **Description**: Navigates the AST to the first cell of the table (`Document` -> `Table` -> `TableRow` -> `TableCell`) and injects a custom `style="font-size:1em"` attribute onto that node before rendering.
- **Node Traversal**:
  1. `node.FirstChild()`: Navigates to the first block element (`Table`).
  2. `.FirstChild()`: Navigates to the first row (`TableHeader` or `TableRow`).
  3. `.FirstChild()`: Navigates to the first cell (`TableCell`).
  4. Type-asserts to `*east.TableCell`.
  5. Injects the attribute `"style"` with the value `"font-size:1em"`.

---

## Test Functions

### 1. `TestTable(t *testing.T)`
- **Purpose**: Verifies core table parsing and rendering against an external test file (`_test/table.txt`).
- **Configuration**:
  - **Parser**: Uses `NewTableParser()`.
  - **HTML Renderer**: Configured with `html.WithUnsafe()`, `html.WithXHTML()`, and `NewTableHTMLRenderer()`.
- **Execution**: Calls `testutil.DoTestCaseFile` to execute test cases defined in `_test/table.txt`.

---

### 2. `TestTableWithAlignDefault(t *testing.T)`
- **Purpose**: Tests table alignment output when `WithTableCellAlignMethod(TableCellAlignDefault)` is specified.
- **Test Cases**:
  1. **XHTML Mode**:
     - **Config**: Rendered with `html.WithXHTML()`.
     - **Expected Outcome**: Cell alignments are rendered using the legacy `align` attribute (e.g., `<th align="center">`, `<td align="right">`).
  2. **HTML5 Mode**:
     - **Config**: Rendered without `html.WithXHTML()`.
     - **Expected Outcome**: Cell alignments are rendered using inline CSS `style` attributes (e.g., `<th style="text-align:center">`, `<td style="text-align:right">`).

---

### 3. `TestTableWithAlignAttribute(t *testing.T)`
- **Purpose**: Tests table alignment output when `WithTableCellAlignMethod(TableCellAlignAttribute)` is explicitly specified.
- **Test Cases**:
  1. **XHTML Mode**: Rendered with `html.WithXHTML()`. Alignment is formatted as `align="..."`.
  2. **HTML5 Mode**: Rendered without `html.WithXHTML()`. Alignment is still explicitly formatted as `align="..."`.

---

### 4. `TestTableWithAlignStyle(t *testing.T)`
- **Purpose**: Tests table alignment output when `WithTableCellAlignMethod(TableCellAlignStyle)` is specified, including attribute modification interactions.
- **Test Cases**:
  1. **XHTML Mode**: Rendered with `html.WithXHTML()`. Alignment is formatted as CSS `style="text-align:..."`.
  2. **HTML5 Mode**: Rendered without `html.WithXHTML()`. Alignment is formatted as CSS `style="text-align:..."`.
  3. **Attribute Merging Case**:
     - Uses `tableStyleTransformer` registered via `parser.WithASTTransformers` (priority 0).
     - **Scenario**: Pre-sets `font-size:1em` on the first `TableCell`.
     - **Expected Outcome**: Alignment logic appends or merges `text-align:center` into the existing `style` attribute without destroying it, producing `<th style="font-size:1em;text-align:center">`.

---

### 5. `TestTableWithAlignNone(t *testing.T)`
- **Purpose**: Verifies behavior when `WithTableCellAlignMethod(TableCellAlignNone)` is configured.
- **Test Cases**:
  1. **Case 1**: Markdown markdown aligned headers/cells (`:-:`, `-----------:`) are parsed, but the alignment information is completely suppressed during HTML rendering.
     - **Expected Outcome**: Plain `<th>` and `<td>` tags with no `align` or `style` attributes.

---

### 6. `TestTableFuzzedPanics(t *testing.T)`
- **Purpose**: Ensures the table parser gracefully handles malformed or edge-case markdown without causing runtime panics.
- **Input Tested**:
  ```markdown
  * 0
  -|
  	0
  ```
- **Expected Outcome**: Successfully renders nested list and table structure into HTML without throwing a panic:
  ```html
  <ul>
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
  </ul>
  ```

---

## Cell Alignment Method Summary

| Alignment Option (`TableCellAlignMethod`) | XHTML Enabled (`html.WithXHTML()`) | HTML Output Format |
| :--- | :--- | :--- |
| `TableCellAlignDefault` | Yes | `align="<align>"` |
| `TableCellAlignDefault` | No | `style="text-align:<align>"` |
| `TableCellAlignAttribute` | Yes / No | `align="<align>"` |
| `TableCellAlignStyle` | Yes / No | `style="text-align:<align>"` |
| `TableCellAlignNone` | Yes / No | *No alignment attributes* |