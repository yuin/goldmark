# Technical Documentation: `extension/ast/definition_list.go`

## Overview

The `extension/ast/definition_list.go` file defines the Abstract Syntax Tree (AST) nodes used to represent Markdown definition lists (specifically adhering to the PHPMarkdownExtra syntax specification). It extends Goldmark's AST (`github.com/yuin/goldmark/v2/ast`) by implementing three block-level AST nodes:

1. **`DefinitionList`**: The outer container node for a definition list.
2. **`DefinitionTerm`**: The node representing a term being defined.
3. **`DefinitionDescription`**: The node representing the description/definition associated with a term.

---

## Constants & Variables

The package declares three package-level variables representing the unique `NodeKind` identifier for each AST node defined in this file:

* `var KindDefinitionList = gast.NewNodeKind("DefinitionList")`
* `var KindDefinitionTerm = gast.NewNodeKind("DefinitionTerm")`
* `var KindDefinitionDescription = gast.NewNodeKind("DefinitionDescription")`

---

## Structs and Methods

### 1. `DefinitionList`

`DefinitionList` is a container block that holds `DefinitionTerm` and `DefinitionDescription` child nodes.

#### Fields
* `gast.BaseBlock`: Embedded Goldmark base block structure providing standard AST block functionality.
* `offset` (`int`): Internal parser state holding the indentation offset of the definition list.
* `temporaryParagraph` (`*gast.Paragraph`): Internal parser state tracking a temporary paragraph node during list parsing.

#### Constructors
* `NewDefinitionList() *DefinitionList`: Allocates, initializes, and returns a new `DefinitionList` AST node.

#### Methods
* **`Offset() int`**: Returns the current parser-internal indentation offset.
* **`SetOffset(v int)`**: Sets the parser-internal indentation offset.
* **`TemporaryParagraph() *gast.Paragraph`**: Returns the associated temporary paragraph node.
* **`SetTemporaryParagraph(p *gast.Paragraph)`**: Sets the associated temporary paragraph node.
* **`Kind() gast.NodeKind`**: Returns `KindDefinitionList`.
* **`Dump(source []byte) *gast.NodeDump`**: Implements `gast.Node.Dump`. Returns a `gast.NodeDump` representing the node structure.
* **`Pos() int`**: Implements `gast.Node.Pos`. Returns the source position of the node's first child if a child exists; otherwise, returns `-1`.

---

### 2. `DefinitionTerm`

`DefinitionTerm` represents a term in a definition list.

#### Fields
* `gast.BaseBlock`: Embedded Goldmark base block structure.

#### Constructors
* `NewDefinitionTerm() *DefinitionTerm`: Allocates, initializes, and returns a new `DefinitionTerm` AST node.

#### Methods
* **`Kind() gast.NodeKind`**: Returns `KindDefinitionTerm`.
* **`Dump(source []byte) *gast.NodeDump`**: Implements `gast.Node.Dump`. Returns a `gast.NodeDump` representing the node structure.
* **`Pos() int`**: Implements `gast.Node.Pos`. Returns the start source position (`Source()[0].Start`) if source slices are present; otherwise, returns `-1`.

---

### 3. `DefinitionDescription`

`DefinitionDescription` represents a description block in a definition list.

#### Fields
* `gast.BaseBlock`: Embedded Goldmark base block structure.
* `IsTight` (`bool`): Indicates whether the definition description is formatted as a tight list element (i.e., without extra blank lines/paragraph spacing).

#### Constructors
* `NewDefinitionDescription() *DefinitionDescription`: Allocates, initializes, and returns a new `DefinitionDescription` AST node.

#### Methods
* **`Kind() gast.NodeKind`**: Returns `KindDefinitionDescription`.
* **`Dump(source []byte) *gast.NodeDump`**: Implements `gast.Node.Dump`. Returns a `gast.NodeDump` representing the node structure.

---

## Summary of Node Responsibilities

| Struct | Node Kind | Key Attributes / Parser State | Source Position Resolution (`Pos()`) |
| :--- | :--- | :--- | :--- |
| **`DefinitionList`** | `KindDefinitionList` | `offset`, `temporaryParagraph` | Returns position of `FirstChild()`, or `-1` if empty. |
| **`DefinitionTerm`** | `KindDefinitionTerm` | None (embeds `BaseBlock`) | Returns `n.Source()[0].Start`, or `-1` if no source. |
| **`DefinitionDescription`** | `KindDefinitionDescription` | `IsTight` | Inherited from `BaseBlock`. |