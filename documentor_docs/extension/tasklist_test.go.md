# Technical Documentation: `extension/tasklist_test.go`

## Overview

The `extension/tasklist_test.go` file contains the unit test suite for the task list extension within the `extension` package. It utilizes the standard Go `testing` package alongside utilities from `github.com/yuin/goldmark/v2` to verify that task list items in Markdown are correctly parsed and rendered to HTML.

---

## Dependencies & Imports

The file relies on standard Go testing constructs and Goldmark v2 library components:

* **`testing`**: Standard Go package for writing automated tests.
* **`github.com/yuin/goldmark/v2/parser`**: Used to construct and configure the Markdown parser.
* **`github.com/yuin/goldmark/v2/renderer/html`**: Used to construct and configure the HTML renderer.
* **`github.com/yuin/goldmark/v2/testutil`**: Provides helper functions for running data-driven test cases from external files.

---

## Function Detailed Explanation

### `TestTaskList(t *testing.T)`

`TestTaskList` is the primary unit test function that executes test cases defined in an external test file (`_test/tasklist.txt`).

#### Key Execution Steps:

1. **Markdown Converter Initialization (`testutil.NewMarkdownToStringFunc`)**
   Configures a Markdown transformation pipeline that takes raw Markdown text and renders it to an HTML string.

   * **Parser Configuration**:
     * Instantiates a new parser via `parser.New()`.
     * Configures the parser with `parser.WithExtensions(NewTaskListItemParser())` to register the parser for task list items.

   * **HTML Renderer Configuration**:
     * Instantiates a new HTML renderer via `html.New()`.
     * Applies `html.WithUnsafe()`, allowing raw or unsafe HTML rendering.
     * Registers the task list HTML renderer extension via `html.WithExtensions(NewTaskListItemHTMLRenderer())`.

2. **Data-Driven Test Execution (`testutil.DoTestCaseFile`)**
   * Invokes `testutil.DoTestCaseFile` to parse and run Markdown test cases stored in `_test/tasklist.txt`.
   * Passes the configured conversion function (`markdown`), the relative file path to test data (`_test/tasklist.txt`), the `testing.T` instance (`t`), and command-line argument flags (`testutil.ParseCliCaseArg()...`).

---

## Summary of Components

| Component / Function | Purpose |
| :--- | :--- |
| `TestTaskList` | Unit test function executing task list parsing and rendering test cases. |
| `NewTaskListItemParser()` | Configures the Markdown parser extension for task list items. |
| `NewTaskListItemHTMLRenderer()` | Configures the HTML renderer extension for task list items. |
| `html.WithUnsafe()` | Renderer option enabling unsafe HTML rendering. |
| `_test/tasklist.txt` | Path to the test case specification file containing test inputs and expected outputs. |
| `testutil.ParseCliCaseArg()` | Parses CLI arguments for filtering or controlling test case execution within `testutil`. |