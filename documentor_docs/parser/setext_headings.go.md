# Technical Documentation: `parser/setext_headings.go`

## Overview

The `parser/setext_headings.go` file implements the parsing logic for Setext-style Markdown headings in the Goldmark Markdown engine. Setext headings are defined by underlining text with `=` (level 1 heading) or `-` (level 2 heading) characters on the following line.

This package provides the `setextHeadingParser` struct, which implements the `BlockParser` interface to detect, parse, and construct Setext heading AST nodes from raw Markdown source text.

---

## Global Variables

### `temporaryParagraphKey`
* **Type:** `ContextKey`
* **Description:** A context key used during parsing to temporarily store a reference to the preceding `ast.Paragraph` node within the parser context (`Context`). This reference is held between the `Open` and `Close` parsing phases.

---

## Data Structures

### `setextHeadingParser`
* **Description:** An unexported struct implementing the `BlockParser` interface for Setext headings.
* **Fields:**
  * `HeadingConfig`: Embedded configuration struct containing settings such as auto heading ID generation (`autoHeadingID`) and attribute parsing (`attribute`).

---

## Constructor

### `NewSetextHeadingParser`

```go
func NewSetextHeadingParser(opts ...HeadingOption) BlockParser
```

* **Purpose:** Constructs and initializes a new `BlockParser` instance configured to parse Setext headings.
* **Parameters:**
  * `opts ...HeadingOption`: Functional option arguments used to configure the parser's embedded `HeadingConfig`.
* **Returns:** A `BlockParser` interface implementation (`*setextHeadingParser`).

---

## Helper Functions

### `matchesSetextHeadingBar`

```go
func matchesSetextHeadingBar(line []byte) (byte, bool)
```

* **Purpose:** Evaluates whether a given raw line of text matches the criteria for a Setext heading underline bar (`=` or `-`).
* **Rules:**
  1. Trims leading spaces (up to 3 spaces allowed; if > 3 spaces exist, returns `0, false`).
  2. Counts consecutive `=` characters (`level1`). If zero, counts consecutive `-` characters (`level2`).
  3. Trims trailing whitespace.
  4. Validates that the entire line content (excluding permitted leading/trailing whitespace) consists exclusively of `=` or `-` characters.
* **Returns:**
  * `byte`: The character used (`'='` or `'-'`), or `0` if invalid.
  * `bool`: `true` if the line matches a valid Setext heading bar, otherwise `false`.

---

## `BlockParser` Interface Implementation

### `Trigger`

```go
func (b *setextHeadingParser) Trigger() []byte
```

* **Purpose:** Defines the trigger characters that prompt the block parser engine to evaluate this parser.
* **Returns:** `[]byte{'-', '='}`

---

### `Open`

```go
func (b *setextHeadingParser) Open(parent ast.Node, reader text.Reader, pc Context) (ast.Node, State)
```

* **Purpose:** Evaluates whether the current position in the input document begins a Setext heading.
* **Execution Steps:**
  1. Retrieves the last opened block from `pc.LastOpenedBlock().Node`. If `nil`, returns `nil, NoChildren`.
  2. Verifies that the last opened block is an `*ast.Paragraph` and that its parent matches `parent`. If not, returns `nil, NoChildren`.
  3. Peeks at the current line from `reader`.
  4. Calls `matchesSetextHeadingBar(line)`. If it returns `false`, parsing aborts (`nil, NoChildren`).
  5. Determines heading level: Level 1 for `'='`, Level 2 for `'-'`.
  6. Creates a new heading AST node (`ast.NewHeading(level, ast.HeadingKindSetext)`).
  7. Appends the current line source segment to the heading node.
  8. Saves the preceding paragraph node reference into context under `temporaryParagraphKey`.
* **Returns:** The newly created `ast.Node` and state flags `NoChildren | RequireParagraph`.

---

### `Continue`

```go
func (b *setextHeadingParser) Continue(_ ast.Node, _ text.Reader, _ Context) State
```

* **Purpose:** Determines the continuation state of the Setext heading block.
* **Returns:** Always returns `Close`, as Setext heading underlines are single-line constructs.

---

### `Close`

```go
func (b *setextHeadingParser) Close(node ast.Node, reader text.Reader, pc Context)
```

* **Purpose:** Finalizes the Setext heading AST node, converting the preceding paragraph into the heading's content or cleaning up empty blocks.
* **Execution Steps:**
  1. Casts `node` to `*ast.Heading` and retrieves its initial source segment.
  2. Retrieves and clears the temporary paragraph stored under `temporaryParagraphKey`.
  3. **Empty Source Handling (`len(tmp.Source()) == 0`):**
     * Trims leading whitespace from the underline bar segment.
     * If the next sibling node is not a paragraph, creates a new paragraph with the trimmed segment and places it after `heading`.
     * If the next sibling node is a paragraph, prepends the trimmed segment to that paragraph's source segments.
     * Removes the `heading` node from its parent.
  4. **Non-Empty Source Handling:**
     * Transfers source segments, starting position, and blank line properties from the preceding paragraph (`tmp`) to `heading`.
     * Removes the preceding paragraph (`tmp`) from its parent.
  5. **Attribute Handling:**
     * If `b.attribute` is enabled, calls `parseLastLineAttributes(heading, reader, pc)` to process inline block attributes.
  6. **Heading ID Generation:**
     * If `b.autoHeadingID` is enabled:
       * If an `id` attribute is already set, registers it with `pc.IDs().Put(...)`.
       * If no `id` attribute exists, calls `generateAutoHeadingID(heading, reader, pc)`.

---

### `CanInterruptParagraph`

```go
func (b *setextHeadingParser) CanInterruptParagraph() bool
```

* **Purpose:** Indicates whether this parser can interrupt an existing paragraph block.
* **Returns:** `true`

---

### `CanAcceptIndentedLine`

```go
func (b *setextHeadingParser) CanAcceptIndentedLine() bool
```

* **Purpose:** Indicates whether the Setext underline bar can be indented by 4 spaces or more.
* **Returns:** `false`