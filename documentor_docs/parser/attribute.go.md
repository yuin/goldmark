# Technical Documentation Guide: `parser/attribute.go`

## Overview

The `parser/attribute.go` file provides parsing logic for HTML-style and Markdown syntax attribute blocks enclosed in curly braces (e.g., `{ #my-id .my-class key="value" }`) within the Goldmark Markdown processor (v2). 

It exports `ParseAttributes`, which reads attribute definitions from a `text.Reader`, constructs AST attribute structures (`gast.Attribute`), handles special syntax like class/ID shorthands (`.class`, `#id`), and manages class attribute merging.

### Build Tags
- `//go:build !goldmark_v1_attribute`: This file is compiled when the `goldmark_v1_attribute` build tag is **not** present.

---

## Package and Dependencies

```go
package parser

import (
    gast "github.com/yuin/goldmark/v2/ast"
    "github.com/yuin/goldmark/v2/text"
    "github.com/yuin/goldmark/v2/util"
)
```

- **`gast` (`github.com/yuin/goldmark/v2/ast`)**: Provides the AST representations, specifically `gast.Attribute`.
- **`text` (`github.com/yuin/goldmark/v2/text`)**: Handles text reading, position tracking, segments, indices, and multiline value generation (`text.Reader`, `text.MultiLineValue`, `text.ValueBuilder`).
- **`util` (`github.com/yuin/goldmark/v2/util`)**: Provides utility functions like byte conversions (`util.BytesToReadOnlyString`) and character checks (`util.IsSpace`, `util.IsPunct`).

---

## Key Functions

### 1. `ParseAttributes`

```go
func ParseAttributes(reader text.Reader) ([]gast.Attribute, bool)
```

#### Purpose
Main entry point for parsing an entire attribute block enclosed within `{` and `}`.

#### Behavior
1. **Position Snapshot**: Saves the initial reader state (`savedLine, savedPosition`) to allow position restoration (backtracking) if parsing fails.
2. **Block Start Check**: Skips leading spaces and checks if the next character is `{`. If not, restores position and returns `nil, false`.
3. **Loop Attributes**:
   - Skips spaces.
   - If `}` is encountered, advances reader past `}` and returns the accumulated slice `[]gast.Attribute` and `true`.
   - If `text.EOF` is encountered before closing `}`, returns `nil, false`.
   - Invokes `parseAttribute(reader)`. If it fails (`ok == false`), restores reader to the saved initial position and returns `nil, false`.
4. **Class Attribute Merging**:
   - If a parsed attribute has the name `"class"`, `ParseAttributes` checks if a `"class"` attribute already exists in the slice.
   - **If found**: Concatenates the existing class value and the new class value separated by a space (`existing + " " + newVal`), updating the existing attribute using `text.NewMultiLineValueFromString`.
   - **If not found**: Appends the new `"class"` attribute.
5. **Other Attributes**: Any non-class attribute is appended directly to the `attrs` slice.

---

### 2. `parseAttribute`

```go
func parseAttribute(reader text.Reader) (gast.Attribute, bool)
```

#### Purpose
Parses a single attribute within the block. Supports shorthands (`#id`, `.class`) and standard key/value attribute syntax (`name`, `name=value`).

#### Behavior
1. **Shorthand Processing (`#` or `.`)**:
   - If the current character is `#` or `.`:
     - Advances past the character.
     - Sets the attribute name to `"id"` (for `#`) or `"class"` (for `.`).
     - Scans characters until a space or a non-allowed punctuation character is met. Allowed punctuation inside shorthand names includes `_`, `-`, `:`, and `.`.
     - Returns `gast.Attribute` with the resolved name and value index range if valid (`i > 0`).
2. **Standard Name Processing**:
   - Reads the current line (`reader.PeekLine()`).
   - Rejects line if empty or if starting character is a space, `=`, `/`, or `}`.
   - Scans until encountering a space, `=`, `/`, or `}` to determine the attribute name length.
   - Advances reader past the attribute name and skips trailing spaces.
3. **Value Assignment Check**:
   - Checks next character:
     - **If NOT `=`**: Treats it as a boolean/valueless attribute. Returns `gast.Attribute` where the value is the slice of text matching the attribute name.
     - **If `=`**: Advances past `=`, skips spaces, and calls `parseAttributeValue(reader)`. Returns `gast.Attribute` with the parsed name and value.

---

### 3. `parseAttributeValue`

```go
func parseAttributeValue(reader text.Reader) (text.MultiLineValue, bool)
```

#### Purpose
Determines how an attribute value is formatted (quoted vs. unquoted) and routes processing accordingly.

#### Behavior
1. Skips leading spaces.
2. Peeks at the next character:
   - `text.EOF`: Returns empty `text.MultiLineValue{}` and `false`.
   - `"` (double quote): Delegates to `parseAttributeQuoted(reader, '"')`.
   - `'` (single quote): Delegates to `parseAttributeQuoted(reader, '\'')`.
   - **Default**: Delegates to `parseAttributeUnquoted(reader)`.

---

### 4. `parseAttributeQuoted`

```go
func parseAttributeQuoted(reader text.Reader, q byte) (text.MultiLineValue, bool)
```

#### Purpose
Parses single- or double-quoted attribute values, which may span multiple lines.

#### Parameters
- `reader`: The `text.Reader` instance.
- `q`: The quote byte delimiter (`'` or `"`).

#### Behavior
1. Advances past the initial quote character.
2. Iterates line-by-line using `reader.PeekLine()` searching for the matching closing quote `q`.
3. Records segment ranges into `lines` slice.
4. When matching quote `q` is found:
   - Advances reader past the closing quote.
   - Builds a `text.MultiLineValue` using `text.ValueBuilder` containing all collected line segments.
   - Returns the value and `true`.
5. If EOF/empty line is reached without finding a closing quote, returns empty `text.MultiLineValue{}` and `false`.

---

### 5. `parseAttributeUnquoted`

```go
func parseAttributeUnquoted(reader text.Reader) (text.MultiLineValue, bool)
```

#### Purpose
Parses an unquoted attribute value.

#### Behavior
1. Peeks at the current line.
2. Reads characters until reaching a space or a `}` character.
3. Advances the reader by the consumed character length `i`.
4. Constructs and returns a `text.MultiLineValue` from the segment index range and `true`.

---

## Processing Summary

| Input Pattern | Example | Parsed Attribute Name | Parsed Attribute Value | Special Handling |
| :--- | :--- | :--- | :--- | :--- |
| `#id` Shorthand | `{#main}` | `"id"` | `"main"` | Character `#` sets name to `"id"` |
| `.class` Shorthand | `{.btn}` | `"class"` | `"btn"` | Character `.` sets name to `"class"` |
| Valueless/Boolean | `{disabled}` | `"disabled"` | `"disabled"` | Value segment mirrors name segment |
| Standard Key/Value | `{lang=en}` | `"lang"` | `"en"` | Unquoted value stopped by space or `}` |
| Double Quoted | `{title="Hello"}` | `"title"` | `"Hello"` | Quoted strings can span multiple lines |
| Single Quoted | `{title='Hello'}` | `"title"` | `"Hello"` | Quoted strings can span multiple lines |
| Multiple Classes | `{.c1 .c2}` | `"class"` | `"c1 c2"` | Merged into a single `"class"` attribute space-separated |