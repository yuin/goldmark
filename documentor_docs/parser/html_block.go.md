# Technical Documentation: `parser/html_block.go`

## Overview

The `parser/html_block.go` file implements the HTML block parser for the Goldmark Markdown parsing engine (v2). It defines the `htmlBlockParser` struct, which implements the `BlockParser` interface. This parser detects, parses, and manages raw HTML blocks within Markdown documents according to the HTML block rules specified by CommonMark (Types 1 through 7).

---

## Global Variables and Data Structures

### `allowedBlockTags`
```go
var allowedBlockTags = map[string]bool{ ... }
```
A map containing set-like `true` lookups for lowercased HTML block-level tag names (e.g., `article`, `div`, `table`, `p`, `h1`-`h6`, etc.). This map is used to determine whether a recognized HTML tag qualifies for CommonMark **Type 6** HTML blocks.

### Regular Expressions and Closure Sequences

The file defines regular expressions and byte sequences to match opening and closing conditions for the standard CommonMark HTML block kinds:

| HTML Block Kind | Type | Opening Pattern (`Regexp`) | Closure Condition (`[]byte` / `Regexp`) |
| :--- | :--- | :--- | :--- |
| **Type 1** | `<script>`, `<pre>`, `<style>`, `<textarea>` | `htmlBlockType1OpenRegexp` | `htmlBlockType1CloseRegexp` |
| **Type 2** | HTML Comment (`<!--`) | `htmlBlockType2OpenRegexp` | `htmlBlockType2Close` (`-->`) |
| **Type 3** | Processing Instruction (`<?`) | `htmlBlockType3OpenRegexp` | `htmlBlockType3Close` (`?>`) |
| **Type 4** | Declaration (`<![A-Z]+...`) | `htmlBlockType4OpenRegexp` | `htmlBlockType4Close` (`>`) |
| **Type 5** | CDATA Section (`<![CDATA[`) | `htmlBlockType5OpenRegexp` | `htmlBlockType5Close` (`]]>`) |
| **Type 6** | Standard HTML block tags | `htmlBlockType6Regexp` | Blank line (`util.IsBlank`) |
| **Type 7** | Custom / General HTML tags | `htmlBlockType7Regexp` | Blank line (`util.IsBlank`) |

---

## Implementation Details: `htmlBlockParser`

`htmlBlockParser` is a stateless struct that handles parsing HTML blocks.

```go
type htmlBlockParser struct{}
```

### Constructor

#### `NewHTMLBlockParser`
```go
func NewHTMLBlockParser() BlockParser
```
Returns a `BlockParser` instance configured for parsing raw HTML blocks using the package-level singleton `defaultHTMLBlockParser`.

---

### Interface Methods

#### `Trigger`
```go
func (b *htmlBlockParser) Trigger() []byte
```
* **Returns:** `[]byte{'<'}`
* **Description:** Specifies that the HTML block parser is triggered when a line starts with the `<` character (after up to 3 spaces of indentation).

#### `Open`
```go
func (b *htmlBlockParser) Open(parent ast.Node, reader text.Reader, pc Context) (ast.Node, State)
```
* **Parameters:**
  * `parent`: The parent AST node (unused here).
  * `reader`: The reader providing access to the current line and text segments.
  * `pc`: The parsing context, used to inspect previously opened blocks.
* **Returns:** `(ast.Node, State)` — An `ast.HTMLBlock` node and `NoChildren` state if an HTML block starts; otherwise `(nil, NoChildren)`.
* **Behavior:**
  1. Peeks at the current line from the reader.
  2. Sequentially tests the line against opening patterns for Types 1 through 5.
  3. Tests against `htmlBlockType7Regexp`. If matched:
     * Extracts the tag name, close tag status, and presence of attributes.
     * If the tag name is in `allowedBlockTags`, it creates a Type 6 HTML block (`ast.HTMLBlockKind6`).
     * Otherwise, if the tag is not `script`, `style`, or `pre`, does not attempt to interrupt a paragraph invalidly, and satisfies tag structure, it creates a Type 7 HTML block (`ast.HTMLBlockKind7`).
  4. If no node has been created yet, tests against `htmlBlockType6Regexp` to check if the tag belongs to `allowedBlockTags`.
  5. If an HTML block matches:
     * Advances the reader to the end of the current line (`reader.AdvanceToEOL()`).
     * Appends the line segment to the `HTMLBlock` node's `Value`.
     * Returns the created `ast.HTMLBlock` node and `NoChildren`.

#### `Continue`
```go
func (b *htmlBlockParser) Continue(node ast.Node, reader text.Reader, pc Context) State
```
* **Parameters:**
  * `node`: The current `ast.Node`, cast to `*ast.HTMLBlock`.
  * `reader`: Text reader positioned at the current line.
  * `pc`: Parsing context.
* **Returns:** `State` (`Close` or `Continue | NoChildren`).
* **Behavior:**
  * Evaluates termination conditions based on the block's `HTMLBlockKind`:
    * **Type 1:** Checks if `htmlBlockType1CloseRegexp` matches the current line (or single first line). Returns `Close` if matched.
    * **Types 2–5:** Selects the corresponding closure pattern (`-->`, `?>`, `>`, or `]]>`). Returns `Close` if the line contains the closure pattern.
    * **Types 6–7:** Checks if the line is blank (`util.IsBlank(line)`). Returns `Close` if a blank line is encountered.
  * If the block does not close on the line:
    * Appends the current segment to `htmlBlock.Value`.
    * Advances the reader to the end of line (`reader.AdvanceToEOL()`).
    * Returns `Continue | NoChildren`.

#### `Close`
```go
func (b *htmlBlockParser) Close(node ast.Node, reader text.Reader, pc Context)
```
* **Description:** No-op method. No state cleanup or post-processing is needed when an HTML block closes.

#### `CanInterruptParagraph`
```go
func (b *htmlBlockParser) CanInterruptParagraph() bool
```
* **Returns:** `true`
* **Description:** Indicates that HTML blocks (subject to Type 7 inner restrictions handled in `Open`) can interrupt an ongoing paragraph block.

#### `CanAcceptIndentedLine`
```go
func (b *htmlBlockParser) CanAcceptIndentedLine() bool
```
* **Returns:** `false`
* **Description:** HTML blocks cannot be initiated with 4 or more spaces of indentation (indented code block precedence).