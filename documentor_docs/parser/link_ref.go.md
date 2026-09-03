# Documentation: `parser/link_ref.go`

## Overview

The `parser/link_ref.go` file provides functionality for identifying, parsing, and extracting Markdown **Link Reference Definitions** from paragraph nodes within the Goldmark AST (Abstract Syntax Tree). 

When a paragraph is parsed, this module scans its text lines for link reference syntax (for example: `[label]: /url "title"`). If definitions are found, it:
1. Constructs `ast.LinkReferenceDefinition` AST nodes and inserts them into the syntax tree directly before the paragraph.
2. Registers the definitions into the parsing context (`Context`).
3. Removes the processed definition lines from the original paragraph node (or removes the paragraph node entirely if all lines were reference definitions).

---

## Dependencies

The file imports and relies on the following packages:
* `github.com/yuin/goldmark/v2/ast`: Defines AST structures such as `Paragraph`, `BlockNode`, and `NewLinkReferenceDefinition`.
* `github.com/yuin/goldmark/v2/text`: Handles underlying buffer readers and line position tracking (`Reader`, `BlockReader`).
* `github.com/yuin/goldmark/v2/util`: Provides utility functions like `IndentWidth`, `IsBlank`, and byte manipulation.

---

## Key Components

### 1. `LinkReferenceParagraphTransformer`

```go
var LinkReferenceParagraphTransformer ParagraphTransformer = &linkReferenceParagraphTransformer{}
```

* **Type**: `ParagraphTransformer` (exported package variable singleton)
* **Purpose**: Serves as the public entry point for transforming paragraph nodes containing link reference definitions.

---

### 2. `linkReferenceParagraphTransformer` (Unexported)

```go
type linkReferenceParagraphTransformer struct{}
```

Implements the `ParagraphTransformer` interface via its `Transform` method.

#### `Transform` Method
```go
func (p *linkReferenceParagraphTransformer) Transform(node *ast.Paragraph, reader text.Reader, pc Context)
```

**Process Flow**:
1. **Reader Initialization**: Extracts source lines from the `node` and creates a `text.BlockReader` to iterate through paragraph lines.
2. **Extraction Loop**:
   * Repeatedly calls `parseLinkReferenceDefinition(block, pc)`.
   * If a valid reference definition is parsed (`start > -1`):
     * If the definition starts at the beginning of the paragraph (`start == 0`), it inherits the paragraph's blank line prefix status via `ref.SetBlankPreviousLines(node.HasBlankPreviousLines())`.
     * Sets the position of the new reference node to match the source line start position.
     * Inserts the parsed `ref` node into the parent AST before `node`.
     * Records the line range (`[start, end]`) for subsequent line removal.
3. **Source Line Adjustment**:
   * Slices the original paragraph lines slice to remove all line ranges recorded in `removes`.
   * Adjusts index offsets as ranges are removed.
4. **AST Cleanup**:
   * If no lines remain in the paragraph (`len(lines) == 0`), the paragraph node is completely removed from its parent AST node (`node.Parent().RemoveChild(node)`).
   * If remaining lines exist, the paragraph's source lines are updated via `node.SetSource(lines)`.

---

### 3. `parseLinkReferenceDefinition` Function

```go
func parseLinkReferenceDefinition(block text.Reader, pc Context) (ast.BlockNode, int, int)
```

Parses a single link reference definition from the current position of the provided `text.Reader`.

* **Parameters**:
  * `block text.Reader`: A block reader for the paragraph's text content.
  * `pc Context`: The active parsing context where parsed link definitions are registered.
* **Returns**:
  * `ast.BlockNode`: The parsed link reference definition node (or `nil` if parsing fails).
  * `int`: The starting line index in the block (or `-1` if parsing fails).
  * `int`: The ending line index in the block (or `-1` if parsing fails).

#### Parsing Rules & Validation Sequence:

1. **Space & Indentation Check**:
   * Skips leading spaces.
   * Calculates indentation width using `util.IndentWidth`. If indentation exceeds 3 spaces, parsing fails (`returns nil, -1, -1`).
2. **Label Parsing**:
   * Expects the current line character to be opening bracket `[`.
   * Calls `findClosure(block, '[', ']')` to extract the label content.
   * Label content must not be blank (`util.IsBlank`).
   * Expects a colon `:` immediately following the label closing bracket.
3. **Destination Parsing**:
   * Advances past `:` and skips spaces.
   * Calls `parseLinkDestination(block)` to extract the URI/destination. If parsing fails, returns `nil, -1, -1`.
   * Evaluates `isNewLine`: whether the destination ends at a newline or blank line.
4. **Title Parsing (Optional)**:
   * Skips spaces after the destination and checks for title opener characters (`"`, `'`, or `(`).
   * **No Opener Present**:
     * If the destination was not followed by a newline/blank line (`!isNewLine`), parsing fails.
     * Otherwise, creates a title-less `LinkReferenceDefinition`, registers it with `pc.AddLinkDefinition`, and returns the node with line indices.
   * **Opener Present**:
     * Must be preceded by at least 1 space (`spaces > 0`).
     * Matches the corresponding closer (`"` for `"`, `'` for `'`, `)` for `(`).
     * Uses `findClosure(block, opener, closer)` to extract title bytes.
     * **Title Closure Not Found**: If invalid, falls back to a title-less definition (if `isNewLine` is true) and advances line.
     * **Title Closure Found**:
       * Checks trailing content on the line after the title.
       * If trailing non-blank content exists on the same line as title closure and `!isNewLine`, parsing fails.
       * Registers and constructs the definition node with title (`ast.WithLinkTitle(titleVal)`).

---

## Helper Function Invocation Summary

The file invokes the following package-internal functions (defined in adjacent files in `parser`):

| Function | Purpose |
| :--- | :--- |
| `findClosure(block, opener, closer)` | Scans `block` for matching delimiter pairs and returns the enclosed byte slice and a boolean indicating success. |
| `parseLinkDestination(block)` | Extracts the URL/destination from the reader. |
| `newLinkDefinitionFromNode(ref, source)` | Constructs a link definition structure from the AST node and source context for registration in `pc`. |