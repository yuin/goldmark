# Strikethrough Extension (`extension/strikethrough.go`)

## Overview

The `extension/strikethrough.go` file provides Goldmark v2 syntax extension components for parsing and rendering **strikethrough** markdown elements (typically denoted using the tilde `~` character). 

It implements the parser and HTML renderer extensions required to convert strikethrough markdown text into HTML `<del>` elements within the Goldmark Markdown processing pipeline.

---

## Key Components

The file consists of four primary structural components:

1. **Delimiter Processor** (`strikethroughDelimiterProcessor`)
2. **Inline Parser** (`strikethroughParser`)
3. **Parser Extension** (`strikethroughParserExtension`)
4. **HTML Renderer Extension** (`strikethroughHTMLRendererExtension`)

---

## Component Details

### 1. Delimiter Processor

#### `strikethroughDelimiterProcessor`

Handles matching opening and closing tilde (`~`) delimiters to construct strikethrough AST nodes.

* **`IsDelimiter(b byte) bool`**
  * Checks if a byte character is a valid strikethrough delimiter.
  * **Returns:** `true` if `b == '~'`, otherwise `false`.

* **`CanOpenCloser(opener, closer *parser.Delimiter) bool`**
  * Verifies whether an opening delimiter match can form a valid pair with a closing delimiter.
  * **Returns:** `true` if `opener.Char == closer.Char`.

* **`OnMatch(_ int) gast.Node`**
  * Invoked when a valid delimiter pair match is identified.
  * **Returns:** A new strikethrough AST node created via `ast.NewStrikethrough()`.

* **`defaultStrikethroughDelimiterProcessor`**
  * Internal package instance of `strikethroughDelimiterProcessor`.

---

### 2. Inline Parser

#### `strikethroughParser`

Implements Goldmark's `parser.InlineParser` interface to parse inline strikethrough markup.

* **`newStrikethroughParser() parser.InlineParser`**
  * Internal constructor returning `defaultStrikethroughParser`.

* **`Trigger() []byte`**
  * Specifies trigger characters that activate this inline parser.
  * **Returns:** `[]byte{'~'}`.

* **`Parse(_ gast.Node, block text.Reader, pc parser.Context) gast.Node`**
  * Handles inline parsing logic when encountering a `~` trigger character:
    1. Checks preceding character using `block.PrecedingCharacter()`. If it is `~`, returns `nil` (prevents double parsing inside existing delimiter runs).
    2. Inspects the upcoming line (`block.PeekLine()`) and counts consecutive `~` characters (`n`).
    3. If there are more than 2 consecutive tildes (`n > 2`), parsing aborts and returns `nil`.
    4. Delegates delimiter parsing to Goldmark's core via `parser.ParseDelimiter(block, 1, defaultStrikethroughDelimiterProcessor, pc)`.

* **`CloseBlock(_ gast.Node, _ parser.Context)`**
  * Empty implementation satisfying the `parser.InlineParser` interface requirement.

---

### 3. Parser Extension

#### `strikethroughParserExtension`

Exposes the inline parser as a configurable Goldmark `parser.Extension`.

* **`NewStrikethroughParser() parser.Extension`**
  * Constructor that creates a new parser extension instance for strikethrough parsing.

* **`ParserOptions(_ *parser.Config) []parser.Option`**
  * Configures parser options by registering `newStrikethroughParser()` with a priority score of `500`.

* **`StrikethroughParser`**
  * Exported package variable providing a default instance of `parser.Extension` via `NewStrikethroughParser()`.

---

### 4. HTML Renderer Extension

#### `strikethroughHTMLRendererExtension`

Implements `html.Extension` to render `ast.KindStrikethrough` AST nodes to HTML output.

* **`NewStrikethroughHTMLRenderer() html.Extension`**
  * Constructor that returns a new `strikethroughHTMLRendererExtension`.

* **`RendererOptions(_ *html.Config) []html.Option`**
  * Registers node renderers for node kind `ast.KindStrikethrough` using `r.renderStrikethrough`.

* **`StrikethroughAttributeFilter`**
  * Exported variable defined as `html.GlobalAttributeFilter`. Defines the allowed attribute names for rendering HTML `<del>` elements.

* **`renderStrikethrough(writer io.Writer, source []byte, n gast.Node, entering bool, rc renderer.Context) (gast.WalkStatus, error)`**
  * Writes HTML output for `ast.KindStrikethrough` nodes:
    * **Entering (`entering == true`):**
      * If node attributes exist (`n.Attributes() != nil`), opens tag as `<del`, renders attributes through `html.RenderAttributes`, filtered by `StrikethroughAttributeFilter`, and appends `>`.
      * If no attributes exist, writes `<del>`.
    * **Exiting (`entering == false`):**
      * Writes the closing tag `</del>`.
  * **Returns:** `(gast.WalkContinue, nil)`.

* **`StrikethroughHTMLRenderer`**
  * Exported package variable providing a default instance of `html.Extension` via `NewStrikethroughHTMLRenderer()`.

---

## Summary of Exported API

| Exported Identifier | Type | Description |
| :--- | :--- | :--- |
| `NewStrikethroughParser()` | `func() parser.Extension` | Constructor for creating a new strikethrough parser extension. |
| `StrikethroughParser` | `parser.Extension` | Default pre-instantiated parser extension. |
| `NewStrikethroughHTMLRenderer()` | `func() html.Extension` | Constructor for creating a new strikethrough HTML renderer extension. |
| `StrikethroughHTMLRenderer` | `html.Extension` | Default pre-instantiated HTML renderer extension. |
| `StrikethroughAttributeFilter` | `html.AttributeFilter` | Attribute filter set to `html.GlobalAttributeFilter` for `<del>` elements. |