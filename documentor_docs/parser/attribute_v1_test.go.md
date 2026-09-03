# Technical Documentation: `parser/attribute_v1_test.go`

## Overview

The `parser/attribute_v1_test.go` file contains unit tests for verifying the attribute parsing functionality within the `goldmark` library, specifically targeting the V1 attribute syntax. It tests the behavior of `parser.ParseAttributes` when processing complex attribute structures such as arrays, key-value pairs, strings with spaces, and nested maps/objects.

---

## Build Constraints & Package Declaration

* **Build Tag**: `//go:build goldmark_v1_attribute`  
  This file is compiled only when the `goldmark_v1_attribute` build tag is specified during test execution (`go test -tags goldmark_v1_attribute`).
* **Package**: `parser_test`  
  The test is defined in an external test package (`parser_test`), meaning it tests the public API exported by `github.com/yuin/goldmark/v2/parser` and `github.com/yuin/goldmark/v2/text`.

---

## Imports

* **`bytes`**: Used to perform byte slice equality checks via `bytes.Equal`.
* **`testing`**: Provides Go standard unit testing infrastructure (`*testing.T`).
* **`github.com/yuin/goldmark/v2/parser`**: Package under test; provides `parser.ParseAttributes`.
* **`github.com/yuin/goldmark/v2/text`**: Provides text reading primitives (`text.NewReader`, `text.NewDecoder`).

---

## Test Function: `TestAttributeV1(t *testing.T)`

`TestAttributeV1` runs three consecutive parsing test cases sequentially against different raw attribute byte strings.

### Test Workflow Summary

In each test case:
1. A raw byte slice source (`[]byte`) containing attribute syntax is defined.
2. A `text.Reader` instance is constructed using `text.NewReader(source, text.NewDecoder())`.
3. `parser.ParseAttributes` is called with the reader.
4. The test verifies `ok == true` and checks the number of parsed attributes (`len(attrs)`).
5. Attribute values are evaluated using `attrs[i].Value.Any(source)` to convert parsed values into Go types.
6. Strict type assertions and value equality checks are performed.

---

## Test Cases Breakdown

### Test Case 1: Array Attribute Values
* **Source**: `[]byte("{key=[1,2,3]}")`
* **Expected Parsed Structure**:
  * **Attribute Count**: 1
  * **Attribute Name**: `"key"`
  * **Value Type**: `[]any`
  * **Value Elements**: `[float64(1), float64(2), float64(3)]` (length: 3)
* **Assertions**:
  * Asserts `ParseAttributes` returns `ok == true`.
  * Verifies attribute count is `1`.
  * Verifies attribute name is `"key"`.
  * Converts attribute value with `Value.Any(source)` and type-asserts to `[]any`.
  * Verifies length is `3` and elements equal `float64(1)`, `float64(2)`, and `float64(3)`.

---

### Test Case 2: Primitive Types and Strings with Spaces
* **Source**: `[]byte("{key=1, key2=\"value with spaces\"}")`
* **Expected Parsed Structure**:
  * **Attribute Count**: 2
  * **First Attribute**:
    * **Name**: `"key"`
    * **Value Type**: `float64`
    * **Value**: `1.0`
  * **Second Attribute**:
    * **Name**: `"key2"`
    * **Value Type**: `[]byte`
    * **Value**: `[]byte("value with spaces")`
* **Assertions**:
  * Asserts `ParseAttributes` returns `ok == true`.
  * Verifies attribute count is `2`.
  * Checks names `"key"` and `"key2"` in index order `0` and `1`.
  * Verifies first attribute converts to `float64` equal to `1`.
  * Verifies second attribute converts to `[]byte` matching `"value with spaces"` using `bytes.Equal`.

---

### Test Case 3: Nested Map Structures
* **Source**: `[]byte("{key={\"nested\":[1,2,3], \"another\"=true}}")`
* **Expected Parsed Structure**:
  * **Attribute Count**: 1
  * **Attribute Name**: `"key"`
  * **Value Type**: `map[string]any`
  * **Map Entries**:
    * Key `"nested"`: `[]any` containing `[float64(1), float64(2), float64(3)]`
    * Key `"another"`: `bool` (`true`)
* **Assertions**:
  * Asserts `ParseAttributes` returns `ok == true`.
  * Verifies attribute count is `1`.
  * Verifies attribute name is `"key"`.
  * Converts attribute value to `map[string]any` and verifies map length is `2`.
  * Type-asserts `m["nested"]` to `[]any` and verifies elements `[1, 2, 3]` as `float64`.
  * Verifies `m["another"]` equals `true`.

---

## Data Type Conversion Reference

The following table summarizes how attribute values parsed from source text map to Go runtime types when `attrs[i].Value.Any(source)` is invoked:

| Syntactic Structure in Source | Evaluated Type via `.Any(source)` |
| :--- | :--- |
| Numbers (e.g., `1`) | `float64` |
| Quoted Strings (e.g., `"value with spaces"`) | `[]byte` |
| Lists/Arrays (e.g., `[1,2,3]`) | `[]any` |
| Objects/Maps (e.g., `{"nested": ...}`) | `map[string]any` |
| Booleans (e.g., `true`) | `bool` |