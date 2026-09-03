# Documentation Guide: `ast/block.go`

## Overview

The `ast/block.go` file defines the **block-level Abstract Syntax Tree (AST) nodes** for a Markdown parser in Go. Block nodes represent structural elements of a Markdown document, such as documents, paragraphs, headings, code blocks, lists, list items, blockquotes, thematic breaks, HTML blocks, and link reference definitions.

This package provides:
- A base block structure (`BaseBlock`) with embedded segment optimization.
- Data structures and constructors for standard CommonMark block types.
- Helper enums, flags, and methods to query, manipulate, and dump node information.

---

## Constants and Internal Flags

The package uses bitwise flags stored in an internal `uint8` field (`flags`) on `BaseBlock`:

```go
const flagBlankPreviousLines = 1 << 0
const flagSingle = 1 << 1
```

*   `flagBlankPreviousLines`: Indicates whether the block is preceded by one or more blank lines.
*   `flagSingle`: Optimization flag indicating that the block currently holds only a single `textm.Segment` stored inline inside the `single` fixed array rather than allocating a slice.

---

## Primary Infrastructure

### `BaseBlock`

`BaseBlock` is the foundation for all block-level AST nodes. It embeds `BaseNode` and partially implements block node behavior.

```go
type BaseBlock struct {
    BaseNode
    source []textm.Segment
    single [1]textm.Segment
    flags  uint8
}
```

#### Fields
*   `source`: A slice of `textm.Segment` containing raw text ranges forming the block content.
*   `single`: A 1-element array used as an optimization to store a single text segment inline, avoiding memory allocations.
*   `flags`: Bitmask storing state such as `flagBlankPreviousLines` and `flagSingle`.

#### Memory Optimization Mechanism
`BaseBlock` optimizes memory usage for single-segment blocks:
1. When `AppendSource` is called for the first time, the segment is placed into `single[0]` and `flagSingle` is set. No slice is allocated.
2. If a second segment is added, `BaseBlock` initializes `source` as a slice with initial capacity `8`, copies the segment from `single[0]`, clears `flagSingle`, and appends the new segment.
3. `Source()` checks if `flagSingle` is set. If true, it returns a slice view of `single[:]`; otherwise, it returns `source`.
4. `SetSource(v)` directly replaces the `source` slice and clears the `flagSingle` flag.

#### Methods
*   `blockNode()`: Marker method implementing the `BlockNode` interface.
*   `HasBlankPreviousLines() bool`: Returns `true` if `flagBlankPreviousLines` is set.
*   `SetBlankPreviousLines(v bool)`: Sets or clears `flagBlankPreviousLines`.
*   `Source() []textm.Segment`: Returns all source text segments associated with this block.
*   `SetSource(v []textm.Segment)`: Overwrites the block's text segments with a new slice.
*   `AppendSource(seg textm.Segment)`: Appends a single source segment, handling single-segment inline storage optimization automatically.

---

## Concrete Block Node Implementations

### 1. `Document`

The root node of a parsed Markdown AST hierarchy.

```go
type Document struct {
    BaseBlock
    metadata map[string]any
}
```

*   **Node Kind**: `KindDocument`
*   **Constructor**: `NewDocument()` initializes the document, sets its self-reference, and marks `SetBlankPreviousLines(true)`.
*   **Methods**:
    *   `Pos() int`: Always returns `0`.
    *   `OwnerDocument() *Document`: Returns itself (`n`).
    *   `Dump(_ []byte) *NodeDump`: Returns node dump structure without extra parameters.
    *   `Metadata() map[string]any`: Returns the metadata map (lazy-initialized if `nil`).
    *   `SetMetadata(meta map[string]any)`: Copies the provided map into the document's metadata map using `maps.Copy`.
    *   `AddMeta(key string, value any)`: Sets a single metadata key-value pair.

---

### 2. `Paragraph`

Represents a standard text paragraph block.

```go
type Paragraph struct {
    BaseBlock
}
```

*   **Node Kind**: `KindParagraph`
*   **Constructor**: `NewParagraph()`
*   **Methods**:
    *   `Pos() int`: Returns `source[0].Start` position if available, or `-1` if source is empty.
    *   `Dump(_ []byte) *NodeDump`: Dumps node standard structure.
