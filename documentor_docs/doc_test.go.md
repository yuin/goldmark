# Documentation Guide: `doc_test.go`

## Overview

The `doc_test.go` file belongs to the `goldmark_test` package. Its primary purpose is to automatically test and validate Go code snippets embedded within the project's `README.md` file. 

By parsing `README.md` with Goldmark itself, extracting Go code blocks, generating temporary Go projects, and executing them, this test ensures that code examples in the documentation remain runnable and up to date.

---

## Package Constants

### `thisPackage`
```go
const thisPackage = "github.com/yuin/goldmark/v2"
```
Defines the module path for the Goldmark library being tested. It is used to construct `go.mod` files for temporary test environments and to distinguish third-party import dependencies from the current package.

---

## Core Functions

### `TestDoc(t *testing.T)`

`TestDoc` is the main entry point for the documentation test suite.

#### Process Flow:
1. **Directory & File Loading**:
   - Retrieves the current working directory (`cwd`).
   - Reads the content of `README.md`.
2. **Markdown Parsing**:
   - Creates a new parser (`parser.New()`) and parses `README.md` into an Abstract Syntax Tree (AST).
   - Traverses the AST using `ast.Walk` to collect all `*ast.CodeBlock` nodes entering the tree.
3. **Filtering Code Blocks**:
   - Checks if the code block language is explicit and set to `"go"`. Skips non-Go code blocks.
   - Checks if the info string contains `"no-run"`. Skips blocks marked with `"no-run"`.
4. **Code Parsing & Import Extraction**:
   - Uses regex `(?s)import\s*\((.*?)\)` to split the code snippet into an `import (...)` block and the code body.
   - Identifies third-party imports by reading lines within the import section containing `.` that do not reference `thisPackage`.
5. **Temporary Go Module Creation**:
   - Creates a temporary directory using `os.MkdirTemp`.
   - Writes a custom `go.mod` file with:
     - Module name `test`
     - Go version matching `runtime.Version()`
     - A requirement for `thisPackage` at `v<majorVersion>.0.0` (derived from the last character of `thisPackage`)
     - A `replace` directive mapping `thisPackage` to the local working directory (`cwd`).
6. **Generating `main.go`**:
   - Creates a `main.go` file inside the temporary directory.
   - Formats the extracted import block and places the snippet body inside `func main()`.
7. **Resolving Dependencies**:
   - For each detected third-party import, executes `go get <imp>@latest` inside the temporary directory using a 15-second timeout context.
   - Runs `go mod tidy` inside the temporary directory using a 15-second timeout context.
8. **Execution and Validation**:
   - Executes `go run main.go` in the temporary directory.
   - If execution fails (`err != nil`), reports an error to `testing.T` containing:
     - The line number in `README.md` where the code block starts (calculated via `posToLine`).
     - Indented content of the generated `main.go` file.
     - Indented execution error/output.

---

## Helper Functions

### `execCmdContext`

```go
func execCmdContext(ctx context.Context, d string, c string, s ...string) error
```

Executes a command within a specified directory and context, piping its `stdout` and `stderr` directly to `os.Stdout` and `os.Stderr`.

* **Parameters**:
  * `ctx` (`context.Context`): Controls timeout or cancellation for the command process.
  * `d` (`string`): Target working directory for the command (`cmd.Dir`). If empty, uses the current execution directory.
  * `c` (`string`): The command executable name (e.g., `"go"`).
  * `s` (`...string`): Arguments to pass to the command executable.
* **Returns**: An `error` if the command execution fails.

---

### `posToLine`

```go
func posToLine(source []byte, pos int) int
```

Calculates the 1-based line number in a source byte slice for a given byte offset index.

* **Parameters**:
  * `source` (`[]byte`): The raw source byte content (e.g., contents of `README.md`).
  * `pos` (`int`): The target byte offset index within `source`.
* **Returns**: `int` — The 1-based line number corresponding to `pos`. Counts occurrences of newline (`\n`) characters prior to `pos`.

---

### `addIndent`

```go
func addIndent(s string, w int) string
```

Indents each line of a multiline string by a specified number of spaces.

* **Parameters**:
  * `s` (`string`): The string content to indent.
  * `w` (`int`): The width (number of space characters) for the indentation.
* **Returns**: `string` — The formatted string with every line prefixed by `w` spaces, right-trimmed of trailing spaces, and terminated with a newline character.

---

## Summary of Execution Lifecycle

```
[ README.md ]
     │
     ▼
[ Goldmark Parser ] ──► Extracts ast.CodeBlock nodes
     │
     ▼
[ Filter ] ──────────► Keep if language == "go" AND info string != "no-run"
     │
     ▼
[ Parsing ] ─────────► Separate import (...) block from main body
     │
     ▼
[ Temp Dir Creation ]► Create temporary directory with go.mod & main.go
     │                 (replace github.com/yuin/goldmark/v2 => local cwd)
     ▼
[ Dependencies ] ────► Execute `go get` for third-party imports & `go mod tidy`
     │
     ▼
[ Execution ] ───────► Execute `go run main.go`
     │
     ├──► Success: Temporary folder removed; test passes.
     └──► Failure: Report error with README.md line number and output.
```