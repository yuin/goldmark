# Technical Documentation: `parser/code_block.go`

## Overview

The `parser/code_block.go` file implements the parsing logic for **indented code blocks** in Markdown according to the Goldmark AST/parsing architecture. 

It defines `codeBlockParser`, which implements the `BlockParser` interface. This parser identifies lines indented by at least 4 spaces (or equivalent tab stops), collects subsequent lines belonging to the code block, handles leading tab preservation, and strips trailing blank lines when the block is closed.

---

## Constants and Package Variables

* **`defaultCodeBlockParser`**: A package-level singleton instance of `codeBlockParser`.

---

## Types & Constructors

### `codeBlockParser`
```go
type codeBlockParser struct {}
```
An unexported struct that implements the `BlockParser` interface for processing indented code blocks.

### `NewCodeBlockParser`
```go
func NewCodeBlockParser() BlockParser
```
Returns the singleton `defaultCodeBlockParser` instance as a `BlockParser` interface.

---

## Method Implementations

### `Trigger`
```go
func (b *codeBlockParser) Trigger() []byte
```
* **Returns**: `nil`
* **Purpose**: Indicates that this parser does not trigger on specific bytes/characters. Instead, line indentation determines whether this parser opens.

---

### `Open`
```go
func (b *codeBlockParser) Open(_ ast.Node, reader text.Reader, _ Context) (ast.Node, State)
```
* **Purpose**: Evaluates whether a new indented code block starts at the current reader position.
* **Logic**:
  1. Peeks at the current line using `reader.PeekLine()`.
  2. Calculates the line's indent position using `util.IndentPosition(line, reader.LineOffset(), 4)`.
  3. Returns `nil, NoChildren` if the line has less than 4 spaces of indentation (`pos < 0`) or if the line is blank (`util.IsBlank(line)`).
  4. Creates an AST node of type `ast.CodeBlock` with kind `ast.CodeBlockKindIndented`.
  5. Advances the reader offset and updates padding with `reader.AdvanceAndSetPadding(pos, padding)`.
  6. Re-peeks the line to retrieve the updated text segment.
  7. If `segment.Padding != 0`, calls `preserveLeadingTabInCodeBlock(&segment, reader, 0)` to handle leading tabs properly.
  8. Sets `segment.ForceNewline = true`.
  9. Appends the segment to `node.Value`.
  10. Advances the reader to the end of the line (`reader.AdvanceToEOL()`).
  11. Returns `node, NoChildren`.

---

### `Continue`
```go
func (b *codeBlockParser) Continue(node ast.Node, reader text.Reader, _ Context) State
```
* **Purpose**: Determines whether the current line continues the existing indented code block.
* **Logic**:
  1. Casts `node` to `*ast.CodeBlock`.
  2. Peeks at the current line.
  3. **Blank Lines**: If the line is blank (`util.IsBlank(line)`), it trims up to 4 spaces of left margin using `segment.TrimLeftSpaceWidth(4, reader.Source())`, appends the segment to `cb.Value`, and returns `Continue | NoChildren`.
  4. **Non-Blank Lines**: Calculates the indentation position using `util.IndentPosition(line, reader.LineOffset(), 4)`.
  5. If `pos < 0` (fewer than 4 spaces of indentation), returns `Close` to close the code block.
  6. Advances reader padding (`reader.AdvanceAndSetPadding(pos, padding)`), re-peeks the line segment, preserves leading tabs if `segment.Padding != 0`, sets `ForceNewline = true`, appends the segment to `cb.Value`, and advances to EOL.
  7. Returns `Continue | NoChildren`.

---

### `Close`
```go
func (b *codeBlockParser) Close(node ast.Node, reader text.Reader, _ Context)
```
* **Purpose**: Finalizes the code block node when parsing of the block completes.
* **Logic**:
  1. Casts `node` to `*ast.CodeBlock`.
  2. Iterates backward through `cb.Value.Segments()` starting from the end to locate trailing blank lines using `util.IsBlank(...)`.
  3. Determines the index `length` of the last non-blank segment line.
  4. Rebuilds `cb.Value` keeping only the segments from index `0` through `length` (discarding trailing blank lines).

---

### `CanInterruptParagraph`
```go
func (b *codeBlockParser) CanInterruptParagraph() bool
```
* **Returns**: `false`
* **Purpose**: Specifies that an indented code block cannot interrupt an active paragraph.

---

### `CanAcceptIndentedLine`
```go
func (b *codeBlockParser) CanAcceptIndentedLine() bool
```
* **Returns**: `true`
* **Purpose**: Indicates that this parser accepts indented lines.

---

## Helper Functions

### `preserveLeadingTabInCodeBlock`
```go
func preserveLeadingTabInCodeBlock(segment *text.Segment, reader text.Reader, indent int)
```
* **Parameters**:
  * `segment`: Pointer to the current `text.Segment`.
  * `reader`: The `text.Reader` instance.
  * `indent`: An integer offset modifier.
* **Purpose**: Adjusts the segment properties when a line starts with a tab character to ensure leading tabs within code blocks are accurately preserved rather than completely converted into virtual padding space.
* **Logic**:
  1. Stores the target offset with padding (`reader.LineOffset() + indent`).
  2. Adjusts reader position back 1 character byte.
  3. If the adjusted offset matches the line offset, sets `segment.Padding = 0` and decrements `segment.Start`.
  4. Restores the reader position back to its original state.