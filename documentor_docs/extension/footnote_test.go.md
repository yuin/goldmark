# Technical Documentation: `extension/footnote_test.go`

## Overview

The `extension/footnote_test.go` file contains unit tests for the footnote extension in the Goldmark Markdown processor library. It validates the parsing and rendering behavior of footnotes, including default functionality loaded from external test files and custom configuration options (such as custom link classes, link titles, ID prefixes, and dynamic prefix functions using AST transformers).

---

## Package and Dependencies

* **Package**: `extension`
* **Imports**:
  * `testing`: Standard Go testing package.
  * `github.com/yuin/goldmark/v2/ast` (aliased as `gast`): Provides Abstract Syntax Tree (AST) node structures.
  * `github.com/yuin/goldmark/v2/parser`: Markdown parsing interfaces and configuration options.
  * `github.com/yuin/goldmark/v2/renderer/html`: HTML renderer options and utilities.
  * `github.com/yuin/goldmark/v2/testutil`: Testing utility helpers for running Markdown test cases.
  * `github.com/yuin/goldmark/v2/text`: Interfaces for reading source text.
  * `github.com/yuin/goldmark/v2/util`: General utility functions and prioritized structure helpers.

---

## Code Components

### 1. `TestFootnote(t *testing.T)`

This function executes data-driven integration tests for default footnote parsing and rendering.

* **Parser Configuration**: Instantiated with `NewFootnoteParser()`.
* **Renderer Configuration**: Instantiated with `html.WithUnsafe()` and `NewFootnoteHTMLRenderer()`.
* **Execution**: Uses `testutil.DoTestCaseFile` to process test cases stored in `_test/footnote.txt`. Accepts command-line argument overrides passed via `testutil.ParseCliCaseArg()`.

---

### 2. `footnoteID` Struct and `Transform` Method

A custom AST transformer implementation used to test dynamic ID prefix resolution via document metadata.

```go
type footnoteID struct {
}

func (a *footnoteID) Transform(node *gast.Document, _ text.Reader, _ parser.Context)
```

* **Purpose**: Implements the `parser.ASTTransformer` interface.
* **Logic**: Mutates the root `gast.Document` node metadata by setting the key `"footnote-prefix"` to `"article12-"`.

---

### 3. `TestFootnoteOptions(t *testing.T)`

This function tests advanced footnote customization options provided by `NewFootnoteHTMLRenderer`. It consists of two sub-test scenarios evaluated using `testutil.DoTestCase`.

#### Test Case 1: Static Customization Options
Tests standard static options supplied directly to `NewFootnoteHTMLRenderer`:

* **Options Tested**:
  * `WithIDPrefix("article12-")`: Prepends static string `"article12-"` to element IDs.
  * `WithLinkClass("link-class")`: Applies class `link-class` to footnote reference links.
  * `WithBacklinkClass("backlink-class")`: Applies class `backlink-class` to backlink anchors.
  * `WithLinkTitle("link-title-%%-^^")`: Defines title template for reference links.
  * `WithBacklinkTitle("backlink-title")`: Defines title attribute for backlinks.
  * `WithBacklinkHTML("^")`: Overrides the default backlink inner HTML string with `"^"`.

* **Test Input**: Markdown text containing multiple references to the same footnote (`[^1]`) and an additional footnote (`[^2]`), along with their corresponding definitions.
* **Verified HTML Output Elements**:
  * `<sup>` tags with formatted IDs (e.g., `id="article12-fnref:1"`, `id="article12-fnref1:1"`).
  * `<a>` tags with `class="link-class"` and generated title strings (e.g., `title="link-title-2-1"`).
  * Footer markup containing `<div class="footnotes" role="doc-endnotes">`, `<ol>`, `<li id="article12-fn:1">`, and backlink tags with `class="backlink-class"` and inner HTML `^`.

#### Test Case 2: Dynamic ID Prefix via Function
Tests dynamic prefix evaluation using an AST transformer and standard context metadata.

* **Parser Configuration**: 
  * Registers `footnoteID` as an AST transformer using `util.Prioritized[parser.ASTTransformer](&footnoteID{}, 100)`.
  * Includes `NewFootnoteParser()`.
* **Renderer Configuration**:
  * Uses `WithIDPrefixFunction(...)` which dynamically inspects `n.OwnerDocument().Metadata()["footnote-prefix"]`.
  * If the key exists, converts the string value to read-only bytes using `util.StringToReadOnlyBytes(...)`; otherwise returns `nil`.
  * Combines `WithIDPrefixFunction` with `WithLinkClass`, `WithBacklinkClass`, `WithLinkTitle`, `WithBacklinkTitle`, and `WithBacklinkHTML`.
* **Verification**: Verifies that the rendered HTML matches the exact output produced by static configuration in Test Case 1.

---

## Test Output Structure Comparison

The tests verify that generated HTML follows the standard Markdown endnote structure:

```html
<p>That's some text with a footnote.<sup id="article12-fnref:1"><a href="#article12-fn:1" class="link-class" title="link-title-2-1" role="doc-noteref">1</a></sup></p>
<p>Same footnote.<sup id="article12-fnref1:1"><a href="#article12-fn:1" class="link-class" title="link-title-2-1" role="doc-noteref">1</a></sup></p>
<p>Another one.<sup id="article12-fnref:2"><a href="#article12-fn:2" class="link-class" title="link-title-1-2" role="doc-noteref">2</a></sup></p>
<div class="footnotes" role="doc-endnotes">
<hr>
<ol>
<li id="article12-fn:1">
<p>And that's the footnote.&#160;<a href="#article12-fnref:1" class="backlink-class" title="backlink-title" role="doc-backlink">^</a>&#160;<a href="#article12-fnref1:1" class="backlink-class" title="backlink-title" role="doc-backlink">^</a></p>
</li>
<li id="article12-fn:2">
<p>Another footnote.&#160;<a href="#article12-fnref:2" class="backlink-class" title="backlink-title" role="doc-backlink">^</a></p>
</li>
</ol>
</div>
```