*   **Helper**:
    *   `IsParagraph(node Node) bool`: Returns `true` if the given node is non-nil and its kind is `KindParagraph`.

---

### 3. `Heading`

Represents headings (both ATX and Setext styles).

```go
type Heading struct {
    BaseBlock
    Level       int
    HeadingKind HeadingKind
}
```

#### Associated Type: `HeadingKind`
An integer enum indicating the heading style:
*   `HeadingKindATX` (1): ATX-style heading (e.g., `## Heading`).
*   `HeadingKindSetext` (2): Setext-style heading (underlined text).
*   `HeadingKind.String()`: Returns `"ATX"`, `"Setext"`, or `"Unknown"`.

#### Node Details
*   **Node Kind**: `KindHeading`
*   **Constructor**: `NewHeading(level int, kind HeadingKind)`
*   **Fields**:
    *   `Level`: Heading level (typically between `1` and `6`).
    *   `HeadingKind`: Indicates whether the heading is ATX or Setext.
*   **Methods**:
    *   `Dump(_ []byte)`: Includes `"Level"` and `"HeadingKind"` string in dump properties.

---

### 4. `ThematicBreak`

Represents horizontal rules / thematic breaks (e.g., `---`, `***`).

```go
type ThematicBreak struct {
    BaseBlock
}
```

*   **Node Kind**: `KindThematicBreak`
*   **Constructor**: `NewThematicBreak()`

---

### 5. `CodeBlock`

Represents indented or fenced code blocks.

```go
type CodeBlock struct {
    BaseBlock
    CodeBlockKind CodeBlockKind
    Info          textm.SingleLineValue
    Value         textm.Lines
}
```

#### Associated Type: `CodeBlockKind`
An integer enum indicating the code block style:
*   `CodeBlockKindIndented` (1): Code block indented by 4 spaces or a tab.
*   `CodeBlockKindFenced` (2): Code block enclosed in backticks or tildes.
*   `CodeBlockKind.String()`: Returns `"Indented"`, `"Fenced"`, or `"Unknown"`.

#### Node Details
*   **Node Kind**: `KindCodeBlock`
*   **Constructor**: `NewCodeBlock(kind CodeBlockKind, value textm.Lines, opts ...CodeBlockOption)`
*   **Fields**:
    *   `CodeBlockKind`: Style of code block (Indented or Fenced).
    *   `Info`: The info string attached to fenced code blocks (e.g., `go`). Empty for indented blocks.
    *   `Value`: The raw content lines stored in `textm.Lines`.
*   **Methods**:
    *   `Language(source []byte) (string, bool)`: Extracts the primary language identifier from the `Info` string up to the first space. Returns `("", false)` if `Info` is empty.
    *   `Dump(_ []byte)`: Includes `"CodeBlockKind"`, `"Value"`, and optionally `"Info"` if non-empty.

#### Options Pattern
*   `CodeBlockOption`: Interface accepting `setCodeBlockOption(*CodeBlock)`.
*   `WithCodeBlockInfo(info textm.SingleLineValue)`: Functional option setting the `Info` field on a `CodeBlock`.

---

### 6. `Blockquote`

Represents blockquote elements (e.g., `> quote`).

```go
type Blockquote struct {
    BaseBlock
}
```

*   **Node Kind**: `KindBlockquote`
*   **Constructor**: `NewBlockquote()`

---

### 7. `List`

Represents ordered or unordered list structures.

```go
type List struct {
    BaseBlock
    Marker  byte
    IsTight bool
    Start   int
}
```

*   **Node Kind**: `KindList`
*   **Constructor**: `NewList(marker byte)` (defaults `IsTight` to `true`).
*   **Fields**:
    *   `Marker`: Bullet or delimiter byte (e.g., `'-'`, `'+'`, `'*'`, `'.'`. `')'`).
    *   `IsTight`: Boolean indicating if list items are tightly spaced.
    *   `Start`: The starting sequence number if ordered (0 if unordered).
