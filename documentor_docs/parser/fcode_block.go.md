# Technical Documentation: `parser/fcode_block.go`

## Overview

The `parser/fcode_block.go` file implements a fenced code block parser for the Goldmark Markdown engine (v2). It provides the mechanisms to detect, open, continue, and close fenced code blocks (lines surrounded by opening and closing sequences of backticks `` ` `` or tildes `~`) according to Markdown parsing rules.

---

## Constants, Variables, and Context Keys

### `fencedCodeBlockInfoKey`
```go
var fencedCodeBlockInfoKey = NewContextKey()
```
A context key used to store and retrieve active `fenceData` state within the parser context (`pc Context`).

### `defaultFencedCodeBlockParser`
```go
var defaultFencedCodeBlockParser = &fencedCodeBlockParser{}
```
A singleton instance of `fencedCodeBlockParser` returned by the public constructor.

---

## Data Structures

### `fencedCodeBlockParser`
```go
type fencedCodeBlockParser struct {}
```
An empty struct implementing the `BlockParser` interface. It contains the logic for parsing fenced code blocks.

### `fenceData`
```go
type fenceData struct {
    char   byte
    indent int
    length int
    node   ast.Node
}
```
A internal state structure maintained in the parser context while a fenced code block is open.

- **`char`**: The fence character used (either `` '`' `` or `'~'`).
- **`indent`**: The indentation level (offset) of the opening fence line.
- **`length`**: The number of repeating fence characters in the opening fence.
- **`node`**: The AST node (`*ast.CodeBlock`) representing the fenced code block being constructed.

---

## Constructor

### `NewFencedCodeBlockParser`
```go
func NewFencedCodeBlockParser() BlockParser
```
Returns a `BlockParser` interface wrapping the singleton `defaultFencedCodeBlockParser`.

---

## Methods and Lifecycle Implementations

The `fencedCodeBlockParser` type implements the `BlockParser` interface methods as described below:

### 1. `Trigger`
```go
func (b *fencedCodeBlockParser) Trigger() []byte
```
- **Returns**: `[]byte{'~', '`'}`
- **Purpose**: Informs the parser engine to invoke this block parser whenever a line starts with either a tilde (`~`) or a backtick (`` ` ``).

---

### 2. `Open`
```go
func (b *fencedCodeBlockParser) Open(_ ast.Node, reader text.Reader, pc Context) (ast.Node, State)
```
Invoked when a line starts with one of the trigger characters to evaluate if a new fenced code block begins.

#### Process Flow:
1. **Character Validation**:
   - Gets the current line and block offset `pos` from context `pc.BlockOffset()`.
   - Validates that `pos >= 0` and `line[pos]` is either `` '`' `` or `'~'`. If not, returns `nil, NoChildren`.
2. **Fence Length Check**:
   - Counts consecutive occurrences of the fence character starting at `pos`.
   - If the opening fence length (`oFenceLength`) is less than `3`, returns `nil, NoChildren`.
3. **Info String Processing**:
   - Extracts characters following the fence sequence (`rest := line[i:]`).
   - Trims leading and trailing spaces.
   - **Backtick Rule**: If the fence character is a backtick (`` '`' ``) and the info string contains a backtick character, the line is invalid as a fenced code block header, returning `nil, NoChildren`.
   - If a valid info string exists, creates a `text.SingleLineValue` using the slice indices and the reader's decoder.
4. **AST Node Creation & State Persistence**:
   - Creates an `ast.CodeBlock` node configured as `ast.CodeBlockKindFenced` with the info string attached.
   - Saves a new `fenceData` struct into the parser context using `fencedCodeBlockInfoKey`.
   - Returns the created `ast.Node` and state `NoChildren`.

---

### 3. `Continue`
```go
func (b *fencedCodeBlockParser) Continue(node ast.Node, reader text.Reader, pc Context) State
```
Called on subsequent lines to process either the closing fence or content inside the code block.

#### Process Flow:
1. **Closing Fence Check**:
   - Retrieves `fenceData` from context `pc`.
   - Checks the line's indentation width (`util.IndentWidth`).
   - If indent width is **less than 4 spaces** (`w < 4`):
     - Counts matching fence characters (`fdata.char`).
     - If the closing fence character count is **greater than or equal to** opening fence length (`length >= fdata.length`) and the remainder of the line is blank (`util.IsBlank(line[i:])`):
       - Advances reader to end-of-line (`reader.AdvanceToEOL()`).
       - Returns `Close` state to close the block.
2. **Content Processing**:
   - Computes relative line position and padding based on the opening fence's indentation.
   - If `pos < 0`, adjusts `pos` using `util.FirstNonSpacePosition(line)` minus segment padding, and sets padding to `0`.
   - Constructs a new `text.SegmentPadding` representing the line content.
   - If `padding != 0`, calls `preserveLeadingTabInCodeBlock(&seg, reader, fdata.indent)` to preserve tabs.
   - Sets `ForceNewline = true` on the segment to handle EOF as a newline.
   - Appends the segment to the `ast.CodeBlock` node's `Value`.
   - Advances reader to end-of-line (`reader.AdvanceToEOL()`).
   - Returns state `Continue | NoChildren`.

---

### 4. `Close`
```go
func (b *fencedCodeBlockParser) Close(node ast.Node, _ text.Reader, pc Context)
```
- **Purpose**: Cleans up parser context when the code block node is closed.
- **Behavior**: Retrieves `fenceData` from `pc`. If `fdata.node == node`, sets `fencedCodeBlockInfoKey` in `pc` to `nil`.

---

### 5. `CanInterruptParagraph`
```go
func (b *fencedCodeBlockParser) CanInterruptParagraph() bool
```
- **Returns**: `true`
- **Purpose**: Indicates that a fenced code block can interrupt an open paragraph block.

---

### 6. `CanAcceptIndentedLine`
```go
func (b *fencedCodeBlockParser) CanAcceptIndentedLine() bool
```
- **Returns**: `false`
- **Purpose**: Indicates that standard 4-space indented lines without fences cannot trigger this parser.