# Technical Documentation: `extension/definition_list_test.go`

## Overview

The `extension/definition_list_test.go` file contains the unit test suite for verifying the functionality of the **Definition List** extension within the `goldmark` v2 library. 

Its primary purpose is to set up a Goldmark Markdown parser and HTML renderer configured with the definition list parser and renderer extensions, and then execute test cases defined in an external test file (`_test/definition_list.txt`).

---

## File Metadata

* **File Path:** `extension/definition_list_test.go`
* **Package:** `extension`

---

## Dependencies

The file relies on the standard Go testing library alongside specific packages from the Goldmark framework:

| Import Path | Description |
| :--- | :--- |
| `testing` | Standard Go package providing support for automated testing. |
| `github.com/yuin/goldmark/v2/parser` | Provides core parsing functionality and option setters. |
| `github.com/yuin/goldmark/v2/renderer/html` | Provides HTML rendering functionality and option setters. |
| `github.com/yuin/goldmark/v2/testutil` | Provides utility functions for running file-based Markdown test cases. |

---

## Functions

### `TestDefinitionList(t *testing.T)`

`TestDefinitionList` is the single test entry point in this file. It configures the Markdown engine with definition list capabilities and runs test cases against an expected output file.

#### Execution Flow

1. **Construct standard Markdown processing function (`markdown`)**:
   It calls `testutil.NewMarkdownToStringFunc` to create a Markdown-to-string transformation function. This function is configured with two main sub-components:
   * **Parser (`parser.New`)**: Configured with `parser.WithExtensions(NewDefinitionListParser())` to attach the definition list parsing extension.
   * **HTML Renderer (`html.New`)**: Configured with:
     * `html.WithUnsafe()`: Allows unsafe HTML content in the output.
     * `html.WithExtensions(NewDefinitionListHTMLRenderer())`: Attaches the definition list HTML renderer extension.

2. **Execute File Test Cases (`testutil.DoTestCaseFile`)**:
   It delegates test execution to `testutil.DoTestCaseFile`, passing:
   * `markdown`: The configured conversion function.
   * `"_test/definition_list.txt"`: Path to the file containing Markdown inputs and expected HTML outputs.
   * `t`: The active testing context (`*testing.T`).
   * `testutil.ParseCliCaseArg()...`: Command-line arguments parsed for filtering or controlling test case execution.

---

## Code Summary

```go
func TestDefinitionList(t *testing.T) {
	markdown := testutil.NewMarkdownToStringFunc(
		parser.New(
			parser.WithExtensions(NewDefinitionListParser()),
		),
		html.New(
			html.WithUnsafe(),
			html.WithExtensions(NewDefinitionListHTMLRenderer()),
		),
	)
	testutil.DoTestCaseFile(markdown, "_test/definition_list.txt", t, testutil.ParseCliCaseArg()...)
}
```

Key constructors called within the test setup:
* `NewDefinitionListParser()`: Instantiates the parser extension for definition lists.
* `NewDefinitionListHTMLRenderer()`: Instantiates the HTML renderer extension for definition lists.