*   **Methods**:
    *   `IsOrdered() bool`: Returns `true` if `Marker` is `'.'` or `')'`.
    *   `CanContinue(marker byte, isOrdered bool) bool`: Evaluates if a list item with the given marker and ordered status can append to this list.
    *   `Dump(_ []byte)`: Dumps `"Ordered"`, `"Marker"`, `"Tight"`, and `"Start"` (if ordered).

---

### 8. `ListItem`

Represents an individual item inside a `List`.

```go
type ListItem struct {
    BaseBlock
    offset int
}
```

*   **Node Kind**: `KindListItem`
*   **Constructor**: `NewListItem()`
*   **Fields & Methods**:
    *   `offset`: Internal parser line indentation offset.
    *   `Offset() int`: Gets the content indentation offset.
    *   `SetOffset(v int)`: Sets the content indentation offset during parsing.

---

### 9. `HTMLBlock`

Represents raw HTML blocks in Markdown text.

```go
type HTMLBlock struct {
    BaseBlock
    HTMLBlockKind HTMLBlockKind
    Value         textm.Lines
}
```

#### Associated Type: `HTMLBlockKind`
Integer enum matching CommonMark HTML block types 1 through 7:
*   `HTMLBlockKind1` through `HTMLBlockKind7`
*   `HTMLBlockKind.String()`: Returns `"Kind1"` through `"Kind7"`, or `"Unknown"`.

#### Node Details
*   **Node Kind**: `KindHTMLBlock`
*   **Constructor**: `NewHTMLBlock(kind HTMLBlockKind)`
*   **Fields**:
    *   `HTMLBlockKind`: Type identifier per CommonMark specification.
    *   `Value`: Raw lines holding the HTML content.
*   **Methods**:
    *   `Dump(_ []byte)`: Includes `"HTMLBlockKind"` and `"Value"`.

---

### 10. `LinkReferenceDefinition`

Represents Markdown link reference definitions (e.g., `[label]: /url "title"`).

```go
type LinkReferenceDefinition struct {
    BaseBlock
    Label       textm.MultiLineValue
    Destination textm.SingleLineValue
    Title       textm.MultiLineValue
}
```

*   **Node Kind**: `KindLinkReferenceDefinition`
*   **Constructor**: `NewLinkReferenceDefinition(label textm.MultiLineValue, destination textm.SingleLineValue, opts ...LinkReferenceDefinitionOption)`
*   **Fields**:
    *   `Label`: Label segment of the reference definition.
    *   `Destination`: Target URL or destination segment.
    *   `Title`: Optional title text segment.
*   **Options Pattern**:
    *   `LinkReferenceDefinitionOption`: Interface accepting `setLinkReferenceDefinitionOption(*LinkReferenceDefinition)`.
    *   Option setting method implemented on `*linkTitle` to set the `Title` field.
*   **Methods**:
    *   `Dump(_ []byte)`: Dumps `"Label"`, `"Destination"`, and `"Title"`.

---

## Summary Table of Block Node Kinds

| Struct | Node Kind Variable | Key Features / Fields |
| :--- | :--- | :--- |
| `Document` | `KindDocument` | Root AST node; contains `metadata` map. |
| `Paragraph` | `KindParagraph` | Text paragraph; position aware. |
| `Heading` | `KindHeading` | `Level` (1-6), `HeadingKind` (ATX/Setext). |
| `ThematicBreak` | `KindThematicBreak` | Represents horizontal separators. |
| `CodeBlock` | `KindCodeBlock` | `CodeBlockKind` (Indented/Fenced), `Info`, language extraction helper. |
| `Blockquote` | `KindBlockquote` | Container for quoted blocks. |
| `List` | `KindList` | `Marker`, `IsTight`, `Start`, list continuation check. |
| `ListItem` | `KindListItem` | Individual item inside a list; carries parser `offset`. |
| `HTMLBlock` | `KindHTMLBlock` | Raw HTML; categorized by `HTMLBlockKind` (Types 1-7). |
| `LinkReferenceDefinition` | `KindLinkReferenceDefinition` | `Label`, `Destination`, and `Title` definition mapping. |