# Technical Documentation: `extension/typographer_test.go`

## Overview

The `extension/typographer_test.go` file defines unit tests for the **Typographer** extension in the Goldmark Markdown processor framework (v2). It uses Goldmark's internal `testutil` package to run data-driven test cases stored in an external file (`_test/typographer.txt`).

---

## Package and Dependencies

* **Package:** `extension`
* **Imports:**
  * `testing`: Standard Go library package for automated testing.
  * `github.com/yuin/goldmark/v2/parser`: Provides Goldmark parser implementation and configuration options.
  * `github.com/yuin/goldmark/v2/renderer/html`: Provides Goldmark HTML renderer implementation and configuration options.
  * `github.com/yuin/goldmark/v2/testutil`: Utility package for running data-driven test suites against Goldmark extensions.

---

## Functions

### `TestTypographer(t *testing.T)`

`TestTypographer` is the main entry point for running the test suite for the typographer extension using Go's `testing` package.

#### Internal Mechanics & Workflow

1. **Markdown Function Construction (`testutil.NewMarkdownToStringFunc`)**
   A helper function (`markdown`) is initialized to convert Markdown input text into rendered HTML output string based on specific parser and renderer options:
   * **Parser Configuration:** Created via `parser.New(...)` equipped with the `NewTypographerParser()` extension using `parser.WithExtensions()`.
   * **Renderer Configuration:** Created via `html.New(...)` with unsafe HTML rendering enabled using `html.WithUnsafe()`.

2. **Test Case Execution (`testutil.DoTestCaseFile`)**
   Executes the test suite by reading test definitions from a file:
   * **Markdown Function:** Passes the configured `markdown` conversion function.
   * **Test File Path:** Target file `_test/typographer.txt` containing input Markdown and expected output HTML test cases.
   * **Testing Context:** Passes standard Go testing handle `t`.
   * **CLI Filter Arguments:** Passes `testutil.ParseCliCaseArg()...` to allow command-line filtering or arguments when executing tests.