# Technical Documentation: `parser/code_span.go`

## Overview

The `parser/code_span.go` file implements the parsing logic for Markdown inline code spans (e.g., `` `code` `` or ``` ``code`` ```). It detects opening and closing backtick delimiters, constructs abstract syntax tree (AST) nodes for inline code spans, and handles text normalization according to the CommonMark specification (e.g., replacing newlines with spaces and trimming leading/trailing whitespace).

---

## Architecture & Key Components

### 1. Parser Definition

#### `codeSpanParser`
An unexported struct that implements the `InlineParser` interface to process inline code spans.

* **`defaultCodeSpanParser`**: A package-level instance of `codeSpanParser` used to avoid redundant allocations.
* **`NewCodeSpanParser() InlineParser`**: Returns `defaultCodeSpanParser`.

#### `Trigger() []byte`
Returns `[]byte{'`'}`, registering the backtick character (``` ` ```) as the trigger byte for inline code span parsing.

#### `Parse(parent ast.Node, block text.Reader, pc Context) ast.Node`
Executes the primary parsing logic for inline code spans:
1. **Opener Detection**: Reads continuous backticks at the trigger position to determine the opening delimiter length (`opener`).
2. **Scan for Closure**: Scans subsequent characters across lines looking for a matching run of backticks of equal length (`closure == opener`) that is not immediately followed by another backtick.
3. **Fallback Handling**: If EOF is reached without finding a matching closing sequence, the reader position is reset to immediately after the opening backticks, and an `ast.Text` node representing the unmatched opening backticks is returned.
4. **AST Construction**: If a matching closure is found, it constructs either a `singleLineCodeSpanValue` or a `multiLineCodeSpanValue` based on the number of text indices collected, returning a new `ast.Node` created via `ast.NewCodeSpan(...)`.

---

### 2. Custom Text Value Wrappers

Code spans require special text processing (trimming outer space and replacing newlines with spaces). This file defines two implementations of the `text.Value` interface to defer or perform this normalization efficiently.

```
       text.Value (Interface)
         /               \
        /                 \
singleLineCodeSpanValue   multiLineCodeSpanValue
```

#### `singleLineCodeSpanValue`
An allocation-optimized implementation of `text.Value` used when the code span spans a single contiguous slice of text.

* **Fields**:
  * `start int`: Starting byte offset in the source buffer.
  * `stop int`: Ending byte offset in the source buffer.
* **Key Characteristics**: Small struct size designed to fit directly inside an interface variable without heap allocations.
* **Methods**:
  * `Index() text.Index`: Returns `text.NewIndex(v.start, v.stop)`.
  * `Indices() []text.Index`: Returns a slice containing `Index()`.
  * `IsEmpty() bool`: Returns `true` if `start >= stop`.
  * `IsOwned() bool`: Returns `false`.
  * `Value(source []byte) string`: Returns normalized content as a string.
  * `Str(source []byte) string`: Returns normalized content as a string using `util.BytesToReadOnlyString`.
  * `Bytes(source []byte) []byte`: Returns trimmed content with newlines converted to spaces.
  * `WriteTo(w io.Writer, source []byte) (int, error)`: Writes trimmed content to an `io.Writer`, converting newlines to spaces.
  * `shouldTrimSpaces(source []byte) bool`: Returns `true` if the string length is $\ge 2$, is not entirely blank, and both the first and last characters are spaces or newlines.
  * `trimmedIndex(source []byte) text.Index`: Returns a new `text.Index` with offset boundaries adjusted (+1 start, -1 stop) if `shouldTrimSpaces` is `true`.

#### `multiLineCodeSpanValue`
An implementation of `text.Value` used when a code span spans across multiple lines or non-contiguous text indices.

* **Fields**:
  * `indices []text.Index`: Slice of source text index ranges.
* **Methods**:
  * `Index() text.Index`: Returns the first index in `indices`, or `text.NewIndex(0, 0)` if empty.
  * `Indices() []text.Index`: Returns `v.indices`.
  * `IsEmpty() bool`: Returns `true` if all index segments are empty (`Start >= Stop`).
  * `IsOwned() bool`: Returns `false`.
  * `Value(source []byte) string`: Returns normalized content as a string.
  * `Str(source []byte) string`: Returns normalized content as a string using `util.BytesToReadOnlyString`.
  * `Bytes(source []byte) []byte`: Concatenates normalized byte slices across all indices, replacing newlines with spaces.
  * `WriteTo(w io.Writer, source []byte) (int, error)`: Writes all index segments to `w`, converting newlines to spaces on the fly.
  * `shouldTrimSpaces(source []byte) bool`: Determines if spaces should be trimmed across multi-line index segments by checking the first character of the first segment and the last character of the last segment, provided at least one segment contains non-blank text.
  * `trimmedIndices(source []byte) []text.Index`: Returns indices with start incremented on the first index and stop decremented on the last index if trimming rules apply.

---

### 3. Utility Functions

#### `isSpaceOrNewline(c byte) bool`
Returns `true` if character `c` is a space (`' '`) or a newline (`'\n'`).

#### `replaceNewlinesWithSpaces(b []byte) []byte`
Replaces all occurrences of `\n` in `b` with a space character (`' '`) using `bytes.ReplaceAll`.

#### `writeNewlinesWithSpaces(w io.Writer, b []byte) (int, error)`
Streams byte slice `b` to `w` segment by segment, writing space bytes (`' '`) in place of newline characters (`'\n'`), avoiding unnecessary full buffer allocations.

---

## Workflow: Parsing an Inline Code Span

1. **Triggering**: The parser triggers when encountering a backtick (`'` `').
2. **Counting Openers**: Counts total consecutive backticks at start position (e.g., `` `` `` = 2).
3. **Iterating Source**: Advances line by line through `text.Reader`.
4. **Matching Closure**: Search line for backtick sequences.
   * If a sequence matches the exact length of `opener` and is not followed by additional backticks, the end of the code span is reached.
5. **Evaluating Match**:
   * **No match (EOF)**: Reader position reverts to immediately after opening backticks; returns an `ast.Text` node.
   * **Match found**:
     * If 1 index range captured $\rightarrow$ returns `ast.NewCodeSpan` with `singleLineCodeSpanValue`.
     * If $>1$ index ranges captured $\rightarrow$ returns `ast.NewCodeSpan` with `multiLineCodeSpanValue`.
6. **Value Retrieval**: During rendering or AST traversal, calling `.Bytes()`, `.Str()`, or `.WriteTo()` automatically trims leading/trailing spaces and replaces intermediate `\n` characters with space characters.