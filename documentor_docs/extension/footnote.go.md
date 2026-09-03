# Technical Documentation: `extension/footnote.go`

## Overview

The `extension/footnote.go` file provides PHP Markdown Extra-style footnote functionality for the Goldmark Markdown parser and HTML renderer. It includes:

1. **AST / Context State Management**: Interfaces and structures to track footnote definitions and references during parsing and rendering.
2. **Block Parser**: Parses footnote definitions (e.g., `[^1]: Footnote text`).
3. **Inline Parser**: Parses footnote references (e.g., `[^1]`).
4. **HTML Renderer & Decorator**: Renders inline reference superscripts and builds the end-of-document footnotes section (`<div class="footnotes" ...>`) containing definitions and backlinks.
5. **Extensions**: Extensions (`FootnoteParser` and `FootnoteHTMLRenderer`) to easily integrate footnote capabilities into Goldmark.

---

## Data Models and Interfaces

### 1. `FootnoteReference`
Represents a footnote reference object.

```go
type FootnoteReference interface {
    Label() []byte
    Index() int
    RefIndex() int
}
```

* **`Label()`**: Returns the byte slice label of the reference.
* **`Index()`**: Returns the 1-based display index assigned to the target footnote definition.
* **`RefIndex()`**: Returns the 0-based index of this specific reference among all references pointing to the same footnote definition.

**Implementation**: `footnoteRefInfo`
* Instantiated via `newFootnoteReferenceFromNode(node *ast.FootnoteReference, src []byte) FootnoteReference`. Initializes `index` and `refIndex` to `-1`.

---

### 2. `FootnoteDefinition`
Represents a footnote definition data object.

```go
type FootnoteDefinition interface {
    Label() []byte
}
```

* **`Label()`**: Returns the byte slice label of the definition.

**Implementation**: `footnoteDefInfo`
* Instantiated via `newFootnoteDefinitionFromNode(node *ast.FootnoteDefinition, src []byte) FootnoteDefinition`.

---

### 3. `Footnotes`
Manages definitions and references during the parsing lifecycle.

```go
type Footnotes interface {
    AddDefinition(def FootnoteDefinition)
    AddReference(ref FootnoteReference) bool
    FindByLabel(label []byte) FootnoteDefinition
}
```

* **`AddDefinition(def FootnoteDefinition)`**: Registers a definition if its label has not already been registered.
* **`AddReference(ref FootnoteReference) bool`**: Registers a reference. Returns `false` if no matching definition exists. If a match is found, assigns the display index (`index`) and reference ordinal (`refIndex`), incrementing the global definition counter if this is the first reference to the definition.
* **`FindByLabel(label []byte) FootnoteDefinition`**: Returns the `FootnoteDefinition` matching `label`, or `nil`.

**Implementation**: `footnotes`
* Structure fields:
  * `definitionIndex int`: Tracks the auto-incrementing display index (starts at 1).
  * `defsByLabel map[string]*defData`: Maps label strings to `defData` (which contains the `FootnoteDefinition`, assigned display index, and slice of reference indices).
* Constructor: `newFootnotes()`

**Context Function**:
* **`ContextFootnotes(pc parser.Context) Footnotes`**: Retrieves or initializes the `Footnotes` instance stored inside the parser context under `footnoteContextKey`.

---

## Block Parser: Footnote Definition

The block parser processes footnote definitions at the block level.

### `footnoteBlockParser`
* **Trigger**: `[`
* **Default Instance**: `defaultFootnoteBlockParser` / `newFootnoteBlockParser()`
* **`CanInterruptParagraph()`**: Returns `true`.
* **`CanAcceptIndentedLine()`**: Returns `false`.

#### Parse Workflow (`Open`, `Continue`, `Close`)
1. **`Open(...)`**:
   * Inspects line for `[^<label>]:`.
   * Calls `findLabelClosure(...)` to find the closing `]`.
   * Verifies that the label is non-blank and followed immediately by `:`.
   * Creates an `ast.FootnoteDefinition(label)` AST node.
   * Advances the text reader past the prefix and sets padding if additional inline content exists.
2. **`Continue(...)`**:
   * Evaluates continuation lines. Blank lines continue the block.
   * Non-blank lines require 4 spaces of indentation (`util.IndentPosition(..., 4)`). If not indented, returns `parser.Close`.
3. **`Close(...)`**:
   * Triggered when the block closes. Registers the definition with `ContextFootnotes(pc).AddDefinition(...)`.

---

## Inline Parser: Footnote Reference

The inline parser processes footnote references within inline text.

### `footnoteParser`
* **Trigger**: `!`, `[`
* **Default Instance**: `defaultFootnoteParser` / `newFootnoteParser()`

#### Parse Workflow (`Parse`)
1. Checks for `[^<label>]` or `![^<label>]`.
2. Locates closing `]` via `findLabelClosure`.
3. Creates an `ast.FootnoteReference(label)`.
4. Registers the reference via `ContextFootnotes(pc).AddReference(...)`. If no matching definition exists, returns `nil` (parsing fails, falling back to standard text/link handling).
5. Populates `Index` and `RefIndex` on the `ast.FootnoteReference` node.
6. If prefixed with `!`, appends a literal `!` text node to the parent AST node before returning the reference node.

---

## Label Parsing Helper

### `findLabelClosure(bs []byte) int`
Iterates over a byte slice to find the unescaped closing bracket `]`. Escaped brackets (`\]`) are skipped. Returns the index of `]` or `-1` if not found.

---

## HTML Renderer Configuration and Options

