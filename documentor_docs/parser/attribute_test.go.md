# Technical Documentation: `parser/attribute_test.go`

## Overview

The `parser/attribute_test.go` file contains unit tests for the attribute parsing functionality in the Goldmark v2 Markdown parser. Specifically, it tests the `parser.ParseAttributes` function under various syntax configurations, including unquoted values, entity references, quoted values, shorthand ID and class attributes, and multi-line values.

---

## Build Constraints and Package

- **Build Tag**: `//go:build !goldmark_v1_attribute`  
  This file is only compiled when the `goldmark_v1_attribute` build tag is **not** defined, ensuring these tests execute against Goldmark v2 attribute handling behavior.
- **Package**: `parser_test`  
  An external test package context (`parser_test`) used to test the public API of the `parser` package.

---

## Dependencies

- **`testing`**: Standard Go testing library.
- **`github.com/yuin/goldmark/v2/parser`**: Provides the `ParseAttributes` function and attribute data types.
- **`github.com/yuin/goldmark/v2/text`**: Provides `NewReader` and `NewDecoder` to construct the text reader required for parsing.

---

## Test Functions

### 1. `TestUnquotedAttribute(t *testing.T)`

Tests the behavior of parsing unquoted attribute values, specifically checking how entity references are parsed and evaluated versus plain unquoted values.

#### Test Cases Executed:
1. **Unquoted Attribute with Entity Reference (`{key=value&quot;}`)**:
   - **Parsing**: Instantiates `text.NewReader(source, text.NewDecoder())` and passes it to `parser.ParseAttributes`.
   - **Assertions**:
     - Parsing succeeds (`ok == true`).
     - Exactly `1` attribute is returned.
     - Attribute `Name` is `"key"`.
     - `attrs[0].Value.Value(source)` evaluates entity references and returns `"value\""`.
     - `attrs[0].Value.Str(source)` returns the raw unescaped string `"value&quot;"`.
     - `attrs[0].Value.IsOwned()` returns `false`.

2. **Plain Unquoted Attribute (`{key=value}`)**:
   - **Parsing**: Re-executes `parser.ParseAttributes` on source without entity references.
   - **Assertions**:
     - Parsing succeeds (`ok == true`).
     - Exactly `1` attribute is returned.
     - Attribute `Name` is `"key"`.
     - `attrs[0].Value.Str(source)` returns `"value"`.
     - `attrs[0].Value.IsOwned()` returns `false`.

---

### 2. `TestQuotedAttribute(t *testing.T)`

Tests parsing standard key-value attributes wrapped in double quotes.

#### Test Execution:
- **Source Input**: `{key="value"}`
- **Assertions**:
  - Parsing succeeds (`ok == true`).
  - Exactly `1` attribute is returned.
  - Attribute `Name` is `"key"`.
  - `attrs[0].Value.Str(source)` returns `"value"`.
  - `attrs[0].Value.IsOwned()` returns `false`.

---

### 3. `TestIdAndClassAttributes(t *testing.T)`

Tests shorthand syntax for parsing HTML IDs (`#`) and CSS classes (`.`), including multiple CSS class attributes.

#### Test Execution:
- **Source Input**: `{#id .class1 .class2}`
- **Assertions**:
  - Parsing succeeds (`ok == true`).
  - Exactly `2` attributes are returned (`id` and combined `class`).
  - **First Attribute (`attrs[0]`)**:
    - `Name` is `"id"`.
    - `Value.Str(source)` is `"id"`.
  - **Second Attribute (`attrs[1]`)**:
    - `Name` is `"class"`.
    - `Value.Str(source)` concatenates space-separated classes to `"class1 class2"`.

---

### 4. `TestMultiLineAttributes(t *testing.T)`

Tests attribute values that extend across multiple lines within double quotes.

#### Test Execution:
- **Source Input**: `{value="aaa\n   bbb"}`
- **Assertions**:
  - Parsing succeeds (`ok == true`).
  - Exactly `1` attribute is returned.
  - Attribute `Name` is `"value"`.
  - `attrs[0].Value.Str(source)` preserves newlines and indentation, returning `"aaa\n   bbb"`.
  - `attrs[0].Value.IsOwned()` returns `false`.

---

## Summary of Observed API Interactions

Based directly on the source code, the attribute parser uses the following mechanisms:

| Method / Property | Description |
| :--- | :--- |
| `parser.ParseAttributes(reader)` | Parses attribute blocks enclosed in `{}` from a `text.Reader`. Returns `([]parser.Attribute, bool)`. |
| `text.NewReader(source, decoder)` | Creates a new reader from a byte slice source and text decoder. |
| `attr.Name` | Holds the parsed string name of the attribute (e.g., `"key"`, `"id"`, `"class"`). |
| `attr.Value.Str(source)` | Extracts the raw string representation of the attribute value using the source byte array. |
| `attr.Value.Value(source)` | Extracts the decoded string representation of the attribute value (e.g., resolving HTML entities like `&quot;` to `"`). |
| `attr.Value.IsOwned()` | Indicates whether the attribute value memory is owned by the structure. |