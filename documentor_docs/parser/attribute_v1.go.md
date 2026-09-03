# Technical Documentation: `parser/attribute_v1.go`

## Overview

The `parser/attribute_v1.go` file provides parsing logic for Markdown block/inline attributes enclosed in curly braces (e.g., `{ #id .class key=value }`). It is built conditionally under the `goldmark_v1_attribute` build tag as part of the Goldmark AST parser package.

The package exposes functionality to parse attribute lists and individual attribute key-value pairs from a `text.Reader`, handling shorthand notation for IDs and classes as well as standard key-value attributes.

---

## Build Constraints & Package Imports

### Build Tag
```go
//go:build goldmark_v1_attribute
```
This file is only included in the build when the `goldmark_v1_attribute` build tag is active.

### Package & Imports
- **Package:** `parser`
- **Imports:**
  - `gast "github.com/yuin/goldmark/v2/ast"`: AST definitions for Goldmark node attributes.
  - `github.com/yuin/goldmark/v2/text`: Text reader and multi-line value manipulation utilities.
  - `github.com/yuin/goldmark/v2/util`: General byte and string manipulation helpers.

---

## Package Variables

- `attrNameID = []byte("id")`: Pre-allocated byte slice representing the attribute name `"id"`.
- `attrNameClass = []byte("class")`: Pre-allocated byte slice representing the attribute name `"class"`.

---

## Functions

### 1. `ParseAttributes`

```go
func ParseAttributes(reader text.Reader) ([]gast.Attribute, bool)
```

#### Description
Parses an attribute block enclosed in `{ ... }` from the provided `text.Reader`. If parsing fails at any point inside the block, it restores the reader's original position prior to parsing and returns `nil, false`.

#### Parameters
- `reader`: A `text.Reader` interface pointing to the source text.

#### Return Values
- `[]gast.Attribute`: A slice of successfully parsed attributes.
- `bool`: `true` if attributes were parsed successfully; `false` otherwise.

#### Workflow & Logic
1. **Save Initial State**: Records the starting line and position using `reader.Position()`.
2. **Opening Brace Validation**:
   - Skips leading spaces.
   - Checks if the next character is `{`.
   - If not `{`, restores position to `savedLine, savedPosition` and returns `nil, false`.
   - Advances reader past `{`.
3. **Attribute Parsing Loop**:
   - Peeks for closing brace `}`. If found, advances past `}` and returns `attrs, true`.
   - Calls `parseAttribute(reader)`.
   - If `parseAttribute` fails (`false`), restores original reader position and returns `nil, false`.
   - **Class Merging Logic**:
     - If the parsed attribute name is `"class"`, checks if a `"class"` attribute already exists in `attrs`.
     - If found: concatenates the existing class string and new class string with a space (`existing + " " + newVal`) using `text.NewMultiLineValueFromString(..., text.IdentityDecoder)`, replacing the existing class attribute value.
     - If not found: appends the new `class` attribute to `attrs`.
   - **Other Attributes**: Non-`class` attributes are appended directly to `attrs`.
   - **Delimiter Handling**: Skips trailing spaces. If a comma `,` is encountered, advances past it and skips subsequent spaces.

---

### 2. `parseAttribute`

```go
func parseAttribute(reader text.Reader) (gast.Attribute, bool)
```

#### Description
Parses an individual attribute item from the reader. Supports two syntaxes:
1. **Shorthand Syntax**: `#id` or `.class`
2. **Key-Value Syntax**: `name=value`

#### Parameters
- `reader`: A `text.Reader` holding the text slice to parse.

#### Return Values
- `gast.Attribute`: Struct containing `Name` (string) and `Value` (`text.MultiLineValue`).
- `bool`: `true` if a valid attribute was parsed; `false` otherwise.

---

## Parsing Mechanics

### A. Shorthand Syntax (`#id` / `.class`)

Triggered if the first non-space character is `#` or `.`.

1. **Indicator Check**: Advances past `#` or `.`.
2. **Name Resolution**:
   - `#` assigns attribute name `id`.
   - `.` assigns attribute name `class`.
3. **Character Rules**: Scans characters on the current line. A character is valid if:
   - It is not a whitespace character (`!util.IsSpace`).
   - AND either it is not punctuation (`!util.IsPunct`) OR it is one of the allowed characters: `_`, `-`, `:`, `.`.
4. **Reader Advance & Output**: Advances the reader by the length of the matched segment and constructs a `gast.Attribute` using `text.NewMultiLineValue`.

---

### B. Key-Value Syntax (`name=value`)

Triggered when the standard key-value format is detected.

1. **Attribute Name Validation**:
   - Inspects the current line. Returns `false` if line length is zero.
   - **First character rule**: Must be a lower/upper-case ASCII letter (`a-z`, `A-Z`), `_`, or `:`.
   - **Subsequent character rules**: Allowed characters include ASCII letters (`a-z`, `A-Z`), digits (`0-9`), `_`, `:`, `.`, or `-`.
2. **Equal Sign Check**:
   - Advances past the name.
   - Skips spaces and checks for `=`. If missing, returns `false`.
   - Advances past `=`.
3. **Attribute Value Parsing**:
   - Skips spaces after `=`.
   - Captures reader position before parsing (`pos1`).
   - Delegates value parsing to `text.ParseAttributeValue(reader)`. If this fails, returns `false`.
   - Captures reader position after parsing (`pos2`).
   - Extracts the raw attribute value using `reader.ValueBetween(pos1.Start, pos2.Start)`.
4. **Return**: Returns `gast.Attribute` containing string-converted name and parsed value segment.

---

## Key Processing Behaviors

| Feature | Behavior |
| :--- | :--- |
| **Backtracking on Failure** | If parsing fails anywhere within `ParseAttributes`, `reader.SetPosition` resets reader state back to before the opening `{`. |
| **Multiple Class Merging** | Multiple class definitions (e.g., `{ .foo .bar class="baz" }`) are merged into a single `class` attribute with space separation (`foo bar baz`). |
| **Comma Separation** | Commas between attributes are optional and consumed automatically if present (e.g., `{ #id, .class }`). |