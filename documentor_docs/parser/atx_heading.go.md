# Technical Documentation: `parser/atx_heading.go`

## Overview

The `parser/atx_heading.go` file implements the parsing logic for ATX headings in a Markdown document (e.g., `# Heading 1`, `## Heading 2`). It provides the implementation of a block parser (`atxHeadingParser`) that recognizes ATX heading syntax, handles optional attributes, handles closing `#` characters, and manages heading ID generation (auto-generated or custom).

---

## Data Structures & Configuration

### `HeadingConfig`
```go
type HeadingConfig struct {
    autoHeadingID bool
    attribute     bool
}
```
A configuration struct embedded within heading parsers.
- `autoHeadingID`: Indicates whether auto-generated heading IDs should be assigned during parsing.
- `attribute`: Indicates whether block attributes on headings should be parsed.

### `HeadingOption`
```go
type HeadingOption interface {
    setHeadingOption(*HeadingConfig)
}
```
An interface used to apply configuration options specifically to heading parsers.

### `withAutoHeadingID`
```go
type withAutoHeadingID struct{}
```
An internal struct that implements both general parser options (`Option`) and `HeadingOption`.
- `SetParserOption(c *Config)`: Enables `autoHeadingID` on global parser configurations.
- `setHeadingOption(p *HeadingConfig)`: Enables `autoHeadingID` on `HeadingConfig`.

---

## Core Parser Component

### `atxHeadingParser`
```go
type atxHeadingParser struct {
    HeadingConfig
}
```
The primary block parser struct that handles ATX headings. It embeds `HeadingConfig` and implements the `BlockParser` interface.

### `NewATXHeadingParser`
```go
func NewATXHeadingParser(opts ...HeadingOption) BlockParser
```
Constructs and returns a new `BlockParser` initialized to parse ATX headings. Configures options passed via `opts`.

---

## `BlockParser` Interface Implementation

### `Trigger() []byte`
- **Returns:** `[]byte{'#'}`
- Defines `#` as the trigger character for this parser.

### `Open(_ ast.Node, reader text.Reader, pc Context) (ast.Node, State)`
Performs the parsing of an ATX heading line.

#### Key Steps in `Open`:
1. **Offset & Character Count**: Obtains the block offset from `pc.BlockOffset()`. Counts the consecutive `#` characters starting at that offset.
2. **Heading Level Validation**:
   - Calculates `level = count`.
   - If `level == 0` or `level > 6`, it returns `nil, NoChildren` (ATX headings support 1 to 6 `#` characters).
3. **Space Check**:
   - Checks for whitespace immediately following the opening `#` characters using `util.TrimLeftSpaceLength`.
   - If there is no space following the `#` characters (and it is not at the end of the line), parsing fails and returns `nil, NoChildren`.
4. **Node Creation**:
   - Creates an `ast.Heading` node (`ast.NewHeading(level, ast.HeadingKindATX)`).
5. **Attribute Handling**:
   - If `b.attribute` is `true`, calls `parseLastLineAttributes` to parse trailing attribute syntax (`{...}`) at the end of the heading line.
6. **Closing `#` Sequence Handling**:
   - Inspects the end of the line for optional trailing `#` sequences (e.g., `## Heading ##`).
   - Strips trailing `#` characters and preceding spaces from the content segment.
   - If the heading content consists solely of closing `#` sequences (e.g., `### ###`), advances to the end of the line and returns the empty heading node.
7. **Finalization**:
   - Appends the adjusted line segment to the heading node using `node.AppendSource(hl)`.
   - Advances the reader to the end of the line (`reader.AdvanceToEOL()`).
   - Returns `node, NoChildren`.

### `Continue(_ ast.Node, _ text.Reader, _ Context) State`
- **Returns:** `Close`
- ATX headings are single-line blocks and cannot continue across multiple lines.

### `Close(node ast.Node, reader text.Reader, pc Context)`
Invoked when the ATX heading block is closed.
- If `b.autoHeadingID` is enabled:
  - Checks if the heading node already contains an `"id"` attribute.
  - If no `"id"` attribute exists, calls `generateAutoHeadingID`.
  - If an `"id"` attribute is present, registers it with the context's ID pool via `pc.IDs().Put(...)`.

### `CanInterruptParagraph() bool`
- **Returns:** `true`
- ATX headings are allowed to interrupt an existing paragraph block.

### `CanAcceptIndentedLine() bool`
- **Returns:** `false`
- ATX headings cannot be started on indented lines.

---

## Helper Functions

### `WithAutoHeadingID`
```go
func WithAutoHeadingID() interface {
    Option
    HeadingOption
}
```
Returns a `withAutoHeadingID` instance. Can be passed as an option during parser initialization to enable auto heading ID generation.

### `generateAutoHeadingID`
```go
func generateAutoHeadingID(node *ast.Heading, reader text.Reader, pc Context)
```
1. Retrieves the source bytes for the heading content.
2. Generates a unique ID using `pc.IDs().Generate(line, ast.KindHeading)`.
3. Sets the `"id"` attribute on `node` using `text.NewMultiLineValueFromString`.

### `parseLastLineAttributes`
```go
func parseLastLineAttributes(node ast.BlockNode, reader text.Reader, _ Context)
```
1. Reads the source segment of the heading.
2. Scans for `{` characters representing the start of block attributes (ignoring escaped characters `\` inside the line).
3. Calls `ParseAttributes(lr)` to attempt parsing attribute pairs inside the braces.
4. If attributes are valid and followed only by trailing blank space or newline:
   - Applies the parsed attributes to `node` via `node.SetAttribute(...)`.
   - Trims the attribute syntax out of the heading's source segment.