# Technical Documentation: `text/decoder.go`

## Overview

The `text/decoder.go` file provides text decoding functionality for CommonMark/Markdown content processing. Its primary objective is to convert backslash-escaped characters and HTML/numeric entity references into their unescaped or resolved UTF-8 byte representations.

> **Note:** As specified in the code, this package performs **decoding** (e.g., entity resolution, escape sequence removal) rather than **normalization** (such as trimming or normalizing whitespace inside code spans).

---

## Interfaces

### `Decoder`

`Decoder` defines the standard interface for decoding byte slices or writing decoded output directly to a stream.

```go
type Decoder interface {
    Decode(b []byte) []byte
    DecodeTo(w io.Writer, b []byte) (int, error)
}
```

#### Methods
* **`Decode(b []byte) []byte`**: Accepts a byte slice `b`, performs decoding, and returns the decoded byte slice.
* **`DecodeTo(w io.Writer, b []byte) (int, error)`**: Decodes byte slice `b` and writes the resulting decoded bytes directly into the provided `io.Writer`. Returns the number of bytes written and any error encountered.

---

## Configuration & Options

### `decoderConfig`

An internal configuration struct that holds option flags for decoders.

```go
type decoderConfig struct {
    EscapedSpace bool
}
```

### `DecoderOption`

A functional option type used to configure a `decoderConfig`.

```go
type DecoderOption func(*decoderConfig)
```

### `WithEscapedSpace`

```go
func WithEscapedSpace() DecoderOption
```

Returns a `DecoderOption` that sets `EscapedSpace` to `true`. When enabled, the decoder treats an escaped space (`\ `) as an empty character (removing both the backslash and the space from the decoded output).

---

## Implementations

### 1. `DefaultDecoder`

`DefaultDecoder` is the standard implementation of the `Decoder` interface. It resolves HTML entities, numeric character references, and backslash escapes according to CommonMark specifications.

```go
type DefaultDecoder struct {
    cfg decoderConfig
}
```

#### Constructor

##### `NewDecoder`
```go
func NewDecoder(opts ...DecoderOption) *DefaultDecoder
```
Constructs a new `DefaultDecoder` instance, applying any provided `DecoderOption` functions.

#### Supported Transformation Rules
* **Backslash Escapes:** Decodes ASCII punctuation preceded by a backslash (e.g., `\*` -> `*`).
  * If configured with `WithEscapedSpace()`, it decodes `\ ` to an empty character.
* **Named HTML5 Entities:** Decodes standard HTML5 entity references (e.g., `&amp;` -> `&`).
* **Hexadecimal Numeric References:** Decodes hex code points (e.g., `&#x26;` or `&#X26;` -> `&`). Max code point representation length is 6 hex digits.
* **Decimal Numeric References:** Decodes decimal code points (e.g., `&#38;` -> `&`). Max length is 7 decimal digits.

---

### 2. `IdentityDecoder`

`IdentityDecoder` is a predefined `Decoder` instance that performs a no-op pass-through, returning or writing input byte slices unchanged.

```go
var IdentityDecoder Decoder = &identityDecoder{}
```

#### Internal Type: `identityDecoder`
* **`Decode(b []byte) []byte`**: Returns `b` without modifications.
* **`DecodeTo(w io.Writer, b []byte) (int, error)`**: Directly calls `w.Write(b)`.

---

## Detailed Execution Logic

### `DefaultDecoder.Decode(b []byte) []byte`

1. **Fast Path Check:**
   Searches for `&` or `\` bytes using `bytes.IndexByte`. If neither character exists, `b` is returned immediately without allocation or processing.
2. **Copy-on-Write Buffer Setup:**
   Instantiates `util.NewCopyOnWriteBuffer(b)`. Memory is only allocated when a sequence requiring decoding is found.
3. **Iterative Scanning:**
   Iterates through the byte slice:
   * **Backslash Handling (`\`):**
     If followed by a punctuation character (via `util.IsPunct`) or a space (if `EscapedSpace` is enabled):
     * Writes accumulated unescaped bytes to the buffer.
     * If the character following `\` is not a space, writes that character byte.
     * Advances the loop index to skip the escaped character.
   * **Entity Handling (`&`):**
     * **Hexadecimal (`&#x...;` / `&#X...;`):** Reads hex digits up to a semicolon `;`. If valid and under length limits (< 7 digits), parses the uint via `strconv.ParseUint`, converts it to a valid rune (`util.ToValidRune`), encodes it into UTF-8 bytes, and appends it to the buffer.
     * **Decimal (`&#...;`):** Reads numeric digits up to `;`. If valid and under length limits (< 8 digits), parses the base-10 value, converts to a valid rune, encodes to UTF-8, and appends to the buffer.
     * **Named Entity (`&name;`):** Reads alphanumeric character sequences up to `;`. Looks up the name using `util.LookUpHTML5EntityByName`. If found, appends the replacement bytes (`entity.Characters`) to the buffer.
4. **Buffer Finalization:**
   If the copy-on-write buffer was modified (`cob.IsCopied()`), any trailing unescaped bytes are appended, and the modified byte slice is returned. Otherwise, the original slice `b` is returned.

---

### `DefaultDecoder.DecodeTo(w io.Writer, b []byte) (int, error)`

1. **Fast Path Check:**
   If `b` contains neither `&` nor `\`, calls `w.Write(b)` directly and returns the result.
2. **Streaming Execution:**
   Follows the same scanning and decoding rules as `Decode()`, but instead of buffering to a byte slice, writes segments directly to `w` using `w.Write(...)`.
3. **Error Handling & Byte Tracking:**
   Tracks total bytes written in the integer `written`. If any call to `w.Write()` returns an error, decoding stops immediately and returns `written` alongside the encountered error.