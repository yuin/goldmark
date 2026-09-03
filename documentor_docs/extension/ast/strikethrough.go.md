# Technical Documentation: `extension/ast/strikethrough.go`

## Overview

The `extension/ast/strikethrough.go` file defines a custom Abstract Syntax Tree (AST) node for GitHub Flavored Markdown (GFM) strikethrough text. It relies on the Goldmark AST framework (`github.com/yuin/goldmark/v2/ast`) and registers a new inline node type (`Strikethrough`).

---

## Constants & Global Variables

### `KindStrikethrough`

```go
var KindStrikethrough = gast.NewNodeKind("Strikethrough")
```

* **Type**: `gast.NodeKind`
* **Description**: A unique node identifier registered with the string name `"Strikethrough"`. It is used by Goldmark to identify and differentiate `Strikethrough` AST nodes from other node types.

---

## Structs

### `Strikethrough`

```go
type Strikethrough struct {
	gast.BaseInline
}
```

* **Description**: Represents a strikethrough inline element within the AST.
* **Embedded Fields**:
  * `gast.BaseInline`: Embeds Goldmark's base inline node structure, providing standard inline AST functionality (child management, attributes, line/range tracking, etc.).

---

## Functions

### `NewStrikethrough`

```go
func NewStrikethrough() *Strikethrough
```

* **Description**: Constructor function that creates and initializes a new `Strikethrough` node.
* **Returns**: `*Strikethrough` — A pointer to the newly allocated and initialized `Strikethrough` AST node.
* **Implementation Details**:
  1. Allocates a new `Strikethrough` struct (`n := &Strikethrough{}`).
  2. Calls `n.Init(n)` (inherited from `gast.BaseInline`) to set up the node's internal state.
  3. Returns `n`.

---

## Methods

### `Kind`

```go
func (n *Strikethrough) Kind() gast.NodeKind
```

* **Description**: Implements the `gast.Node.Kind` interface method.
* **Parameters**: None (receiver `n *Strikethrough`).
* **Returns**: `gast.NodeKind` — Returns the global `KindStrikethrough` identifier.

---

### `Dump`

```go
func (n *Strikethrough) Dump(_ []byte) *gast.NodeDump
```

* **Description**: Implements the `gast.Node.Dump` interface method for debugging and inspecting the node representation.
* **Parameters**: 
  * `_ []byte`: The raw Markdown source buffer (unused in this implementation).
* **Returns**: `*gast.NodeDump` — Returns a node dump object constructed via `gast.NewNodeDump(n, nil)`.