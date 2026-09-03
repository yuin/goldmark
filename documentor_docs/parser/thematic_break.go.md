# Technical Documentation: `parser/thematic_break.go`

## Overview

The `parser/thematic_break.go` file implements the block parsing logic for thematic breaks (commonly referred to as horizontal rules) in Markdown. It provides a parser that satisfies the `BlockParser` interface from the Goldmark library, identifying valid sequence patterns of `-`, `*`, or `_` characters and constructing corresponding `ast.ThematicBreak` nodes.

---

## Constants and Variables

### `defaultThematicBreakParser`
```go
var defaultThematicBreakParser = &thematicBreakParser{}
```
A package-level singleton instance of `thematicBreakParser`. Because the parser holds no internal state, a single instance is reused across parser invocations.

---

## Data Structures

### `thematicBreakParser`
```go
type thematicBreakParser struct {
}
```
An unexported struct that implements the `BlockParser` interface for processing thematic breaks.

---

## Public Functions

### `NewThematicBreakParser`
```go
func NewThematicBreakParser() BlockParser
```
Returns a `BlockParser` interface backed by the package-level `defaultThematicBreakParser` instance.

* **Returns:** `BlockParser` — The thematic break parser instance.

---

## Helper Functions

### `isThematicBreak`
```go
func isThematicBreak(line []byte, offset int) bool
```
Determines whether a given byte slice (`line`) starting at a specific `offset` forms a valid thematic break.

#### Validation Rules & Algorithm:
1. **Indentation Check:** Uses `util.IndentWidth(line, offset)` to compute the line's indentation width `w` and starting position `pos`. If `w > 3` (more than 3 spaces of indentation), the function immediately returns `false`.
2. **Character Scanning:** Scans the slice from `pos` to the end of `line`:
   - Ignores whitespace characters (checked via `util.IsSpace(c)`).
   - Establishes the target delimiter character (`mark`) from the first non-space byte encountered.
   - Requires `mark` to be one of: `'*'`, `'-'`, or `'_'`. If it is any other character, the function returns `false`.
   - Ensures all subsequent non-space characters in the line match `mark`. If a mismatch is found, the function returns `false`.
   - Increments a `count` for each matching delimiter character found.
3. **Count Verification:** Returns `true` if `count > 2` (at least 3 matching characters), otherwise `false`.

---

## `BlockParser` Interface Implementation

`thematicBreakParser` implements the following methods required by the Goldmark block parser framework:

### `Trigger`
```go
func (b *thematicBreakParser) Trigger() []byte
```
* **Returns:** `[]byte{'-', '*', '_'}`
* **Purpose:** Defines the trigger bytes that cause the parser engine to evaluate this parser when scanning lines.

### `Open`
```go
func (b *thematicBreakParser) Open(_ ast.Node, reader text.Reader, _ Context) (ast.Node, State)
```
* **Parameters:**
  * `_ ast.Node`: Unused parent node.
  * `reader text.Reader`: Reader providing access to the document text.
  * `_ Context`: Unused parser context.
* **Behavior:**
  1. Peeks at the current line from the `reader`.
  2. Evaluates the line with `isThematicBreak(line, reader.LineOffset())`.
  3. If valid:
     - Advances the reader position to the end of the line (`reader.AdvanceToEOL()`).
     - Returns a new AST node via `ast.NewThematicBreak()` along with the parser state `NoChildren`.
  4. If invalid:
     - Returns `nil, NoChildren`.

### `Continue`
```go
func (b *thematicBreakParser) Continue(_ ast.Node, _ text.Reader, _ Context) State
```
* **Returns:** `Close`
* **Purpose:** Indicates that a thematic break block is complete on the line it was opened, as thematic breaks cannot span multiple lines.

### `Close`
```go
func (b *thematicBreakParser) Close(_ ast.Node, _ text.Reader, _ Context)
```
* **Purpose:** Performs cleanup upon closing the block. This method contains no operations as no cleanup is required.

### `CanInterruptParagraph`
```go
func (b *thematicBreakParser) CanInterruptParagraph() bool
```
* **Returns:** `true`
* **Purpose:** Signals that a thematic break line is allowed to interrupt an existing paragraph block without requiring a preceding blank line.

### `CanAcceptIndentedLine`
```go
func (b *thematicBreakParser) CanAcceptIndentedLine() bool
```
* **Returns:** `false`
* **Purpose:** Specifies that this parser cannot accept indented lines beyond the standard block limit.