# Technical Documentation: `text/reader_test.go`

## Overview

The `text/reader_test.go` file contains a unit test for the `text` package within the `goldmark` library (v2). Its primary purpose is to verify the behavior of the `text.Reader` type's `FindSubMatch` method when processing text containing non-ASCII Unicode characters (specifically CJK / Chinese characters).

---

## Package and Imports

- **Package Name**: `text_test`
- **Imports**:
  - `regexp`: Go standard library package used to compile and execute regular expressions.
  - `testing`: Go standard library package providing support for automated testing.
  - `github.com/yuin/goldmark/v2/text`: The internal library package being tested, specifically providing `text.NewReader`, `text.NewDecoder`, and the `text.Reader` interface/type.

---

## Unit Test Summary

### `TestFindSubMatchReader(t *testing.T)`

This function tests that `text.Reader.FindSubMatch` correctly extracts regex submatches when matching against Unicode letter characters enclosed in delimiters.

#### Execution Steps

1. **Input Preparation**:
   - Defines a target CJK string `s := "微笑"`.
   - Creates a slice of bytes formatting the input as `":微笑:"`.
   - Instantiates a new reader `r` using `text.NewReader([]byte(":"+s+":"), text.NewDecoder())`.

2. **Regex Compilation**:
   - Compiles a regular expression: `regexp.MustCompile(`:(\p{L}+):`)`.
   - The pattern `:(\p{L}+):` searches for a colon `:`, followed by one or more Unicode letters (`\p{L}+`) captured in a submatch group, followed by another colon `:`.

3. **Submatch Extraction**:
   - Executes `match := r.FindSubMatch(reg)` on the custom reader instance `r`.

4. **Assertion**:
   - Checks two conditions:
     1. The length of `match` must be `2` (where `match[0]` is the full match `":微笑:"` and `match[1]` is the captured group `"微笑"`).
     2. The string value of `match[1]` must equal `s` (`"微笑"`).
   - If either condition fails, the test halts with `t.Fatal("no match cjk")`.

---

## Key Components Tested

| Component / Method | Description |
| :--- | :--- |
| `text.NewReader([]byte, text.NewDecoder())` | Initializes a reader instance wrapping the provided byte slice using a decoder. |
| `r.FindSubMatch(reg)` | Evaluates the compiled regular expression against the content managed by the reader `r` and returns the matched slices/submatches. |