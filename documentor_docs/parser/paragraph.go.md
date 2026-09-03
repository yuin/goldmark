# Technical Documentation: `parser/paragraph.go`

## Overview

The `parser/paragraph.go` file defines the `paragraphParser` type and its associated methods. It forms a part of the Goldmark Markdown parsing framework (`github.com/yuin/goldmark`) under the `parser` package. 

Its primary purpose is to identify, construct, extend, and finalize paragraph AST (`ast.Paragraph`) nodes during the parsing of Markdown source text.

---

## Key Components

### 1. `paragraphParser`
```go
type paragraphParser struct {
}
```
An unexported struct that implements the `BlockParser` interface for standard Markdown paragraphs. It holds no internal state fields.

### 2. `defaultParagraphParser`
```go
var defaultParagraphParser = &paragraphParser{}
```
A package-level singleton instance of `paragraphParser` used to avoid redundant memory allocations.

### 3. `NewParagraphParser`
```go
func NewParagraphParser() BlockParser
```
A public constructor function that returns the singleton `defaultParagraphParser` instance as a `BlockParser`.

---

## Implementation Details & Methods

The `paragraphParser` implements the lifecycle methods required for block parsing:

### `Trigger() []byte`
* **Returns:** `nil`
* **Behavior:** Returning `nil` indicates that paragraphs do not require specific trigger bytes to initiate parsing. A paragraph can open on standard text lines.

---

### `Open(parent ast.Node, reader text.Reader, pc Context) (ast.Node, State)`
Determines if a new paragraph block should be created at the current reader position.

* **Logic:**
  1. Peeks at the current line using `reader.PeekLine()`.
  2. Checks if the line is blank via `util.IsBlank(line)`.
     * If **blank**: Returns `nil` and state `NoChildren`.
  3. If **not blank**:
     * Instantiates a new paragraph AST node (`ast.NewParagraph()`).
     * Trims leading white space from the line segment using `segment.TrimLeftSpace(reader.Source())` and appends it to the node source.
     * Advances the reader position to the end of the line using `reader.AdvanceToEOL()`.
     * Returns the created paragraph node and the state flag `NoChildren`.

---

### `Continue(node ast.Node, reader text.Reader, pc Context) State`
Handles subsequent lines for an already open paragraph block.

* **Logic:**
  1. Peeks at the current line using `reader.PeekLine()`.
  2. Checks if the line is blank via `util.IsBlank(line)`.
     * If **blank**: Returns the state `Close`, signaling that the paragraph end has been reached.
  3. If **not blank**:
     * Trims leading whitespace from the segment (`segment.TrimLeftSpace(reader.Source())`) and appends the source segment to the `ast.Paragraph` node.
     * Advances the reader to the end of the line (`reader.AdvanceToEOL()`).
     * Returns the combined state `Continue | NoChildren`.

---

### `Close(node ast.Node, reader text.Reader, pc Context)`
Finalizes the paragraph block when parsing completes or transitions to another block.

* **Logic:**
  1. Casts `node` to `*ast.Paragraph` and retrieves its slice of source segments (`lines`).
  2. If `lines` is non-empty:
     * Trims trailing spaces from the last line segment using `TrimRightSpace(reader.Source())`.
  3. If `lines` is empty (`len(lines) == 0`):
     * Removes the paragraph node from its parent (`node.Parent().RemoveChild(node)`).

---

### Configuration Flags

#### `CanInterruptParagraph() bool`
* **Returns:** `false`
* **Behavior:** Indicates that a standard paragraph parser itself cannot interrupt another open paragraph block.

#### `CanAcceptIndentedLine() bool`
* **Returns:** `false`
* **Behavior:** Indicates that the paragraph parser does not explicitly handle indented code block lines as paragraph content.

---

## Paragraph Parsing Lifecycle Summary

1. **Initialization:** The parser receives a `text.Reader`.
2. **Opening (`Open`):** If the current line is non-blank, an `ast.Paragraph` node is instantiated, leading spaces are trimmed, and source text is attached to the node.
3. **Continuation (`Continue`):** Successive lines are processed. Non-blank lines have their leading spaces trimmed and are appended to the paragraph. Blank lines signal the block to close.
4. **Closing (`Close`):** Trailing space on the final line of the paragraph is trimmed. If no source lines exist in the node, it is removed from the AST.