# Technical Documentation Guide: `ast/ast.go`

## Overview

The `ast/ast.go` file is part of the `ast` package in Goldmark. It defines the foundational Abstract Syntax Tree (AST) architecture used to represent Markdown elements. 

This file includes:
- Core AST node interfaces (`Node`, `BlockNode`, `InlineNode`).
- A base implementation (`BaseNode`) providing doubly-linked list tree structure and attribute management.
- Dynamic node type tracking (`NodeKind`).
- AST tree traversal utilities (`Walk`).
- Tree inspection and debugging tools (`NodeDump`, `PrettyPrint`).
- A helper function (`N`) for constructing AST node hierarchies.

---

## 1. Node Types and Kinds

### `NodeKind`
```go
type NodeKind int
```
An integer type that uniquely identifies the concrete type of an AST node.

- **`String() string`**: Returns the human-readable string representation of the `NodeKind`.
- **`CurrentKindValue`** (`NodeKind`): Global counter tracking the current kind value. 
  > **Warning:** Callers **MUST NOT** use or modify this variable directly. It is exported solely for goldmark internal package usage.
- **`NewNodeKind(name string) NodeKind`**: Registers a new node kind with a given name, increments `CurrentKindValue`, and returns the newly generated `NodeKind`.

### `Attribute`
```go
type Attribute struct {
    Name  string
    Value textm.MultiLineValue
}
```
Represents a generic key-value attribute associated with an AST node, where the value is stored as a `textm.MultiLineValue`.

---

## 2. AST Interfaces

### `Node`
The primary interface that every AST node must implement. It defines methods for node positioning, tree navigation, parent-child manipulation, document ownership, dumping, and attribute management.

```go
type Node interface {
    Kind() NodeKind
    Pos() int
    SetPos(v int)

    // Tree Navigation
    NextSibling() Node
    PreviousSibling() Node
    Parent() Node
    SetParent(Node)
    SetPreviousSibling(Node)
    SetNextSibling(Node)

    // Child Querying and Iteration
    HasChildren() bool
    ChildCount() int
    Children() iter.Seq[Node]
    FirstChild() Node
    LastChild() Node

    // Tree Modification
    AppendChild(child Node)
    RemoveChild(child Node)
    RemoveChildren()
    ReplaceChild(target, insertee Node)
    InsertBefore(target, insertee Node)
    InsertAfter(target, insertee Node)

    // Context & Debugging
    OwnerDocument() *Document
    Dump(source []byte) *NodeDump

    // Attribute Management
    SetAttribute(name string, value textm.MultiLineValue)
    Attribute(name string) (textm.MultiLineValue, bool)
    Attributes() []Attribute
    RemoveAttributes()
}
```

### `BlockNode`
```go
type BlockNode interface {
    Node
    blockNode()
    HasBlankPreviousLines() bool
    SetBlankPreviousLines(v bool)
    Source() []textm.Segment
    SetSource([]textm.Segment)
    AppendSource(textm.Segment)
}
```
Extends `Node` to represent block-level elements (e.g., Headings, Paragraphs, Code Blocks).
- Manages raw source text segments (`Source()`, `SetSource()`, `AppendSource()`). Block elements parsed into inline elements store their input source segments here; elements that do not parse inline content leave this empty.
- Tracks blank line spacing via `HasBlankPreviousLines()` and `SetBlankPreviousLines()`.

### `InlineNode`
```go
type InlineNode interface {
    Node
    inlineNode()
}
```
Extends `Node` to distinguish inline-level elements (e.g., Emphasis, Text, Links).

---

## 3. Base Implementation: `BaseNode`

`BaseNode` provides a default implementation of the core `Node` methods using a doubly-linked list for child nodes. Node structs embed `BaseNode` to inherit standard tree behavior.

### Fields
- `self Node`: Pointer to the concrete node interface.
- `firstChild Node`, `lastChild Node`: Boundary pointers for the node's children list.
- `parent Node`: Reference to the parent node.
- `next Node`, `prev Node`: Pointers to neighboring sibling nodes.
- `childCount int`: Total number of immediate children.
- `attributes []Attribute`: Key-value pair attributes attached to the node.
- `pos int`: Source text offset (-1 if position is undefined).

### Key Operations

