# Technical Documentation: `extension/gfm.go`

## Overview

The `extension/gfm.go` file provides bundled extensions for GitHub Flavored Markdown (GFM) within the `goldmark` (v2) Markdown processing engine. Instead of requiring developers to manually register each individual GFM extension (e.g., tables, task lists, strikethrough, linkify), this file bundles them into unified **Parser** and **HTML Renderer** extensions.

---

## Package and Dependencies

* **Package:** `extension`
* **Imports:**
  * `github.com/yuin/goldmark/v2/parser`: Provides Goldmark's parser interfaces and configuration options.
  * `github.com/yuin/goldmark/v2/renderer/html`: Provides Goldmark's HTML renderer interfaces and configuration options.

---

## Key Components

### 1. GFM Parser Extension

#### `gfmParserExtension` Struct
An unexported empty struct that implements the `parser.Extension` interface.

```go
type gfmParserExtension struct {}
```

#### `NewGFMParser()`
Constructs and returns a new `parser.Extension` configured for GFM parsing.

* **Signature:** `func NewGFMParser() parser.Extension`
* **Returns:** `parser.Extension` (specifically, `*gfmParserExtension`)

#### `(*gfmParserExtension).ParserOptions()`
Collects and returns parser options required for GFM functionality.

* **Signature:** `func (e *gfmParserExtension) ParserOptions(c *parser.Config) []parser.Option`
* **Logic:** Aggregates options from four sub-parsers:
  1. `NewLinkifyParser().ParserOptions(c)`
  2. `NewTableParser().ParserOptions(c)`
  3. `NewStrikethroughParser().ParserOptions(c)`
  4. `NewTaskListItemParser().ParserOptions(c)`

---

### 2. GFM HTML Renderer Extension

#### `gfmHTMLRendererExtension` Struct
An unexported empty struct that implements the `html.Extension` interface.

```go
type gfmHTMLRendererExtension struct {}
```

#### `NewGFMHTMLRenderer()`
Constructs and returns a new `html.Extension` configured for rendering GFM constructs to HTML.

* **Signature:** `func NewGFMHTMLRenderer() html.Extension`
* **Returns:** `html.Extension` (specifically, `*gfmHTMLRendererExtension`)

#### `(*gfmHTMLRendererExtension).RendererOptions()`
Collects and returns HTML renderer options required for GFM output.

* **Signature:** `func (e *gfmHTMLRendererExtension) RendererOptions(c *html.Config) []html.Option`
* **Logic:** Aggregates options from three sub-renderers:
  1. `NewTableHTMLRenderer().RendererOptions(c)`
  2. `NewStrikethroughHTMLRenderer().RendererOptions(c)`
  3. `NewTaskListItemHTMLRenderer().RendererOptions(c)`

---

### 3. Package Variables

For convenience, the package provides pre-instantiated package-level variables that can be passed directly to Goldmark configuration without re-instantiating.

* **`GFMParser`**: Pre-configured instance of `parser.Extension` instantiated via `NewGFMParser()`.
* **`GFMHTMLRenderer`**: Pre-configured instance of `html.Extension` instantiated via `NewGFMHTMLRenderer()`.

---

## Component Summary Table

| Feature / Extension | Included in GFM Parser (`ParserOptions`) | Included in GFM HTML Renderer (`RendererOptions`) |
| :--- | :---: | :---: |
| **Linkify** | Yes (`NewLinkifyParser`) | No |
| **Table** | Yes (`NewTableParser`) | Yes (`NewTableHTMLRenderer`) |
| **Strikethrough** | Yes (`NewStrikethroughParser`) | Yes (`NewStrikethroughHTMLRenderer`) |
| **Task List Items** | Yes (`NewTaskListItemParser`) | Yes (`NewTaskListItemHTMLRenderer`) |

---

## How It Works

1. **Parser Aggregation**: When `GFMParser` (or an instance created by `NewGFMParser()`) is registered with Goldmark's parser engine, its `ParserOptions` method is invoked. This method executes the `ParserOptions` methods of the Linkify, Table, Strikethrough, and TaskListItem sub-parsers, combining all returned `parser.Option` values into a single slice.
2. **Renderer Aggregation**: When `GFMHTMLRenderer` (or an instance created by `NewGFMHTMLRenderer()`) is registered with Goldmark's HTML renderer, its `RendererOptions` method is invoked. This method executes the `RendererOptions` methods of the Table, Strikethrough, and TaskListItem sub-renderers, returning a combined slice of `html.Option` values.