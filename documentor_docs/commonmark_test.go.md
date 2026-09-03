# Technical Documentation: `commonmark_test.go`

## Overview

The `commonmark_test.go` file contains compliance unit tests for the Goldmark Markdown engine. It reads the standard CommonMark specification test suite from a local JSON file (`_test/spec.json`), filters and parses individual test cases, configures a Goldmark processor with XHTML and unsafe HTML options, and runs the test cases using Goldmark's test utility framework.

---

## Package and Dependencies

**Package:** `goldmark_test`

### Imported Packages

* **`encoding/json`**: Used to unmarshal the JSON representation of the CommonMark spec test cases into Go structs.
* **`os`**: Used to read the spec JSON file (`_test/spec.json`) from the file system.
* **`slices`**: Used for `slices.Contains()` to determine if specific test numbers requested via command-line arguments should be executed.
* **`testing`**: Provides Go's standard testing framework primitives (`*testing.T`).
* **`github.com/yuin/goldmark/v2/parser`**: Provides the base Markdown parser (`parser.New()`).
* **`github.com/yuin/goldmark/v2/renderer/html`**: Provides the HTML renderer implementation and configuration functions (`html.New`, `html.WithXHTML`, `html.WithUnsafe`).
* **`github.com/yuin/goldmark/v2/testutil`**: Provides helper functions and structures for executing Goldmark Markdown test suites (`MarkdownTestCase`, `ParseCliCaseArg`, `NewMarkdownToStringFunc`, `DoTestCases`).

---

## Data Structures

### `commonmarkSpecTestCase`

`commonmarkSpecTestCase` defines the structure matching the JSON objects present in the `_test/spec.json` file.

```go
type commonmarkSpecTestCase struct {
	Markdown  string `json:"markdown"`
	HTML      string `json:"html"`
	Example   int    `json:"example"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
	Section   string `json:"section"`
}
```

#### Fields

| Field | Type | JSON Tag | Description |
| :--- | :--- | :--- | :--- |
| `Markdown` | `string` | `markdown` | The input Markdown text for the test case. |
| `HTML` | `string` | `html` | The expected output HTML string. |
| `Example` | `int` | `example` | The numeric identifier/example number of the spec test case. |
| `StartLine` | `int` | `start_line` | The starting line number of the example within the original specification document. |
| `EndLine` | `int` | `end_line` | The ending line number of the example within the original specification document. |
| `Section` | `string` | `section` | The section name of the CommonMark specification to which the test case belongs. |

---

## Functions

### `TestSpec(t *testing.T)`

`TestSpec` is the main Go unit test function. It executes the CommonMark specification test suite against Goldmark.

#### Execution Flow Step-by-Step

1. **Load Test File**: Reads the raw bytes of `_test/spec.json` using `os.ReadFile`. If an error occurs, it causes a panic.
2. **Unmarshal JSON**: Parses the JSON bytes into a slice of `commonmarkSpecTestCase` structs (`testCases`). If JSON parsing fails, it causes a panic.
3. **Parse CLI Case Arguments**: Calls `testutil.ParseCliCaseArg()` to retrieve a slice of specific test example numbers (`nos`) requested by the user.
4. **Filter and Map Test Cases**:
   * Iterates through each `commonmarkSpecTestCase` in `testCases`.
   * Evaluates `shouldAdd`:
     * If `nos` is empty, all cases are selected (`shouldAdd = true`).
     * If `nos` is not empty, it checks if `c.Example` is present in `nos` using `slices.Contains`.
   * If `shouldAdd` is `true`, converts the case into a `testutil.MarkdownTestCase` structure containing:
     * `No`: Set to `c.Example`
     * `Markdown`: Set to `c.Markdown`
     * `Expected`: Set to `c.HTML`
   * Appends the mapped case to the `cases` slice.
5. **Initialize Converter Function**:
   * Constructs a Markdown processing function using `testutil.NewMarkdownToStringFunc`.
   * Configures the Markdown parser using `parser.New()`.
   * Configures the HTML renderer using `html.New(...)` supplied with two options:
     * `html.WithXHTML()`: Configures the renderer to output XHTML-compliant tags.
     * `html.WithUnsafe()`: Configures the renderer to allow raw HTML and potentially unsafe elements embedded in the input Markdown.
6. **Execute Test Cases**: Calls `testutil.DoTestCases(markdown, cases, t)` to run the constructed processing function against the filtered test cases and report results through `*testing.T`.