# Technical Documentation: `extension/ast/footnote.go`

## Overview

The `extension/ast/footnote.go` file defines custom Abstract Syntax Tree (AST) nodes for handling Markdown footnote references and footnote definitions in accordance with PHP Markdown Extra formatting specifications. It provides two core AST node types—`FootnoteReference` and `FootnoteDefinition`—that extend Goldmark v2's base inline and block nodes respectively.

---

## Package and Dependencies

* **Package**: `ast`
* **Imports**:
  * `gast "github.com/yuin/goldmark/v2/ast"`: Provides foundational AST types (`BaseInline`, `BaseBlock`, `NodeKind`, `NodeDump`).
  * `"github.com/yuin/goldmark/v2/text"`: Provides text representation primitives (`text.SingleLineValue`).

---

## Node Kinds

The file defines two unique `gast.NodeKind` identifiers used to classify footnote AST nodes during AST traversal and rendering:

* **`KindFootnoteReference`**: Node kind identifier for `FootnoteReference` (created with `gast.NewNodeKind("FootnoteReference")`).
* **`KindFootnoteDefinition`**: Node kind identifier for `FootnoteDefinition` (created with `gast.NewNodeKind("FootnoteDefinition")`).

---

## Structures and API Reference

### 1. `FootnoteReference`

`FootnoteReference` represents an inline AST node referring to a footnote definition within the document text (e.g., `[^1]`).

#### Type Definition
```go
type FootnoteReference struct {
    gast.BaseInline

    Label    text.SingleLineValue
    Index    int
    RefIndex int
}
```

#### Fields
| Field | Type | Description |
| :--- | :--- | :--- |
| `BaseInline` | `gast.BaseInline` | Embedded base struct denoting this as an inline AST node. |
| `Label` | `text.SingleLineValue` | The text label assigned to the footnote reference (e.g., `"1"` for `[^1]`). |
| `Index` | `int` | The display index of the referenced `FootnoteDefinition`. Set by the footnote parser after registration. |
| `RefIndex` | `int` | The 0-based position of this reference relative to all references pointing to the same `FootnoteDefinition`. |

#### Functions and Methods

##### `NewFootnoteReference(label text.SingleLineValue) *FootnoteReference`
Constructs and initializes a new `FootnoteReference` AST node.
* **Parameters**: `label` (`text.SingleLineValue`) - The reference label.
* **Initial State**:
  * Sets `Label` to the provided `label`.
  * Initializes `Index` to `-1`.
  * Initializes `RefIndex` to `-1`.
  * Calls `n.Init(n)` to initialize the underlying Goldmark AST node structure.
* **Returns**: `*FootnoteReference`

##### `Kind() gast.NodeKind`
Returns the node kind.
* **Returns**: `KindFootnoteReference`

##### `Dump(_ []byte) *gast.NodeDump`
Implements the AST `Node.Dump` interface method for debugging and AST inspection.
* **Parameters**: `_ []byte` (source byte slice, unused).
* **Returns**: A `*gast.NodeDump` containing the node instance and a property map with keys `"Label"`, `"Index"`, and `"RefIndex"`.

---

### 2. `FootnoteDefinition`

`FootnoteDefinition` represents a block AST node defining the content associated with a footnote reference (e.g., `[^1]: Footnote content`).

#### Type Definition
```go
type FootnoteDefinition struct {
    gast.BaseBlock

    Label text.SingleLineValue
}
```

#### Fields
| Field | Type | Description |
| :--- | :--- | :--- |
| `BaseBlock` | `gast.BaseBlock` | Embedded base struct denoting this as a block-level AST node. |
| `Label` | `text.SingleLineValue` | The text label assigned to the footnote definition (e.g., `"1"` for `[^1]: ...`). |

#### Functions and Methods

##### `NewFootnoteDefinition(label text.SingleLineValue) *FootnoteDefinition`
Constructs and initializes a new `FootnoteDefinition` AST node.
* **Parameters**: `label` (`text.SingleLineValue`) - The definition label.
* **Initial State**:
  * Sets `Label` to the provided `label`.
  * Calls `n.Init(n)` to initialize the underlying Goldmark AST node structure.
* **Returns**: `*FootnoteDefinition`

##### `Kind() gast.NodeKind`
Returns the node kind.
* **Returns**: `KindFootnoteDefinition`

##### `Dump(_ []byte) *gast.NodeDump`
Implements the AST `Node.Dump` interface method for debugging and AST inspection.
* **Parameters**: `_ []byte` (source byte slice, unused).
* **Returns**: A `*gast.NodeDump` containing the node instance and a property map with key `"Label"`.

---

## Summary of Differences

| Feature | `FootnoteReference` | `FootnoteDefinition` |
| :--- | :--- | :--- |
| **AST Class** | Inline Node (`gast.BaseInline`) | Block Node (`gast.BaseBlock`) |
| **Node Kind** | `KindFootnoteReference` | `KindFootnoteDefinition` |
| **Default Field Initializations** | `Index: -1`, `RefIndex: -1` | None |
| **Dump Output Attributes** | `Label`, `Index`, `RefIndex` | `Label` |