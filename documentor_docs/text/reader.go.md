# Technical Documentation: `text/reader.go`

The `text/reader.go` file provides standard interfaces and concrete implementations for reading, navigating, peeking, and matching text streams within the document processing pipeline. It defines mechanisms to operate both over raw contiguous byte slices (`Reader`) and over disconnected segment lists (`BlockReader`), with support for line-offset calculations, virtual space padding, UTF-8 rune decoding, and regular expression matching.

---

## Constants

```go
const invalidValue = -1
const EOF = byte(0xff)
```

* `invalidValue`: A sentinel value (`-1`) used internally to signify uninitialized or invalid segment indices/offsets.
* `EOF`: A sentinel byte value (`0xff`) returned by peeking operations when reading reaches or exceeds the end of the source buffer.

---

## Interfaces

### `Reader`

The `Reader` interface abstracts text navigation and retrieval over a source buffer. It embeds Go's standard `io.RuneReader`.

```go
type Reader interface {
	io.RuneReader
	Source() []byte
	ResetPosition()
	Peek() byte
	PeekLine() ([]byte, Segment)
	PrecedingCharacter() rune
	ValueBetween(start, stop int) MultiLineValue
	Decoder() Decoder
	LineOffset() int
	Position() (int, Segment)
	SetPosition(int, Segment)
	SetPadding(int)
	Advance(int)
	AdvanceAndSetPadding(int, int)
	AdvanceToEOL()
	AdvanceLine()
	SkipSpaces() (Segment, int, bool)
	SkipBlankLines() (Segment, int, bool)
	Match(reg *regexp.Regexp) bool
	FindSubMatch(reg *regexp.Regexp) [][]byte
}
```

#### Method Summary
* **`Source() []byte`**: Returns the underlying byte slice bound to the reader.
* **`ResetPosition()`**: Resets all internal indices and positions back to the beginning of the text source.
* **`Peek() byte`**: Returns the byte at the current reading position without advancing the internal pointer. Returns `space[0]` if virtual padding exists, or `EOF` if past the end of source.
* **`PeekLine() ([]byte, Segment)`**: Returns the current line's byte slice and its corresponding `Segment` without advancing the pointer.
* **`PrecedingCharacter() rune`**: Decodes and returns the UTF-8 rune located directly before the current reading position.
* **`ValueBetween(start, stop int) MultiLineValue`**: Returns a decoded `MultiLineValue` covering the specified byte range `[start, stop)`.
* **`Decoder() Decoder`**: Returns the `Decoder` instance attached to the reader.
* **`LineOffset() int`**: Calculates and returns the visual distance (taking tab stops into account via `util.TabWidth`) from the beginning of the line to the current position, subtracting padding.
* **`Position() (int, Segment)`**: Returns the current 0-indexed line number and current line `Segment`.
* **`SetPosition(int, Segment)`**: Sets the current line number and line `Segment`.
* **`SetPadding(int)`**: Sets a virtual space padding count on the current line position.
* **`Advance(int)`**: Advances the reading pointer forward by $n$ bytes or virtual padding spaces.
* **`AdvanceAndSetPadding(n int, padding int)`**: Advances the pointer by $n$ bytes and applies the specified padding if it exceeds the remaining padding.
* **`AdvanceToEOL()`**: Advances the pointer to the end of the current line (just before or at the newline/EOF).
* **`AdvanceLine()`**: Moves the pointer directly to the head of the next line.
* **`SkipSpaces()`**: Advances past horizontal space characters on the current line.
* **`SkipBlankLines()`**: Advances past fully blank or whitespace-only lines.
* **`Match(reg *regexp.Regexp) bool`**: Checks if the line matches a regular expression at the current position, advancing the pointer on match.
* **`FindSubMatch(reg *regexp.Regexp) [][]byte`**: Searches for regular expression submatches on the line, advancing the pointer to the end of the full match and returning submatch byte slices.

---

### `BlockReader`

`BlockReader` extends `Reader` to support reading non-contiguous document fragments or lists of segments that make up a Markdown block.

```go
type BlockReader interface {
	Reader
	Reset(segs []Segment)
}
```

* **`Reset(segs []Segment)`**: Replaces the active segment slice with `segs` and resets position tracking.

---

## Implementations

### 1. `reader`

`reader` is the primary implementation of `Reader`, operating on a continuous byte slice (`source`).

#### Struct Definition
```go
type reader struct {
	source       []byte
	sourceLength int
	line         int
	peekedLine   []byte
	pos          Segment
	head         int
	lineOffset   int
	decoder      Decoder
}
```

