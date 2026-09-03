# Technical Documentation: `fuzz/fuzz_test.go`

## Overview

The `fuzz/fuzz_test.go` file implements automated fuzz testing for the `goldmark` Markdown processing library using Go's built-in `testing` fuzzing framework. Its primary purpose is to pass randomized string inputs into a heavily configured Goldmark parser and HTML renderer to detect potential crashes, memory corruption, or unexpected panics.

---

## Imported Packages

*   **`bytes`**: Provides `bytes.Buffer` to hold the output of HTML rendering.
*   **`encoding/json`**: Used to parse JSON-formatted test spec files for seed corpus generation.
*   **`os`**: Used to read external spec files from the file system (`os.ReadFile`).
*   **`testing`**: Provides the Go standard library fuzzing infrastructure (`testing.F`, `testing.T`).
*   **`github.com/yuin/goldmark/v2/extension`**: Provides extension parsers and HTML renderers (Definition List, Footnotes, GFM, Typographer, Linkify, Tables, Task Lists).
*   **`github.com/yuin/goldmark/v2/parser`**: Provides core parsing structures and options.
*   **`github.com/yuin/goldmark/v2/renderer/html`**: Provides HTML rendering options and constructors.
*   **`github.com/yuin/goldmark/v2/util`**: Utility module used to convert strings into read-only byte slices efficiently.

---

## Functions Overview

### 1. `FuzzDefault(f *testing.F)`

`FuzzDefault` is the primary entry point for the fuzz test suite run by `go test -fuzz=FuzzDefault`.

#### Workflow:
1. **Load Seed Corpus File**: Reads `../_test/spec.json` using `os.ReadFile`. If reading fails, the function panics.
2. **Parse Test Cases**: Unmarshals the JSON contents into a slice of maps (`[]map[string]any`). If unmarshaling fails, the function panics.
3. **Populate Fuzz Corpus**: Iterates through each test case in `testCases` and adds the value associated with the `"markdown"` key into the fuzzing target's seed corpus using `f.Add()`.
4. **Execute Fuzzing Target**: Invokes the internal helper function `fuzz(f)` to start the target function.

---

### 2. `fuzz(f *testing.F)`

`fuzz` is an unexported helper function that configures the fuzz target function via `f.Fuzz(...)`.

#### Workflow:
1. **Define Target Function**: Accepts a target function `func(_ *testing.T, orig string)` that receives generated string inputs (`orig`).
2. **Configure Parser**: Initializes a new Goldmark parser (`parser.New`) configured with:
   * **Options**:
     * `parser.WithAutoHeadingID()`: Automatically generates IDs for headings.
     * `parser.WithAttribute()`: Enables custom attribute support.
   * **Parser Extensions**:
     * Definition List (`extension.NewDefinitionListParser()`)
     * Footnote (`extension.NewFootnoteParser()`)
     * GitHub Flavored Markdown / GFM (`extension.NewGFMParser()`)
     * Typographer (`extension.NewTypographerParser()`)
     * Linkify (`extension.NewLinkifyParser()`)
     * Table (`extension.NewTableParser()`)
     * Task List Item (`extension.NewTaskListItemParser()`)
3. **Configure HTML Renderer**: Initializes a new HTML renderer (`html.New`) configured with:
   * **Options**:
     * `html.WithUnsafe()`: Allows rendering unsafe HTML tags/attributes.
     * `html.WithXHTML()`: Formats rendered output as XHTML.
   * **Renderer Extensions**:
     * Definition List (`extension.NewDefinitionListHTMLRenderer()`)
     * Footnote (`extension.NewFootnoteHTMLRenderer()`)
     * GFM (`extension.NewGFMHTMLRenderer()`)
     * Table (`extension.NewTableHTMLRenderer()`)
     * Task List Item (`extension.NewTaskListItemHTMLRenderer()`)
4. **Input Processing**:
   * Converts the input string `orig` to a read-only byte slice using `util.StringToReadOnlyBytes(orig)`.
5. **Parse and Render**:
   * Parses the source bytes `p.Parse(src)` into an AST.
   * Renders the AST into a `bytes.Buffer` instance `b`.
6. **Error Handling**:
   * If `r.Render()` returns a non-nil error, the test panics.

---

## Logic Summary

```
FuzzDefault(f *testing.F)
    │
    ├── Read "../_test/spec.json"
    ├── Unmarshal JSON to []map[string]any
    ├── Add c["markdown"] inputs to seed corpus via f.Add()
    └── Call fuzz(f)
           │
           └── f.Fuzz(func(_ *testing.T, orig string))
                  ├── Construct Parser with extensions & options
                  ├── Construct Renderer with extensions & options
                  ├── Convert orig string -> read-only []byte
                  ├── Parse input -> AST
                  ├── Render AST -> bytes.Buffer
                  └── Panic if Render returns an error
```