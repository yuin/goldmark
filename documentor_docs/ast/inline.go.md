# AST Inline Elements Documentation (`ast/inline.go`)

## Overview

The `ast/inline.go` file defines the Abstract Syntax Tree (AST) node structures and types representing **inline elements** in Markdown. Inline elements are parsed segments that appear within block-level containers (such as paragraphs or headings). 

The file includes nodes for:
* Plain text (`Text`) with line-break indicators
* Code spans (`CodeSpan`)
* Styled text (`Emphasis`, `Strong`)
* Links (`Link`) and Images (`Image`)
* Autolinks (`AutoLink`)
* Raw inline HTML (`RawHTML`)
* Shared helper structs, enum kinds, and functional options for link/autolink construction

---

## Key Components & Structures

### 1. `BaseInline`

`BaseInline` serves as the foundational struct embedded by all inline AST nodes. It embeds `BaseNode` and implements the `InlineNode` marker interface.

```go
type BaseInline struct {
    BaseNode
}
```

* **`inlineNode()`**: Unexported marker method used to designate the struct as an inline node.

---

### 2. `Text` Node

`Text` represents simple text within Markdown. It tracks source positions/strings and handles soft/hard line breaks through bitwise flags.

#### Fields
* `Value` (`textm.SingleLineValue`): Holds the single-line text value (either a source range position or owned string memory).
* `flags` (`uint8`): Internal bitmask for tracking trailing line break types.

#### Flags & Constants
* `textSoftLineBreak` (`1 << 0`): Flag set when the text ends with a soft line break.
* `textHardLineBreak` (`1 << 1`): Flag set when the text ends with a hard line break.

#### Node Identification
* `KindText`: The `NodeKind` identifier (`NewNodeKind("Text")`).
* `Kind() NodeKind`: Returns `KindText`.

#### Constructors & Methods
* **`NewText(value textm.SingleLineValue) *Text`**: Creates and initializes a `Text` node.
* **`Pos() int`**: Returns the starting position in source code. Returns `-1` if `Value` is an owned string rather than a position index.
* **`SoftLineBreak() bool` / `SetSoftLineBreak(v bool)`**: Getter and setter for the soft line break flag.
* **`HardLineBreak() bool` / `SetHardLineBreak(v bool)`**: Getter and setter for the hard line break flag.
* **`Dump(_ []byte) *NodeDump`**: Dumps node details including value and flag strings (`SoftLineBreak`, `HardLineBreak`).

---

### 3. `CodeSpan` Node

`CodeSpan` represents inline code content (e.g., `` `code` ``).

#### Fields
* `Value` (`textm.Value`): Holds the raw code span content, normalized to a single line (newlines replaced with spaces as per CommonMark requirements).

#### Node Identification
* `KindCodeSpan`: The `NodeKind` identifier (`NewNodeKind("CodeSpan")`).
* `Kind() NodeKind`: Returns `KindCodeSpan`.

#### Constructors & Methods
* **`NewCodeSpan(value textm.Value) *CodeSpan`**: Creates and initializes a `CodeSpan` node.
* **`IsBlank(source []byte) bool`**: Checks whether the span content consists entirely of whitespace using `util.IsBlank`.
* **`Dump(_ []byte) *NodeDump`**: Returns node dump info containing `Value`.

---

### 4. `Emphasis` and `Strong` Nodes

These structural nodes wrap child inline nodes to represent styled text (e.g., `*emphasis*` or `**strong**`).

#### `Emphasis`
* **Node Kind**: `KindEmphasis` (`NewNodeKind("Emphasis")`)
* **`NewEmphasis() *Emphasis`**: Constructor.
* **`Dump(_ []byte) *NodeDump`**: Returns `NodeDump` with `nil` parameters.

#### `Strong`
* **Node Kind**: `KindStrong` (`NewNodeKind("Strong")`)
* **`NewStrong() *Strong`**: Constructor.
* **`Dump(_ []byte) *NodeDump`**: Returns `NodeDump` with `nil` parameters.

---

### 5. Links and Images Infrastructure

Links and images share standard fields, option patterns, and reference kinds via `baseLink`.

#### `ReferenceLinkKind` (Enum)
Represents the flavor of a reference link:
* `ReferenceLinkKindFull` (`1`): Full reference, e.g., `[foo][bar]`
* `ReferenceLinkKindCollapsed` (`2`): Collapsed reference, e.g., `[foo][]`
* `ReferenceLinkKindShortcut` (`3`): Shortcut reference, e.g., `[foo]`

Methods:
* **`String() string`**: Returns string representation (`"Full"`, `"Collapsed"`, `"Shortcut"`, or `"Unknown(n)"`).