### `footnoteHTMLRendererConfig`
Holds renderer configuration options:
* `XHTML bool`: Controls XHTML vs HTML output format (e.g., `<hr />` vs `<hr>`).
* `IDPrefix []byte`: Constant prefix applied to footnote `id` attributes.
* `IDPrefixFunction func(gast.Node) []byte`: Dynamic function to compute `id` attribute prefixes.
* `LinkTitle []byte`: Title attribute for footnote reference links.
* `BacklinkTitle []byte`: Title attribute for back links.
* `LinkClass []byte`: CSS class for footnote reference links (Default: `"footnote-ref"`).
* `BacklinkClass []byte`: CSS class for footnote backlinks (Default: `"footnote-backref"`).
* `BacklinkHTML []byte`: HTML content/symbol for backlink anchors (Default: `"&#x21a9;&#xfe0e;"`).

### Functional Options
All options implement `FootnoteHTMLRendererOption` via `applyFootnoteHTMLRendererOption(*footnoteHTMLRendererConfig)`:

* **`WithIDPrefix[T []byte | string](a T)`**: Sets `IDPrefix`.
* **`WithIDPrefixFunction(a func(gast.Node) []byte)`**: Sets `IDPrefixFunction`.
* **`WithLinkTitle[T []byte | string](a T)`**: Sets `LinkTitle`.
* **`WithBacklinkTitle[T []byte | string](a T)`**: Sets `BacklinkTitle`.
* **`WithLinkClass[T []byte | string](a T)`**: Sets `LinkClass`.
* **`WithBacklinkClass[T []byte | string](a T)`**: Sets `BacklinkClass`.
* **`WithBacklinkHTML[T []byte | string](a T)`**: Sets `BacklinkHTML`.

---

## HTML Renderer Extension

### `footnoteHTMLRendererExtension`
Instantiated via `NewFootnoteHTMLRenderer(opts ...FootnoteHTMLRendererOption)`.
Exported default instance: `FootnoteHTMLRenderer`.

#### Rendering Methods
* **`renderFootnoteReference`**:
  * Renders `<sup id="[prefix]fnref[refIndex]:[index]"><a href="#[prefix]fn:[index]" class="..." title="..." role="doc-noteref">[index]</a></sup>`.
  * `refIndex` is omitted from the ID if it is `0`.
  * Templates `LinkClass` and `LinkTitle` using `applyFootnoteTemplate`.
* **`renderFootnoteDefinition`**:
  * Returns `gast.WalkSkipChildren, nil`. Definition bodies are skipped during standard AST walking and rendered at the end of the document by the decorator.
* **`idPrefix(node gast.Node) []byte`**:
  * Resolves `IDPrefix` or evaluates `IDPrefixFunction(node)`.
* **`RendererOptions(c *html.Config)`**:
  * Registers node renderers for `ast.KindFootnoteReference` and `ast.KindFootnoteDefinition`.
  * Hooks the document decorator (`footnoteDecorator`) to `gast.KindDocument`.

---

## Document Decorator (`footnoteDecorator`)

The `footnoteDecorator` intercepts document rendering to construct the footer list of referenced footnotes.

### Process Flow in `Decorate`

1. **Entering Document (`entering == true`)**:
   * Walks the AST to build lookup maps:
     * `defsMap`: Maps string labels to `*ast.FootnoteDefinition`.
     * `infos`: Maps string labels to `*defInfo` (tracking display index and reference indices).
   * Stores `defsMap` and `infos` in the `renderer.Context` under `footnoteDefsKey` and `footnoteDefsInfoKey`.
   * Delegates execution to the inner node renderer.

2. **Exiting Document (`entering == false`)**:
   * Collects all definitions that have at least one reference (`len(info.references) > 0`).
   * If no definitions were referenced, passes through without rendering footnotes.
   * Sorts referenced definitions by display index (`info.index`) ascending using `slices.SortFunc`.
   * Renders the container wrapper:
     ```html
     <div class="footnotes" role="doc-endnotes">
     <hr> <!-- or <hr /> if XHTML -->
     <ol>
     ```
   * Renders each referenced definition via `renderDefinition(...)`.
   * Closes `</ol></div>`.

### Helper Decorator Methods
* **`renderDefinition`**:
  * Renders `<li id="[prefix]fn:[index]" ...>`.
  * If the definition has attributes, renders attributes matching `html.ListItemAttributeFilter`.
  * Renders child blocks. If the last child is a paragraph (`gast.IsParagraph`), renders it using `renderParagraphWithBacklinks`.
  * If no paragraph exists, appends backlinks directly to the list item.
* **`renderParagraphWithBacklinks`**:
  * Renders `<p>`, renders paragraph children, appends backlinks via `renderBacklinks`, and closes `</p>`.
* **`renderBacklinks`**:
  * For each reference index in `info.references`, writes:
    ```html
    &#160;<a href="#[prefix]fnref[refIdx]:[index]" class="..." title="..." role="doc-backlink">[BacklinkHTML]</a>
    ```
  * Templatizes `BacklinkClass`, `BacklinkTitle`, and `BacklinkHTML` using `applyFootnoteTemplate`.

---

## Parser Extension

### `footnoteParserExtension`
Created via `NewFootnoteParser()`.
Exported default instance: `FootnoteParser`.

* **`ParserOptions(_ *parser.Config)`**:
  * Registers `footnoteBlockParser` with priority `999`.
  * Registers `footnoteParser` with priority `101`.

---

## Utility Function

### `applyFootnoteTemplate(b []byte, index, refCount int) []byte`
Replaces template placeholders in byte slices:
* `^^`: Replaced by the display index string (`index`).
* `%%`: Replaced by the total reference count string (`refCount`).

**Fast Path**:
Scans the byte slice first. If neither `^^` nor `%%` is found, it returns `b` directly without memory allocation.