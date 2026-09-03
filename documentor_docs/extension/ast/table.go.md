# Technical Documentation: `extension/ast/table.go`

## Overview

The `extension/ast/table.go` file defines the Abstract Syntax Tree (AST) nodes and alignment types required to represent GitHub Flavored Markdown (GFM) tables within the Goldmark Markdown parser (`github.com/yuin/goldmark/v2/ast`).

It introduces an `Alignment` enumeration for table cells and five distinct AST block nodes:
* `Table`
* `TableHeader`
* `TableBody`
* `TableRow`
* `TableCell`

---

## Constants & Alignment Type

### `Alignment`

`Alignment` is an integer type representing the text alignment inside a table cell.

```go
type Alignment int
```

#### Constants

* **`AlignLeft`** (`iota + 1`): Text is left-justified.
* **`AlignRight`**: Text is right-justified.
* **`AlignCenter`**: Text is centered.
* **`AlignNone`**: Default alignment behavior (no explicit alignment set).

#### Methods

* **`func (a Alignment) String() string`**
  Returns the string representation of an `Alignment` value:
  * `AlignLeft` $\rightarrow$ `"left"`
  * `AlignRight` $\rightarrow$ `"right"`
  * `AlignCenter` $\rightarrow$ `"center"`
  * `AlignNone` $\rightarrow$ `"none"`
  * Unrecognized values $\rightarrow$ `""`

---

## AST Nodes

All AST nodes in this file embed `gast.BaseBlock` from the standard Goldmark AST package and implement the `gast.Node` interface through `Kind()` and `Dump()` methods.

---

### 1. `Table`

Represents an entire Markdown table block.

#### Definition
```go
type Table struct {
    gast.BaseBlock
}
```

#### Package-Level Variables
* **`KindTable`**: A `gast.NodeKind` registered with the name `"Table"`.

#### Functions & Methods
* **`NewTable() *Table`**
  Constructs and initializes a new `Table` AST node.
* **`Kind() gast.NodeKind`**
  Returns `KindTable`.
* **`Dump(source []byte) *gast.NodeDump`**
  Returns the node dump for debugging without additional attributes.

---

### 2. `TableHeader`

Represents the header section of a table (`<thead>` equivalent in HTML).

#### Definition
```go
type TableHeader struct {
    gast.BaseBlock
}
```

#### Package-Level Variables
* **`KindTableHeader`**: A `gast.NodeKind` registered with the name `"TableHeader"`.

#### Functions & Methods
* **`NewTableHeader() *TableHeader`**
  Constructs and initializes a new `TableHeader` AST node.
* **`Kind() gast.NodeKind`**
  Returns `KindTableHeader`.
* **`Dump(source []byte) *gast.NodeDump`**
  Returns the node dump for debugging without additional attributes.

---

### 3. `TableBody`

Represents the main body section of a table (`<tbody>` equivalent in HTML).

#### Definition
```go
type TableBody struct {
    gast.BaseBlock
}
```

#### Package-Level Variables
* **`KindTableBody`**: A `gast.NodeKind` registered with the name `"TableBody"`.

#### Functions & Methods
* **`NewTableBody() *TableBody`**
  Constructs and initializes a new `TableBody` AST node.
* **`Kind() gast.NodeKind`**
  Returns `KindTableBody`.
* **`Dump(source []byte) *gast.NodeDump`**
  Returns the node dump for debugging without additional attributes.

---

### 4. `TableRow`

Represents a single table row (`<tr>` equivalent in HTML), located either inside a `TableHeader` or `TableBody`.

#### Definition
```go
type TableRow struct {
    gast.BaseBlock
}
```

#### Package-Level Variables
* **`KindTableRow`**: A `gast.NodeKind` registered with the name `"TableRow"`.

#### Functions & Methods
* **`NewTableRow() *TableRow`**
  Constructs and initializes a new `TableRow` AST node.
* **`Kind() gast.NodeKind`**
  Returns `KindTableRow`.
* **`Dump(source []byte) *gast.NodeDump`**
  Returns the node dump for debugging without additional attributes.

---

### 5. `TableCell`

Represents an individual table cell (`<td>` or `<th>` equivalent in HTML) inside a `TableRow`.

#### Definition
```go
type TableCell struct {
    gast.BaseBlock
    Alignment Alignment
}
```

#### Struct Fields
* **`Alignment`** (`Alignment`): Specifies the text alignment of the cell (left, right, center, or none).

#### Package-Level Variables
* **`KindTableCell`**: A `gast.NodeKind` registered with the name `"TableCell"`.

#### Functions & Methods
* **`NewTableCell(alignment Alignment) *TableCell`**
  Constructs and initializes a new `TableCell` AST node with the specified alignment.
* **`Kind() gast.NodeKind`**
  Returns `KindTableCell`.
* **`Dump(source []byte) *gast.NodeDump`**
  Returns a `gast.NodeDump` containing a map with the cell's `"Alignment"` string representation (`"left"`, `"right"`, `"center"`, or `"none"`).

---

## Component Summary

| Node Struct | Constructor | `NodeKind` Variable | Associated Fields |
| :--- | :--- | :--- | :--- |
| `Table` | `NewTable()` | `KindTable` | `gast.BaseBlock` |
| `TableHeader` | `NewTableHeader()` | `KindTableHeader` | `gast.BaseBlock` |
| `TableBody` | `NewTableBody()` | `KindTableBody` | `gast.BaseBlock` |
| `TableRow` | `NewTableRow()` | `KindTableRow` | `gast.BaseBlock` |
| `TableCell` | `NewTableCell(alignment Alignment)` | `KindTableCell` | `gast.BaseBlock`, `Alignment` |