# Technical Documentation: `options_test.go`

## Overview

The `options_test.go` file contains a unit test function for the Goldmark Markdown processing library (`github.com/yuin/goldmark/v2`). Its primary purpose is to execute Markdown conversion tests that specifically evaluate the parser's behavior when both attribute support and automatic heading ID generation are enabled.

---

## Build Directives

```go
//go:build !goldmark_v1_attribute
```

* **`!goldmark_v1_attribute`**: This build tag ensures that the file is included in the build set only when the `goldmark_v1_attribute` build tag is **not** specified. 

---

## Package and Dependencies

* **Package**: `goldmark_test` (External test package for testing `goldmark` functionality).
* **Imports**:
  * `testing`: Standard Go package providing automated testing support.
  * `github.com/yuin/goldmark/v2/parser`: Provides the Markdown parser implementation and parser configuration options.
  * `github.com/yuin/goldmark/v2/renderer/html`: Provides the HTML renderer implementation.
  * `github.com/yuin/goldmark/v2/testutil`: Provides testing helper functions for running data-driven test cases against spec files.

---

## Functions

### `TestAttributeAndAutoHeadingID(t *testing.T)`

```go
func TestAttributeAndAutoHeadingID(t *testing.T)
```

#### Purpose
`TestAttributeAndAutoHeadingID` tests Markdown-to-HTML conversion when the parser is configured with both `WithAttribute()` and `WithAutoHeadingID()` options enabled.

#### Workflow & Key Components

1. **Markdown Function Initialization**:
   ```go
   markdown := testutil.NewMarkdownToStringFunc(
       parser.New(parser.WithAttribute(), parser.WithAutoHeadingID()),
       html.New(),
   )
   ```
   * Calls `testutil.NewMarkdownToStringFunc` to create a Markdown conversion function that returns an HTML string.
   * **Parser Configuration**: Instantiates a new parser via `parser.New(...)` with two specific options:
     * `parser.WithAttribute()`: Enables attribute parsing.
     * `parser.WithAutoHeadingID()`: Enables automatic heading ID generation.
   * **Renderer Configuration**: Instantiates an HTML renderer via `html.New()`.

2. **Test Execution**:
   ```go
   testutil.DoTestCaseFile(markdown, "_test/options.txt", t, testutil.ParseCliCaseArg()...)
   ```
   * Calls `testutil.DoTestCaseFile` to load and execute test cases.
   * **Arguments**:
     * `markdown`: The configured Markdown transformation function.
     * `"_test/options.txt"`: Relative file path containing the test case input Markdown and expected HTML output.
     * `t`: The current test runner state pointer (`*testing.T`).
     * `testutil.ParseCliCaseArg()...`: Parses command-line arguments to allow filtered execution of specific test cases defined in the test file.