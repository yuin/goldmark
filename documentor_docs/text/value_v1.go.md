# Technical Documentation: `text/value_v1.go`

## Overview

The `text/value_v1.go` file provides functionality for parsing attribute values from text streams. It implements a JSON-like parser capable of parsing complex nested attributes, including maps, arrays, double-quoted strings, numeric values, booleans, null values, and custom identifier strings.

This file is conditionally compiled under the `goldmark_v1_attribute` build tag.

---

## Build Constraints & Dependencies

### Build Tag
```go
//go:build goldmark_v1_attribute
```
This file is only included in the build when the `goldmark_v1_attribute` build tag is active.

### Dependencies
* **Standard Library**: `bytes`, `io`, `strconv`
* **External Package**: `github.com/yuin/goldmark/v2/util`

---

## Exported API

### `(*MultiLineValue) Any`

```go
func (v MultiLineValue) Any(source []byte) any
```

#### Description
Converts a `MultiLineValue` slice reference from `source` into a decoded Go data structure (`any`). It extracts the underlying bytes from `source`, creates a `Reader` backed by an `IdentityDecoder`, and attempts to parse an attribute value.

#### Parameters
* `source []byte`: The source byte slice containing the text data referenced by `MultiLineValue`.

#### Returns
* `any`: The parsed Go value (such as `map[string]any`, `[]any`, `[]byte`, `float64`, `bool`, or `nil`). Returns `nil` if parsing fails.

---

### `ParseAttributeValue`

```go
func ParseAttributeValue(reader Reader) (any, bool)
```

#### Description
The main entry point for attribute value parsing. It skips leading whitespace from the provided `Reader`, evaluates the next character, and routes the parsing execution to the appropriate specific parser function.

#### Parameters
* `reader Reader`: An instance of `Reader` from which characters are consumed.

#### Returns
* `any`: The parsed value if successful, or `nil` on failure.
* `bool`: `true` if an attribute value was successfully parsed; `false` otherwise.

#### Parsing Logic Flow
1. Skips leading whitespace (`reader.SkipSpaces()`).
2. Checks the next byte (`reader.Peek()`):
   * `EOF`: Returns `nil, false`.
   * `{`: Calls `parseAttributeMap`.
   * `[`: Calls `parseAttributeArray`.
   * `"`: Calls `parseAttributeString`.
   * `-`, `+`, or numeric digit (via `util.IsNumeric`): Calls `parseAttributeNumber`.
   * Default: Calls `parseAttributeOthers`.

---

## Internal Utility & Parser Functions

### `parseAttributeMap`

```go
func parseAttributeMap(reder Reader) (map[string]any, bool)
```

Parses object syntax starting with `{` and ending with `}` into a `map[string]any`.

* **Syntax Rules**:
  * Key must be a double-quoted string (parsed via `parseAttributeString`).
  * Key-value delimiter can be either `:` or `=`.
  * Key-value pairs are separated by commas (`,`).
  * Trailing commas before closing `}` are invalid and cause parsing to fail.
* **Returns**: `map[string]any, true` on success, or `nil, false` on failure.

---

### `parseAttributeArray`

```go
func parseAttributeArray(reader Reader) ([]any, bool)
```

Parses sequence syntax starting with `[` and ending with `]` into a `[]any` slice.

* **Syntax Rules**:
  * Elements can be any valid attribute value parsed recursively via `ParseAttributeValue`.
  * Elements are separated by commas (`,`).
  * Trailing commas before closing `]` are invalid and cause parsing to fail.
* **Returns**: `[]any, true` on success, or `nil, false` on failure.

---

### `parseAttributeString`

```go
func parseAttributeString(reader Reader) ([]byte, bool)
```

Parses double-quoted string literals (`"..."`).

* **Supported Escape Sequences**:
  * `\"` $\rightarrow$ `"`
  * `\/` $\rightarrow$ `/`
  * `\\` $\rightarrow$ `\`
  * `\b` $\rightarrow$ Backspace (`\b`)
  * `\f` $\rightarrow$ Form feed (`\f`)
  * `\n` $\rightarrow$ Linefeed (`\n`)
  * `\r` $\rightarrow$ Carriage return (`\r`)
  * `\t` $\rightarrow$ Horizontal tab (`\t`)
  * Unrecognized escape sequences keep the literal `\` character followed by the escaped byte.
* **Returns**: `[]byte, true` on success (excluding outer quotes), or `nil, false` if string is unclosed.

---

### `parseAttributeNumber`

```go
func parseAttributeNumber(reader Reader) (float64, bool)
```

Parses numeric literals into a Go `float64`.

* **Supported Formats**:
  * Optional leading sign (`+` or `-`).
  * Integer part (scanned via `scanAttributeDecimal`).
  * Optional fractional part starting with `.` followed by digits.
  * Optional scientific notation exponent part (`e` or `E`) followed by an optional sign (`+` or `-`) and digits.
* Uses `strconv.ParseFloat` internally to parse the construct string.
* **Returns**: `float64, true` on success, or `0, false` on failure.

---

### `scanAttributeDecimal`

```go
func scanAttributeDecimal(reader Reader, w io.ByteWriter)
```

Helper function used by `parseAttributeNumber`. Sequentially consumes contiguous numeric digits (`util.IsNumeric`) from `reader` and writes them byte-by-byte into the provided `io.ByteWriter`.

---

### `parseAttributeOthers`

```go
func parseAttributeOthers(reader Reader) (any, bool)
```

Parses unquoted identifiers, booleans, and `null` values.

* **Identifier Validation**:
  * **First character**: Must be ASCII letter (`a-z`, `A-Z`), underscore (`_`), or colon (`:`).
  * **Subsequent characters**: ASCII letters (`a-z`, `A-Z`), digits (`0-9`), underscore (`_`), colon (`:`), period (`.`), or hyphen (`-`).
* **Reserved Literal Conversions**:
  * `"true"` $\rightarrow$ Go `bool` (`true`)
  * `"false"` $\rightarrow$ Go `bool` (`false`)
  * `"null"` $\rightarrow$ Go `nil`
  * Any other matching identifier string $\rightarrow$ Go `[]byte` containing the matched identifier.

---

## Summary of Type Mapping

| Parsed Syntax / Literal | Output Go Type |
| :--- | :--- |
| `{ "key": value }` / `{ "key" = value }` | `map[string]any` |
| `[ elem1, elem2 ]` | `[]any` |
| `"text"` | `[]byte` |
| `123`, `-45.67`, `1e-3` | `float64` |
| `true` / `false` | `bool` |
| `null` | `nil` |
| `unquoted_identifier` | `[]byte` |