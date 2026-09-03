# Technical Documentation: `util/util.go`

## Overview

The `util/util.go` package provides low-level performance-focused utility functions and data structures for the Goldmark Markdown processor. It includes custom byte buffers, unicode and ASCII string/byte manipulations, HTML and URL escaping routines, sticky-error buffered writers, generic prioritized collections, and memory-efficient byte set filters.

---

## Key Data Structures and Interfaces

### 1. `CopyOnWriteBuffer`

A wrapper around a byte slice (`[]byte`) that implements a copy-on-write memory strategy to avoid unnecessary memory allocations when modifying read-only byte slices.

```go
type CopyOnWriteBuffer struct {
    buffer []byte
    copied bool
}
```

#### Methods
- **`NewCopyOnWriteBuffer(buffer []byte) CopyOnWriteBuffer`**: Initializes a new buffer wrapping the provided byte slice. `copied` is set to `false`.
- **`Write(value []byte)`**: Overwrites the existing buffer contents on first call by allocating a new slice of length 0 (with capacity `len(buffer) + 20`), sets `copied = true`, and appends `value`.
- **`WriteString(value string)`**: Calls `Write` using `StringToReadOnlyBytes(value)`.
- **`Append(value []byte)`**: Preserves existing buffer contents on first call by copying them to a newly allocated slice (with capacity `len(buffer) + 20`), sets `copied = true`, and appends `value`.
- **`AppendString(value string)`**: Calls `Append` using `StringToReadOnlyBytes(value)`.
- **`WriteByte(c byte) error`**: Overwrites on first call (allocating new capacity `len(buffer) + 20`) and appends single byte `c`.
- **`AppendByte(c byte)`**: Preserves existing content on first call by copying, then appends single byte `c`.
- **`Bytes() []byte`**: Returns the current slice.
- **`IsCopied() bool`**: Returns `true` if a new buffer allocation has occurred.

---

### 2. Buffered Writer Interfaces & Implementations

#### Interfaces
- **`BufWriter`**: A sub-interface of `bufio.Writer` requiring `io.Writer`, `WriteByte(byte) error`, `WriteRune(rune) (int, error)`, `WriteString(string) (int, error)`, and `Flush() error`.
- **`ErrorBufWriter`**: Extends `BufWriter` with `Error() error`, designed for sticky-error tracking where the first encountered error is captured and returned on subsequent write attempts.

#### Functions & Implementations
- **`NewErrorBufWriter(w io.Writer) ErrorBufWriter`**
- **`NewErrorBufWriterSize(w io.Writer, size int) ErrorBufWriter`**

##### Behavior:
- If `w` is `*bytes.Buffer`, returns a wrapped `bw` type (errors and flushes are no-ops returning `nil`).
- If `w` is `*strings.Builder`, returns a wrapped `sw` type (errors and flushes are no-ops returning `nil`).
- If `w` implements `BufWriter`, wraps it in an `errorBufWriter`.
- Otherwise, wraps `w` in `bufio.NewWriter` or `bufio.NewWriterSize` inside an `errorBufWriter`.

##### `errorBufWriter` Mechanics:
Stores a sticky error (`err`). If `err != nil`, calls to `Write`, `WriteByte`, `WriteRune`, `WriteString`, and `Flush` immediately return without operating on the underlying writer.

---

### 3. Generic Prioritized Collections

Structures used to maintain sorted sequences of items with associated priorities.

#### `PrioritizedValue[T any]`
Holds an arbitrary value `Value T` and an integer `Priority`.

#### `PrioritizedValues[T comparable]`
Slice type `[]PrioritizedValue[T]`.

#### Methods & Constructor
- **`Prioritized[T any](v T, priority int) PrioritizedValue[T]`**: Constructor helper.
- **`Sort()`**: Sorts the slice in-place in ascending order based on `Priority`.
- **`Remove(v T) PrioritizedValues[T]`**: Searches for the first element matching `Value == v` and removes it using `slices.Delete`.

---

### 4. `BytesFilter`

An interface and implementation for high-speed set-membership checks for byte slices and strings. It utilizes a combination of bitmask character filtering and hash bucket lookups.

```go
type BytesFilter interface {
    Add([]byte)
    AddString(string)
    Contains([]byte) bool
    ContainsString(string) bool
    Extend(...[]byte) BytesFilter
    ExtendString(string) BytesFilter
}
```

#### Implementation Details (`bytesFilter`)
- Uses a character bitmask table `chars [256]uint8` and a `threshold` (default `3`).
- Maps elements into 64 slot buckets (`slots [][][]byte`) using `bytesHash(b uint64)`.
- **`Add([]byte)`**: Sets bitmask indicators for the first `min(len(b), threshold)` bytes and appends the byte slice to the hash bucket slot.
- **`Contains([]byte) bool`**:
  1. Performs fast bitmask checking on the first `min(len(b), threshold)` bytes.
  2. Computes hash slot index and searches through matching byte slices in the bucket using `bytes.Equal`.
- **`Extend(...) / ExtendString(...)`**: Deep copies the current `bytesFilter` structures and inserts additional elements, returning a new `BytesFilter`.

---

## Function Categories

### 1. White Space and Indentation Analysis

