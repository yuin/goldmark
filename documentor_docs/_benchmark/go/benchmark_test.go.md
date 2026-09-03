# Technical Documentation: `_benchmark/go/benchmark_test.go`

## Overview

The `_benchmark/go/benchmark_test.go` file contains performance benchmarking tests written in Go. Its primary purpose is to compare the rendering speed and efficiency of various Markdown parsing and HTML rendering libraries in the Go ecosystem against a standardized benchmark dataset (`_data.md`).

---

## File Information

* **File Path:** `_benchmark/go/benchmark_test.go`
* **Package Name:** `benchmark`
* **Test Framework:** Standard Go `testing` package (`testing.B`)

---

## External Dependencies

The benchmark imports standard Go modules and five distinct third-party Markdown rendering libraries:

### Standard Library
* `bytes`: Used for in-memory buffer operations during rendering.
* `os`: Used to read the benchmark source dataset file (`_data.md`).
* `testing`: Provides the benchmark runner capabilities via `testing.B`.

### Third-Party Markdown Engines
1. **gomarkdown**: `github.com/gomarkdown/markdown`
2. **Lute**: `github.com/88250/lute`
3. **golang-commonmark**: `gitlab.com/golang-commonmark/markdown`
4. **goldmark (v2)**: `github.com/yuin/goldmark/v2` (including `parser`, `renderer/html`, and `util`)
5. **goldmark (v1)**: `github.com/yuin/goldmark` (and `github.com/yuin/goldmark/renderer/html`)

---

## Functions

### 1. `BenchmarkMarkdown(b *testing.B)`

The main benchmark entry point executed by Go's testing tool. It sets up individual sub-benchmarks using `b.Run()` for each target Markdown library. Each sub-benchmark defines a standard closure (`render func(src []byte) ([]byte, error)`) adapting the library's specific API to a unified signature, which is then passed to `doBenchmark`.

#### Sub-Benchmarks Configured:

1. **`GoMarkdown(not CM)`**
   * **Library:** `gomarkdown`
   * **Implementation:** Calls `gomarkdown.ToHTML(src, nil, nil)` with default extensions and renderer parameters set to `nil`.

2. **`Lute`**
   * **Library:** `lute`
   * **Initialization/Configuration:** Creates an engine instance via `lute.New()` and explicitly disables several GFM and formatting features:
     * `SetGFMAutoLink(false)`
     * `SetGFMStrikethrough(false)`
     * `SetGFMTable(false)`
     * `SetGFMTaskListItem(false)`
     * `SetCodeSyntaxHighlight(false)`
     * `SetSoftBreak2HardBreak(false)`
     * `SetAutoSpace(false)`
     * `SetFixTermTypo(false)`
   * **Implementation:** Renders string-converted byte arrays via `luteEngine.MarkdownStr("Benchmark", util.BytesToReadOnlyString(src))` and converts the result back to bytes using `util.StringToReadOnlyBytes`.

3. **`golang-commonmark`**
   * **Library:** `gitlab.com/golang-commonmark/markdown`
   * **Initialization:** Instantiates markdown parser with XHTML output enabled: `markdown.New(markdown.XHTMLOutput(true))`.
   * **Implementation:** Writes rendered bytes into a `bytes.Buffer` using `md.Render(&out, src)`.

4. **`goldmark/v2`**
   * **Library:** `github.com/yuin/goldmark/v2`
   * **Initialization:** Creates a parser (`parser.New()`) and an HTML renderer configured with XHTML rendering and unsafe HTML allowance (`html.New(html.WithXHTML(), html.WithUnsafe())`).
   * **Implementation:** Parses source bytes into an AST using `gp.Parse(src)` and renders the parsed tree into a `bytes.Buffer` using `gr.Render(&out, src, tree)`.

5. **`goldmark/v1`**
   * **Library:** `github.com/yuin/goldmark` (v1)
   * **Initialization:** Creates a converter instance configured with XHTML rendering and unsafe HTML allowance using `goldmark.New(goldmark.WithRendererOptions(v1html.WithXHTML(), v1html.WithUnsafe()))`.
   * **Implementation:** Converts source bytes directly into a `bytes.Buffer` using `markdown.Convert(src, &out)`.

---

### 2. `doBenchmark(b *testing.B, render func(src []byte) ([]byte, error))`

A helper adapter function that executes the benchmarking loop under uniform conditions across all library implementations.

#### Logic Flow:
1. **Timer Pause:** Calls `b.StopTimer()` to pause timer metrics while reading input file data.
2. **File Loading:** Reads the benchmark input dataset from the relative path `_data.md` using `os.ReadFile("_data.md")`. If an error occurs, the benchmark halts immediately via `b.Fatal(err)`.
3. **Timer Resume:** Calls `b.StartTimer()` to resume measurement right before the iteration loop.
4. **Execution Loop:** Runs `b.N` iterations invoking the passed `render` closure with the `source` bytes.
5. **Validation Checks:** Inside the loop:
   * Checks if the `render` call returned an error. If true, halts with `b.Fatal(err)`.
   * Checks if the length of the generated HTML output `out` is less than 100 bytes (`len(out) < 100`). If true, halts the test with `b.Fatal("No result")` to ensure valid rendering occurred.

---

## File Dependencies for Execution

To successfully run the benchmark function in this file, a file named `_data.md` containing sample Markdown content must exist in the working directory where the test is executed.