#### Initialization
- **`Init(self Node)`**: Must be called in concrete node constructors. Assigns the `self` self-reference pointer and sets default `pos` to `-1`.

#### Tree Structure Management
- **`AppendChild(v Node)`**: Detaches `v` from its current parent (if any) and appends it as the last child.
- **`RemoveChild(v Node)`**: Disconnects `v` from `BaseNode`'s children list if `v.Parent() == n.self`. Resets sibling/parent pointers on `v`.
- **`RemoveChildren()`**: Iterates through all children, unlinks them, and clears list metadata.
- **`InsertBefore(target, insertee Node)`**: Inserts `insertee` prior to `target`. If `target` is `nil`, appends `insertee` to the end.
- **`InsertAfter(target, insertee Node)`**: Inserts `insertee` after `target` by delegating to `InsertBefore(target.NextSibling(), insertee)`.
- **`ReplaceChild(target, insertee Node)`**: Inserts `insertee` before `target` and removes `target`.
- **`Children() iter.Seq[Node]`**: Returns an iterator (`iter.Seq[Node]`) over the node's children.

#### Document & Attribute Querying
- **`OwnerDocument() *Document`**: Traverses up parent nodes until reaching the root. If the root node is a `*Document`, returns it; otherwise returns `nil`.
- **`SetAttribute(name string, value textm.MultiLineValue)`**: Updates an attribute value in-place if `name` exists, or appends a new `Attribute`.
- **`Attribute(name string) (textm.MultiLineValue, bool)`**: Lookups an attribute by `Name`.
- **`Attributes() []Attribute`**: Returns the internal slice of attributes.
- **`RemoveAttributes()`**: Clears all attributes by setting the internal slice to `nil`.

---

## 4. AST Dump and Pretty Printing

The package provides AST inspection tools primarily aimed at debugging.

### `NodeDump`
```go
type NodeDump struct {
    Node       Node
    Properties map[string]any
}
```
Constructed via `NewNodeDump(node Node, properties map[string]any)`.
- **`Children(source []byte) iter.Seq[*NodeDump]`**: Yields `*NodeDump` objects for all child nodes.

### Pretty Printing
Outputs a formatted text representation of a node tree to an `io.Writer`.

#### Options
- **`WithLevel(level int) PrettyPrintOption`**: Sets the starting indentation level (4 spaces per level).
- **`WithSource(include bool) PrettyPrintOption`**: Configures whether source text contents are extracted and included in the output for `BlockNode` instances.

#### Usage Method
```go
func (d *NodeDump) PrettyPrint(w io.Writer, source []byte, opts ...PrettyPrintOption) error
```
> **Note:** The `PrettyPrint` output format is intended strictly for debugging purposes and is subject to change.

---

## 5. Tree Traversal (`Walk`)

The `Walk` system allows depth-first traversal of the AST using a callback function.

### `WalkStatus`
An enumeration returned by a `Walker` to control traversal flow:
- **`WalkContinue`**: Continue normal depth-first traversal.
- **`WalkSkipChildren`**: Skip traversal of the current node's children.
- **`WalkStop`**: Abort tree walking immediately.

### `Walker` Callback
```go
type Walker func(n Node, entering bool) (WalkStatus, error)
```
- Called twice for each visited node:
  1. **Pre-order (`entering = true`)**: Before walking child nodes.
  2. **Post-order (`entering = false`)**: After child nodes have been walked.
- If `Walker` returns an error or `WalkStop`, traversal halts immediately and propagates the error.

### Execution
```go
func Walk(n Node, walker Walker) error
```
Initiates recursive depth-first traversal starting at node `n`.

---

## 6. Tree Construction Helper (`N`)

```go
func N(node Node, children ...any) Node
```
A convenience helper for building AST structures containing children.

### Supported Child Types:
1. **`Node`**: Directly appended to `node` via `AppendChild`.
2. **`string`**: 
   - If the string contains no newlines (`\n`), it creates a single `Text` node.
   - If the string contains newlines (`\n`), it splits the content by line boundaries and creates multiple `Text` nodes.
3. **Other types**: Panics if a child is neither a `Node` nor a `string`.

### Example Logic
```go
// Adds decoded strings or child nodes to a parent element
parent := N(NewParagraph(), "Hello World")
```