#### `ReferenceLink` Struct
Defines the reference target metadata:
* `ReferenceLinkKind` (`ReferenceLinkKind`)
* `Value` (`textm.MultiLineValue`)
* **`NewReferenceLink(kind ReferenceLinkKind, value textm.MultiLineValue) *ReferenceLink`**: Constructor.

#### `baseLink` Struct
Internal embedded base for `Link` and `Image`.
* `Destination` (`textm.SingleLineValue`): Target URL or destination path.
* `Title` (`textm.MultiLineValue`): Link title attribute text.
* `Reference` (`*ReferenceLink`): Pointer to a reference link descriptor if applicable; otherwise `nil`.

#### Link Options
* **`LinkOption` Interface**: Implemented by options that configure a `baseLink`.
* **`WithLinkTitle(title textm.MultiLineValue)`**: Returns a functional option to assign the link title (also satisfies `LinkReferenceDefinitionOption`).
* **`WithLinkReference(kind ReferenceLinkKind, value textm.MultiLineValue)`**: Returns a functional option to configure a reference link.

---

### 6. `Link` Node

Represents a standard hypertext link (`[text](url)` or `[text][ref]`).

#### Structure
Embeds `baseLink`.

#### Node Identification
* `KindLink`: The `NodeKind` identifier (`NewNodeKind("Link")`).
* `Kind() NodeKind`: Returns `KindLink`.

#### Constructors & Methods
* **`NewLink(destination textm.SingleLineValue, opts ...LinkOption) *Link`**: Instantiates a new link with the provided destination and options.
* **`Dump(_ []byte) *NodeDump`**: Dumps `Destination`, `Title` (if non-empty), and `Reference` (if present).

---

### 7. `Image` Node

Represents an image element (`![alt](url)`).

#### Structure
Embeds `baseLink`.

#### Node Identification
* `KindImage`: The `NodeKind` identifier (`NewNodeKind("Image")`).
* `Kind() NodeKind`: Returns `KindImage`.

#### Constructors & Methods
* **`NewImage(destination textm.SingleLineValue, opts ...LinkOption) *Image`**: Instantiates a new image node with the provided destination and options.
* **`Dump(_ []byte) *NodeDump`**: Dumps `Destination`, `Title` (if non-empty), and `Reference` (if present).

---

### 8. `AutoLink` Node

Represents an automatic link, such as `<https://example.com>` or `<user@example.com>`.

#### Fields
* `Destination` (`textm.SingleLineValue`): The target URL/href (includes `mailto:` for email links).
* `Label` (`textm.SingleLineValue`): Display text shown within the link element.
* `Text` (`textm.SingleLineValue`): Original text from source including delimiters (e.g., `<` and `>`).

#### Options
* **`AutoLinkOption` Interface**: Option interface for configuring `AutoLink`.
* **`WithAutoLinkText(text textm.SingleLineValue)`**: Option to set original source text `Text`.

#### Node Identification
* `KindAutoLink`: The `NodeKind` identifier (`NewNodeKind("AutoLink")`).
* `Kind() NodeKind`: Returns `KindAutoLink`.

#### Constructors & Methods
* **`NewAutoLink(destination, label textm.SingleLineValue, opts ...AutoLinkOption) *AutoLink`**: Constructor.
* **`Dump(_ []byte) *NodeDump`**: Dumps `Destination`, `Label`, and `Text`.

---

### 9. `RawHTML` Node

Represents inline raw HTML tags or snippets (e.g., `<span>text</span>`).

#### Fields
* `Value` (`textm.MultiLineValue`): Contains the raw HTML content.

#### Node Identification
* `KindRawHTML`: The `NodeKind` identifier (`NewNodeKind("RawHTML")`).
* `Kind() NodeKind`: Returns `KindRawHTML`.

#### Constructors & Methods
* **`NewRawHTML(value textm.MultiLineValue) *RawHTML`**: Creates a `RawHTML` node.
* **`Dump(_ []byte) *NodeDump`**: Returns dump information containing `Value`.

---

## Summary of AST Node Kinds

| Struct | NodeKind Variable | String Representation |
| :--- | :--- | :--- |
| `Text` | `KindText` | `"Text"` |
| `CodeSpan` | `KindCodeSpan` | `"CodeSpan"` |
| `Emphasis` | `KindEmphasis` | `"Emphasis"` |
| `Strong` | `KindStrong` | `"Strong"` |
| `Link` | `KindLink` | `"Link"` |
| `Image` | `KindImage` | `"Image"` |
| `AutoLink` | `KindAutoLink` | `"AutoLink"` |
| `RawHTML` | `KindRawHTML` | `"RawHTML"` |