- **`TabWidth(currentPos int) int`**: Calculates the remaining width to the next 4-space tab stop (`4 - currentPos%4`).
- **`IndentWidth(bs []byte, currentPos int) (width, pos int)`**: Iterates over leading spaces and tabs in `bs`, returning total visual width and number of bytes processed (`pos`).
- **`IndentPosition(bs []byte, currentPos, width int) (pos, padding int)`**: Wraps `IndentPositionPadding(bs, currentPos, 0, width)`.
- **`IndentPositionPadding(bs []byte, currentPos, paddingv, width int) (pos, padding int)`**: Finds byte offset `pos` and target offset padding `padding` needed to reach target visual `width`, accounting for initial `paddingv` and tab expansions. Returns `(-1, -1)` if `width` cannot be met.
- **`FirstNonSpacePosition(bs []byte) int`**: Returns the index of the first character that is not `' '` or `'\t'`. Returns `-1` if a newline (`'\n'`) or end-of-slice is encountered first.
- **`IsBlank(bs []byte) bool`**: Returns `true` if all bytes in `bs` satisfy `IsSpace(b)`.
- **`VisualizeSpaces(bs []byte) []byte`**: Replaces non-printable/whitespace bytes with printable debug labels (e.g., `[SPACE]`, `[TAB]`, `[NEWLINE]\n`, `[CR]`, `[VTAB]`, `[NUL]`, `[U+FFFD]`).

---

### 2. Trimming Utilities

Efficient trimming operations that slice byte arrays without re-allocating memory:

| Function | Description |
| :--- | :--- |
| **`TrimLeft(bs, b []byte)`** | Slices off leading bytes found in `b`. |
| **`TrimRight(bs, b []byte)`** | Slices off trailing bytes found in `b`. |
| **`TrimLeftLength(bs, b []byte)`** | Returns count of trimmed leading bytes. |
| **`TrimRightLength(bs, b []byte)`** | Returns count of trimmed trailing bytes. |
| **`TrimLeftSpace(bs []byte)`** | Slices off standard whitespace characters (`\s`, `\t`, `\n`, `\v`, `\f`, `\r`). |
| **`TrimRightSpace(bs []byte)`** | Slices off trailing standard whitespace characters. |
| **`TrimLeftSpaceLength(bs []byte)`** | Returns count of leading whitespace bytes. |
| **`TrimRightSpaceLength(bs []byte)`** | Returns count of trailing whitespace bytes. |

---

### 3. Text Transformations and Escaping

- **`DoFullUnicodeCaseFolding(v []byte) []byte`**: Performs Unicode case folding. Converts ASCII `A-Z` directly to `a-z`, and uses lookup table `unicodeCaseFoldings` for non-ASCII Unicode runes. Uses `CopyOnWriteBuffer` to avoid allocations when no changes are necessary.
- **`ReplaceSpaces(bs []byte, repl byte) []byte`**: Replaces contiguous sequences of whitespace characters with a single `repl` byte.
- **`ToLinkReference(v []byte) string`**: Normalizes Markdown link reference definitions: trims space, applies full Unicode case folding, and collapses space sequences into single spaces.
- **`EscapeHTMLByte(b byte) []byte`**: Returns escaped byte replacements for `"`, `&`, `<`, `>`, and `\x00` from `htmlEscapeTable`. Returns `nil` if `b` does not require escaping.
- **`EscapeHTML(v []byte) []byte`**: Scans slice `v` and replaces HTML-sensitive characters with entity representations using `CopyOnWriteBuffer`.
- **`URLEscape(v []byte) []byte`**: Percent-encodes characters unsafe for URLs while preserving existing valid `%xx` percent-encoded sequences, ASCII alphanumeric characters, and specific safe symbols dictated by `urlEscapeTable`. Space is encoded to `%20`, `&` to `&amp;`. UTF-8 multi-byte sequences are encoded via `url.QueryEscape`.

---

### 4. Character Classification Predicates

Lookup-table and boolean helper functions:

```go
UTF8Len(b byte) int8           // Returns byte length of UTF-8 character via utf8lenTable
IsPunct(c byte) bool           // Checks if ASCII byte is punctuation via punctTable
IsPunctRune(r rune) bool       // Checks unicode.IsSymbol(r) || unicode.IsPunct(r)
IsSpace(c byte) bool           // Checks spaceTable for whitespace bytes
IsSpaceRune(r rune) bool       // Wraps unicode.IsSpace(r)
IsNumeric(c byte) bool         // Checks '0' <= c <= '9'
IsHexDecimal(c byte) bool      // Checks hex characters (0-9, a-f, A-F)
IsAlphaNumeric(c byte) bool    // Checks ASCII letters and numbers
```

---

### 5. Rune & Reader Helpers

- **`ToRune(bs []byte, pos int) rune`**: Scans backward from `pos` to find the UTF-8 rune start byte, then decodes and returns the rune.
- **`ToValidRune(v rune) rune`**: Validates `v`. If invalid or `0`, returns Replacement Character `\uFFFD` (`0xFFFD`).
- **`ReadWhile(bs []byte, index [2]int, pred func(byte) bool) (int, bool)`**: Iterates `j` from `index[0]` to `index[1]`, testing `pred(bs[j])`. Returns the final index reached and `true` if at least one byte satisfied `pred`.