#### Constructor
```go
func NewReader(b []byte, decoder Decoder) Reader
```
Creates a `reader` initialized with source byte slice `b` and a decoder. Calls `ResetPosition()` before returning.

#### Key Implementation Details
* **Position Tracking**: Line indices are managed via `pos` (`Segment`), where `pos.Start` tracks the current byte index and `pos.Stop` tracks the end of the current line (including `\n`).
* **Line Navigation (`AdvanceLine`)**: Scans forward from `pos.Start` using `bytes.IndexByte` to locate newline characters (`\n`) and updates `pos.Stop` to encompass the line.
* **Virtual Padding**: `pos.Padding` tracks virtual space characters added to the line. When padding is $> 0$, `Peek()` returns a space character (`' '`), and advancing decrements padding before consuming source bytes.
* **Cached Peeking (`PeekLine`)**: Caches the current line slice in `peekedLine`. Any structural shift or pointer advance invalidates `peekedLine` by setting it to `nil`.

---

### 2. `blockReader`

`blockReader` implements `BlockReader`. It navigates across an array of discrete `Segment` items rather than relying purely on searching raw bytes for line breaks.

#### Struct Definition
```go
type blockReader struct {
	source     []byte
	segments   []Segment
	line       int
	pos        Segment
	head       int
	last       int
	lineOffset int
	decoder    Decoder
}
```

#### Constructor
```go
func NewBlockReader(source []byte, segs []Segment, decoder Decoder) BlockReader
```
Creates a `blockReader` over the given `source` byte slice and block `segs`.

#### Key Implementation Details
* **Segment Iteration**: Lines map to individual entries in `segments`. Setting position (`SetPosition`) looks up the pre-calculated segment at `segments[line]`.
* **`ValueBetween(start, stop int) MultiLineValue`**: Builds a multi-segment value by iterating over internal `segments`, matching segments that fall within `[start, stop)`, clipping boundaries as needed, and aggregating them using a `ValueBuilder`.
* **Line Advancement (`AdvanceLine`)**: Increments `line` count and sets the current position segment to invalid range `(-1, -1)`, which lazily loads the next segment on sub-sequent calls to `SetPosition` or position restoration.

---

## Non-Exported Helper Functions

The file provides internal helper functions used by both `reader` and `blockReader` implementations to reduce duplicate logic.

### `readRuneReader(r Reader) (rune, int, error)`
Reads a single UTF-8 rune using the `Reader`'s `PeekLine()` interface.
* Decodes the rune at the head of the current line via `utf8.DecodeRune`.
* Returns `io.EOF` if no line bytes exist or if a `utf8.RuneError` occurs.
* On success, calls `r.Advance(size)` and returns `(rune, size, nil)`.

### `skipSpacesReader(r Reader) (Segment, int, bool) -> (Segment, int, bool)`
Advances the reader past space characters (`util.IsSpace`).
* Iterates over bytes from `PeekLine()`.
* Calls `r.Advance(1)` for each space encountered.
* Returns the modified segment starting position, total skipped character count, and a `bool` indicating whether a non-space character was found (`false` if EOF was hit).

### `skipBlankLinesReader(r Reader) (Segment, int, bool)`
Advances the reader past lines containing only whitespace (`util.IsBlank`).
* Calls `PeekLine()` repeatedly.
* If a line is blank according to `util.IsBlank`, increments line counter and calls `r.AdvanceLine()`.
* Returns the current `Segment`, total skipped line count, and `true` when encountering a non-blank line (or `false` if EOF).

### `matchReader(r Reader, reg *regexp.Regexp) bool`
Tests if a regular expression matches the text stream at the current position.
* Saves initial line position using `r.Position()`.
* Runs `reg.FindReaderSubmatchIndex(r)`.
* Restores position back to original state using `r.SetPosition()`.
* If a match is found, calls `r.Advance(match[1] - match[0])` to advance past the full match length and returns `true`. Returns `false` otherwise.

### `findSubMatchReader(r Reader, reg *regexp.Regexp) [][]byte`
Performs regular expression matching and extracts all capturing submatches as byte slices.
* Saves initial position, executes `reg.FindReaderSubmatchIndex(r)`, and immediately restores initial position.
* If matched:
  1. Reads runes up to the total match length (`match[1]`) into a `bytes.Buffer`.
  2. Extracts slice pairs corresponding to submatch offsets `[match[i]:match[i+1]]`.
  3. Appends empty byte slices `[]byte{}` for unmatched/optional subgroups (`match[i] < 0`).
  4. Restores position again and advances by the overall match length (`match[1] - match[0]`).
* Returns the slice of byte slices `[][]byte`.