# Technical Documentation: `parser/raw_html.go`

## Overview

The `parser/raw_html.go` file implements an inline HTML parser for the Goldmark Markdown parsing engine. Its primary purpose is to identify, parse, and process raw HTML constructs embedded within Markdown text. 

When a trigger character (`<`) is encountered during inline parsing, this parser inspects the text to determine whether it matches standard HTML structures—such as opening tags, closing tags, comments, processing instructions, declarations, or CDATA sections. If a match is found, it constructs an AST node (`ast.RawHTML`) representing the raw HTML slice and advances the text reader position.

---

## Constants and Package Variables

### Singletons and Instances
* **`defaultRawHTMLParser`**: A package-private singleton instance of `rawHTMLParser`.
* **`NewRawHTMLParser() InlineParser`**: Public constructor function returning `defaultRawHTMLParser` as an `InlineParser` interface.

### Regular Expressions
* **`tagnamePattern`**: Matches valid tag names (`[A-Za-z][A-Za-z0-9-]*`).
* **`spaceOrOneNewline`**: Matches whitespace or at most one newline sequence (`(?:[ \t]|(?:\r\n|\n){0,1})`).
* **`attributePattern`**: Matches HTML attribute key-value pairs, supporting unquoted, single-quoted, and double-quoted attribute values.
* **`openTagRegexp`**: Compiled regular expression matching opening HTML tags, including self-closing tags (`<tag attr="val">`, `<tag />`).
* **`closeTagRegexp`**: Compiled regular expression matching closing HTML tags (`</tag>`).

### Delimiters and Byte Slices
* **`openProcessingInstruction` (`<?`) / `closeProcessingInstruction` (`?>`)**: Processing instruction bounds.
* **`openCDATA` (`<![CDATA[`) / `closeCDATA` (`]]>`)**: CDATA block bounds.
* **`closeDecl` (`>`)**: Declaration end delimiter.
* **`openComment` (`<!--`) / `closeComment` (`-->`)**: Standard HTML comment bounds.
* **`emptyComment1` (`<!-->`) / `emptyComment2` (`<!--->`)**: Edge-case short/empty HTML comments.

---

## Core Types and Methods

### `rawHTMLParser` Struct

`rawHTMLParser` is an unexported struct implementing the `InlineParser` interface.

#### `Trigger() []byte`
Returns the trigger byte slice: `[]byte{'<'}`. The inline parsing engine calls `Parse()` when it encounters a `<` byte in the source document.

#### `Parse(parent ast.Node, block text.Reader, pc Context) ast.Node`
The main entry point for parsing raw HTML. It inspects the character(s) immediately following the `<` trigger on the current line and delegates parsing to specific helper methods based on the prefix:

1. **Open HTML Tag**: If line length $> 1$ and `line[1]` is alphanumeric $\rightarrow$ calls `parseMultiLineRegexp(openTagRegexp, block, pc)`.
2. **Close HTML Tag**: If line length $> 2$, `line[1] == '/'`, and `line[2]` is alphanumeric $\rightarrow$ calls `parseMultiLineRegexp(closeTagRegexp, block, pc)`.
3. **Comment**: If line starts with `<!--` (`openComment`) $\rightarrow$ calls `parseComment(block, pc)`.
4. **Processing Instruction**: If line starts with `<?` (`openProcessingInstruction`) $\rightarrow$ calls `parseUntil(block, closeProcessingInstruction, pc)`.
5. **Declaration**: If line length $> 2$, `line[1] == '!'`, and `line[2]` is an uppercase ASCII letter (`A-Z`) $\rightarrow$ calls `parseUntil(block, closeDecl, pc)`.
6. **CDATA Section**: If line starts with `<![CDATA[` (`openCDATA`) $\rightarrow$ calls `parseUntil(block, closeCDATA, pc)`.
7. **No Match**: Returns `nil` if none of the above rules match.

---

## Helper Parsing Functions

### `parseComment(block text.Reader, _ Context) ast.Node`
Parses HTML comments, handling both single-line and multi-line comments as well as short comment edge cases.

* **Edge Case Check**:
  * Checks if the line starts with `<!-->` (`emptyComment1`) or `<!--->` (`emptyComment2`). If matched, it advances the block reader by the length of the comment prefix and returns an `ast.RawHTML` node wrapping the single-line segment index.
* **Multi-line / Standard Comment Search**:
  * Offsets past the opening `<!--` prefix.
  * Continuously searches line by line for `-->` (`closeComment`).
  * Collects `text.Index` values for each line spanned by the comment.
  * If `-->` is found:
    * Advances the reader past the closing tag.
    * Returns an `ast.RawHTML` node constructed from the collected indices using `text.IdentityDecoder`.
  * If the end of input is reached without finding `-->`:
    * Restores the block reader position to its saved state (`savedLine`, `savedSegment`).
    * Returns `nil`.

### `parseUntil(block text.Reader, closer []byte, _ Context) ast.Node`
A general-purpose delimiter-matching parser used for CDATA, declarations, and processing instructions.

* **Execution Flow**:
  1. Saves current reader position (`savedLine`, `savedSegment`).
  2. Iterates line-by-line using `block.PeekLine()`.
  3. Checks for the occurrence of the `closer` byte slice within the line using `bytes.Index`.
  4. Collects segment indices (`text.NewIndex`) for each line scanned.
  5. If `closer` is found:
     * Constructs a final index ending at the closer boundary.
     * Advances the block reader past `closer`.
     * Returns an `ast.RawHTML` node constructed from the line indices.
  6. If reader hits EOF without matching `closer`:
     * Restores reader position using `block.SetPosition`.
     * Returns `nil`.

### `parseMultiLineRegexp(reg *regexp.Regexp, block text.Reader, _ Context) ast.Node`
Parses HTML elements (open and close tags) that match regular expressions across one or more lines.

* **Execution Flow**:
  1. Records initial position (`sline`, `ssegment`).
  2. Calls `block.Match(reg)` to check if the regular expression matches starting from the current position.
  3. If matched, records the ending reader position (`eline`, `esegment`), then temporarily resets to `sline`, `ssegment`.
  4. Iterates through the block line-by-line from `sline` to `eline`:
     * Calculates `start` byte index (adjusts for initial line segment start).
     * Calculates `end` byte index (adjusts for final line segment end).
     * Appends each computed line index slice to `indices`.
     * Advances the reader line-by-line until reaching `eline`.
  5. Advances the block reader by `end - start` on the final line.
  6. Returns an `ast.RawHTML` node constructed from `indices` using `text.IdentityDecoder`.
  7. If `block.Match` returns `false`, returns `nil`.

---

## Dependency Summary

* **`bytes`**: Used for prefix checks (`bytes.HasPrefix`) and finding delimiter occurrences (`bytes.Index`).
* **`regexp`**: Used for matching complex tag and attribute patterns spanning across text boundaries.
* **`github.com/yuin/goldmark/v2/ast`**: Provides the `ast.Node` interface and `ast.NewRawHTML` constructor.
* **`github.com/yuin/goldmark/v2/text`**: Provides source text line reading primitives (`text.Reader`), slice indices (`text.Index`), and decoders (`text.IdentityDecoder`).
* **`github.com/yuin/goldmark/v2/util`**: Provides helper functions such as `util.IsAlphaNumeric`.