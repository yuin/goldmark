# Technical Documentation: `testutil/testutil.go`

The `testutil` package provides specialized utility structures and functions for writing and executing unit tests for Goldmark Markdown parsing and rendering components.

---

## Overview

The `testutil` package automates the process of reading structured test-case files, executing Markdown-to-HTML conversion functions, comparing actual output against expected output, and formatting diagnostic diffs upon test failure.

---

## Interfaces and Core Types

### `TestingT`
An interface representing a minimal subset of Go's standard `*testing.T` struct.
```go
type TestingT interface {
    Logf(string, ...any)
    Skipf(string, ...any)
    Errorf(string, ...any)
    FailNow()
}
```
* **Purpose**: Allows functions in `testutil` to accept standard `testing.T` pointers or custom test runners implementing these methods.

---

### `MarkdownToStringFunc` & Configuration Types

#### `MarkdownToStringFunc`
```go
type MarkdownToStringFunc func(source string, opts ...MarkdownToStringFuncOption) (string, error)
```
A function signature representing a Markdown rendering operation that accepts a source string and optional configuration options, returning rendered HTML string output.

#### `MarkdownToStringFuncOption`
```go
type MarkdownToStringFuncOption func(*markdownToStringFuncConfig)
```
A functional option pattern used to pass parser (`parser.ParseOption`) or renderer (`renderer.RenderOption`) options when invoking a `MarkdownToStringFunc`.

#### `markdownToStringFuncConfig`
Internal configuration struct containing option slices:
* `parseOptions`: `[]parser.ParseOption`
* `renderOptions`: `[]renderer.RenderOption`

---

### `MarkdownTestCase` & `MarkdownTestCaseOptions`

#### `MarkdownTestCase`
Holds data for an individual Markdown test case.
```go
type MarkdownTestCase struct {
    No          int
    Description string
    Options     MarkdownTestCaseOptions
    Markdown    string
    Expected    string
}
```

#### `MarkdownTestCaseOptions`
Options defining preprocessing transformations applied to the test input and expected output.
```go
type MarkdownTestCaseOptions struct {
    EnableEscape bool
    Trim         bool
}
```
* `EnableEscape`: When `true`, interprets byte escape sequences (e.g., `\n`, `\t`, `\xHH`, `\uXXXX`) in both input Markdown and expected output.
* `Trim`: When `true`, strips leading and trailing whitespace from both input Markdown and expected output.

---

## File Format Specification

The `DoTestCaseFile` function parses test cases from plain text files structured with specific section delimiters and syntax rules:

### Delimiters
* **Attribute Separator**: `//- - - - - - - - -//`
* **Case Separator**: `//= = = = = = = = = = = = = = = = = = = = = = = =//`

### Structure
```text
<Case Number>[: <Description>]
[options:{"EnableEscape": true, "Trim": true}]
//- - - - - - - - -//
<Input Markdown Content>
//- - - - - - - - -//
<Expected Rendered Content>
//= = = = = = = = = = = = = = = = = = = = = = = =//
```

1. **Header Line**: Starts with a case number integer (e.g., `1`), optionally followed by a colon `:` and a description string. Blank lines in the header area are ignored.
2. **Options Line** *(Optional)*: Matches the regular expression `(?i)\s*options:(.*)`. Contains a JSON string mapping to `MarkdownTestCaseOptions`.
3. **Attribute Separator**: `//- - - - - - - - -//`
4. **Input Markdown**: Text lines representing source Markdown.
5. **Attribute Separator**: `//- - - - - - - - -//`
6. **Expected Output**: Text lines representing expected HTML output.
7. **Case Separator**: `//= = = = = = = = = = = = = = = = = = = = = = = =//`

---

## Functions Reference

### Test Setup and Execution Functions

#### `NewMarkdownToStringFunc`
```go
func NewMarkdownToStringFunc(p parser.Parser, r renderer.Renderer[io.Writer]) MarkdownToStringFunc
```
Creates a `MarkdownToStringFunc` wrapping a specific Goldmark `parser.Parser` and `renderer.Renderer[io.Writer]`. 
* Converts string input to byte slice using `util.StringToReadOnlyBytes`.
* Executes `p.Parse` and `r.Render`, capturing rendered output in a buffer.

#### `ParseCliCaseArg`
```go
func ParseCliCaseArg() []int
```
Inspects `os.Args` for arguments starting with `case=`. Parses comma-separated integers (e.g., `case=1,3,5`) into a slice of test case numbers (`[]int`). Used to selectively run specific test cases from the command line.

#### `DoTestCaseFile`
```go
func DoTestCaseFile(m MarkdownToStringFunc, filename string, t TestingT, no ...int)
```
Opens and scans the specified test case file, constructs `MarkdownTestCase` instances, filters them based on `no` (or command line arguments mapped through `no`), and passes them to `DoTestCases`. Panics if file IO or parsing format errors occur.

#### `DoTestCases`
```go
func DoTestCases(m MarkdownToStringFunc, cases []MarkdownTestCase, t TestingT)
```
Iterates through a slice of `MarkdownTestCase` objects and executes each case using `DoTestCase`.

#### `DoTestCase`
```go
func DoTestCase(m MarkdownToStringFunc, testCase MarkdownTestCase, t TestingT)
```
Executes a single test case:
1. Applies trimming and escape conversions using helper functions `source()` and `expected()`.
2. Invokes the `MarkdownToStringFunc`.
3. Compares actual rendered output with expected output using `bytes.Equal(bytes.TrimSpace(...))`.
4. **Panic/Failure Recovery**: Utilizes `defer` and `recover()` to catch runtime panics or assertion failures. Reports details back to `t.Errorf` with the case number, description, raw Markdown input, expected output, actual output, stack traces (on panic), or visual diffs generated via `DiffPretty`.

---

### Diffing Utilities

#### `DiffPretty`
```go
func DiffPretty[T1 []byte | string, T2 []byte | string](v1 T1, v2 T2) []byte
```
Generates a human-readable visual difference between two inputs (`v1` as expected, `v2` as actual).
* Uses `simpleDiff` to calculate line changes.
* Formats added lines with `+ | `, removed lines with `- | `, and unchanged lines with `  | `.
* Modified/added/removed lines display space visualizer characters via `util.VisualizeSpaces`.

#### `simpleDiff` & `simpleDiffAux`
```go
func simpleDiff(v1, v2 []byte) []diff
func simpleDiffAux(v1lines, v2lines [][]byte) []diff
```
Internal line-based diff algorithms that compute recursive block matches between two byte slices (`v1` and `v2`) and categorize diff sections into `diffRemoved`, `diffAdded`, or `diffNone`.

---

### Helper Functions

#### `applyEscapeSequence`
```go
func applyEscapeSequence(b []byte) []byte
```
Parses escape sequences in a byte slice and converts them into literal characters:
* Standard escape characters: `\a`, `\b`, `\f`, `\n`, `\r`, `\t`, `\v`, `\\`
* Hexadecimal byte values: `\xHH` (2 hex digits)
* Unicode code points: `\uXXXX` (4 hex digits) or `\UXXXXXXXX` (8 hex digits)

#### `source` & `expected`
```go
func source(t *MarkdownTestCase) string
func expected(t *MarkdownTestCase) string
```
Internal helper functions that inspect a `MarkdownTestCase`'s `Options` flags (`Trim`, `EnableEscape`) and return modified string representations of `Markdown` and `Expected` attributes respectively.