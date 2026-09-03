# Technical Documentation: `extension/definition_list.go`

## Overview

The `extension/definition_list.go` file implements support for **PHP Markdown Extra Definition Lists** within the Goldmark Markdown parsing framework. It provides parsing components to recognize definition lists in Markdown text and rendering components to convert definition list AST nodes into standard HTML elements (`<dl>`, `<dt>`, `<dd>`).

---

## Architecture & Component Breakdown

The file is organized into four main functional areas:
1. **Block Parsers** (`definitionListParser` and `definitionDescriptionParser`)
2. **Parser Extension** (`definitionListParserExtension`)
3. **HTML Renderers** (`definitionListHTMLRendererExtension`)
4. **Exported Package Variables and Constructors**

---

## 1. Block Parsers

Definition list parsing relies on two cooperative block parsers: `definitionListParser` and `definitionDescriptionParser`. Both register `:` as their trigger character.

### `definitionListParser`

Handles the outer `DefinitionList` block structure, managing line continuation and block alignment.

* **Trigger**: `:`
* **`Open(parent gast.Node, reader text.Reader, pc parser.Context) (gast.Node, parser.State)`**:
  * Evaluates whether the current line starts with a colon (`:`) followed by at least 1 space without block indentation (`indent == 0`).
  * Rejects parent nodes that are already instances of `*ast.DefinitionList`.
  * Checks the preceding sibling (`last`). If it is a `Paragraph`, the paragraph lines will serve as definition terms.
  * Adjusts list offset width based on space indentation following `:`. If space width is $\ge 8$, it caps the width calculation to treat 5 spaces as the prefix padding.
  * Sets temporary paragraphs on the `DefinitionList` AST node to track terms pending creation.
* **`Continue(node gast.Node, reader text.Reader, _ parser.Context) parser.State`**:
  * Continues processing if the line is blank.
  * Checks line indentation against `list.Offset()`. If the indentation is less than the required offset, returns `parser.Close` to end the definition block.
  * Advances reader padding if indentation matches or exceeds `list.Offset()`.
* **`Close`**: No-op.
* **`CanInterruptParagraph()`**: Returns `true`.
* **`CanAcceptIndentedLine()`**: Returns `false`.

### `definitionDescriptionParser`

Handles the creation of `DefinitionTerm` nodes from temporary paragraphs and creates `DefinitionDescription` nodes.

* **Trigger**: `:`
* **`Open(parent gast.Node, reader text.Reader, pc parser.Context) (gast.Node, parser.State)`**:
  * Ensures the parent node is a `*ast.DefinitionList`.
  * Extracts any temporary paragraph (`list.TemporaryParagraph()`).
  * Iterates over each line in the temporary paragraph, creates a new `ast.DefinitionTerm` node for each line, trims trailing whitespace, and appends the terms as children of the `DefinitionList`.
  * Removes the original temporary paragraph node from the AST.
  * Advances the reader position beyond the colon marker and space padding.
  * Returns a new `ast.NewDefinitionDescription()` node with `parser.HasChildren`.
* **`Continue`**: Returns `parser.Continue | parser.HasChildren` (actual end detection is managed by `definitionListParser`).
* **`Close(node gast.Node, _ text.Reader, _ parser.Context)`**:
  * Sets `desc.IsTight` on `ast.DefinitionDescription` based on whether there were blank lines preceding it (`!desc.HasBlankPreviousLines()`).
* **`CanInterruptParagraph()`**: Returns `true`.
* **`CanAcceptIndentedLine()`**: Returns `false`.

---

## 2. Parser Extension

### `definitionListParserExtension`

Connects the block parsers to the Goldmark parser framework.

* **`NewDefinitionListParser()`**: Factory function returning `parser.Extension`.
* **`ParserOptions(_ *parser.Config)`**: Registers the parsers with priorities:
  * `definitionListParser`: Priority `101`
  * `definitionDescriptionParser`: Priority `102`

---

## 3. HTML Extensions & Renderers

### `definitionListHTMLRendererExtension`

Maps AST definition list nodes to HTML elements and manages tight block formatting.

* **`NewDefinitionListHTMLRenderer()`**: Factory function returning `html.Extension`.
* **`RendererOptions(_ *html.Config)`**: Configures HTML renderers for three AST kinds:
  * `ast.KindDefinitionList` $\rightarrow$ `renderDefinitionList`
  * `ast.KindDefinitionTerm` $\rightarrow$ `renderDefinitionTerm`
  * `ast.KindDefinitionDescription` $\rightarrow$ `renderDefinitionDescription`
  * Also sets the tight block checking function `html.WithIsInTightBlockFunc(definitionListIsInTightBlock)`.

### Attribute Filters

Determines which HTML attributes are allowed on elements (defaults to `html.GlobalAttributeFilter`):
* `DefinitionListAttributeFilter`
* `DefinitionTermAttributeFilter`
* `DefinitionDescriptionAttributeFilter`

### Node Rendering Functions

1. **`renderDefinitionList`**:
   * **Entering**: Writes `<dl` + attributes + `>\n` (or `<dl>\n` if no attributes exist).
   * **Exiting**: Writes `</dl>\n`.

2. **`renderDefinitionTerm`**:
   * **Entering**: Writes `<dt` + attributes + `>` (or `<dt>` if no attributes exist).
   * **Exiting**: Writes `</dt>\n`.

3. **`renderDefinitionDescription`**:
   * **Entering**: Writes `<dd` + attributes + `>` (or `<dd>`). If `n.IsTight` is `false`, adds a trailing newline (`>\n`).
   * **Exiting**: Writes `</dd>\n`.

---

## 4. Helper Functions

### `definitionListIsInTightBlock(n gast.Node) bool`

Determines if a node resides within a tight block context:
1. Returns `false` if `n` has no parent.
2. If the parent is `*ast.DefinitionDescription`, returns `desc.IsTight`.
3. If the grandparent is `*gast.List`, returns `list.IsTight`.
4. Otherwise, returns `false`.

---

## 5. Exported Package Variables & Functions Summary

| Identifier | Type | Description |
| :--- | :--- | :--- |
| `NewDefinitionListParser()` | `func() parser.Extension` | Returns a new parser extension for definition lists. |
| `NewDefinitionListHTMLRenderer()` | `func() html.Extension` | Returns a new HTML renderer extension for definition lists. |
| `DefinitionListParser` | `parser.Extension` | Default instance of definition list parser extension. |
| `DefinitionListHTMLRenderer` | `html.Extension` | Default instance of definition list HTML renderer. |
| `DefinitionListAttributeFilter` | `html.AttributeFilter` | Filter for `<dl>` element attributes (default: `GlobalAttributeFilter`). |
| `DefinitionTermAttributeFilter` | `html.AttributeFilter` | Filter for `<dt>` element attributes (default: `GlobalAttributeFilter`). |
| `DefinitionDescriptionAttributeFilter` | `html.AttributeFilter` | Filter for `<dd>` element attributes (default: `GlobalAttributeFilter`). |