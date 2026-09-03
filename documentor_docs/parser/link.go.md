# Technical Documentation: `parser/link.go`

## Overview

The `parser/link.go` file implements inline parsing logic for Markdown links and images according to the CommonMark specification. It defines the `linkParser` structure, which integrates with Goldmark's AST (Abstract Syntax Tree) and text parsing pipeline to parse:

- Inline links: `[text](destination "title")`
- Full reference links: `[text][label]`
- Collapsed reference links: `[text][]`
- Shortcut reference links: `[text]`
- Images (inline or reference): `![text](...)` or `![text][...]`

It also manages tracking state across nested bracket contexts using a doubly-linked list (`linkLabelState`) and a context-based stack system for delimiter boundaries (`linkBottom`).

---

## Key Context Keys

The parser uses two internal context keys stored in the `Context` (`pc`):

1. **`linkLabelStateKey`**: Context key holding the current active `*linkLabelState` linked list for tracking opening square brackets `[` and `![`.
2. **`linkBottom`**: Context key holding delimiter nodes (as an `ast.Node` or `[]ast.Node`) that mark delimiter boundaries prior to processing link labels.

---

## Primary Structures

### `linkLabelState`

`linkLabelState` represents an open link or image bracket state within the parsing context. It embeds `ast.BaseInline` and functions as a node in a doubly-linked list.

#### Fields
- **`value` (`text.Segment`)**: The text segment corresponding to the bracket(s) (`[` or `![`).
- **`IsImage` (`bool`)**: Set to `true` if the state represents an image open `![`, or `false` for a standard link `[`.
- **`Prev`, `Next` (`*linkLabelState`)**: Links to previous and next label states in the chain.
- **`First`, `Last` (`*linkLabelState`)**: Pointers to the head and tail of the label state list maintained at the list head.

#### Node Implementation Methods
- **`Kind()`**: Returns `ast.NodeKind` with the name `"LinkLabelState"`.
- **`Dump(source []byte)`**: Returns node dump containing the `"IsImage"` property for debugging AST output.

---

### `linkParser`

`linkParser` implements the `InlineParser` interface to handle link and image bracket parsing.

- **`NewLinkParser() InlineParser`**: Instantiates and returns the default `InlineParser` implementation (`defaultLinkParser`).
- **`Trigger()`**: Returns `[]byte{'!', '[', ']'}`—the character triggers that cause Goldmark to execute this parser.

---

## Component Functions and Workflow

### State Management Functions

- **`newLinkLabelState(segment text.Segment, isImage bool) *linkLabelState`**
  Constructs a new `linkLabelState` tracking a specific text segment and image flag.

- **`linkLabelStateLength(v *linkLabelState) int`**
  Calculates the character length between `v.First.value.Start` and `v.Last.value.Stop`.

- **`pushLinkLabelState(pc Context, v *linkLabelState)`**
  Pushes a new state node onto the `linkLabelStateKey` list in the parser context. If the list is empty, `v` becomes the head and updates its `First` and `Last` fields. Otherwise, `v` is appended to the tail (`Last`).

- **`removeLinkLabelState(pc Context, d *linkLabelState)`**
  Removes node `d` from the doubly-linked list maintained in `pc`. Adjusts `Next`, `Prev`, `First`, and `Last` pointers accordingly, clearing all pointers on `d`.

- **`pushLinkBottom(pc Context)` / `popLinkBottom(pc Context) ast.Node`**
  Manages a stack of delimiter bottom nodes in the `linkBottom` context value.
  - `pushLinkBottom`: Appends the context's last delimiter (`pc.LastDelimiter()`) to the bottom list/slice.
  - `popLinkBottom`: Pops and returns the most recent delimiter node added.

---

### Parsing Logic & Methods

#### `Parse(parent ast.Node, block text.Reader, pc Context) ast.Node`

This is the entry point called when encountering `!`, `[`, or `]`.

1. **Trigger `!` (Image Open):**
   Checks if followed by `[`. If so, advances block reader by 1, calls `pushLinkBottom(pc)`, and invokes `processLinkLabelOpen` with `isImage = true`.

2. **Trigger `[` (Link Open):**
   Pushes link bottom via `pushLinkBottom(pc)` and calls `processLinkLabelOpen` with `isImage = false`.

3. **Trigger `]` (Closing Bracket):**
   - Retrieves the last `linkLabelState` node from `linkLabelStateKey`.
   - Validates length limit: CommonMark limits link labels to at most 999 characters. If `linkLabelStateLength` > 998, the label state is removed, converted back to text via `mergeOrReplaceTextSegment`, and discarded.
   - Prevents nested links: If `!last.IsImage` and the label content already contains an `*ast.Link` (checked via `containsLink`), parsing aborts, turning the bracket into a text segment.
   - **Evaluates link type following `]`:**
     - **`(` encountered:** Executes `s.parseLink(...)` to parse inline destinations/titles.
     - **`[` encountered:** Executes `s.parseReferenceLink(...)` to parse full or collapsed reference links.
     - **Fallback (Shortcut reference link):** Looks up reference using text between brackets via `pc.LinkDefinition(...)`.
   - **AST Node Construction:**
     - If `last.IsImage` is true, converts the generated `*ast.Link` to an `*ast.Image`, moving all child nodes to the image node.
     - Sets the node starting position via `SetPos(...)`.
     - Returns the created `ast.Node` (`*ast.Link` or `*ast.Image`).

---

#### Supporting Parsing Helpers

- **`processLinkLabelOpen(block text.Reader, pos int, isImage bool, pc Context) *linkLabelState`**
  Creates a new `linkLabelState`, pushes it to context, advances the block reader by 1, and returns the state node.

- **`processLinkLabel(parent ast.Node, link *ast.Link, last *linkLabelState, pc Context)`**
  Pops delimiter bottom via `popLinkBottom`, invokes `ProcessDelimiters(bottom, pc)` to resolve inline formatting (e.g., emphasis) inside the link text, and transfers all sibling nodes between the opening bracket and `]` into the `link` node as children.

- **`parseLink(parent ast.Node, last *linkLabelState, block text.Reader, pc Context) *ast.Link`**
  Parses inline parenthesized link specifications: `(destination "title")`.
  - Skips opening `(`.
  - Calls `parseLinkDestination` to extract URIs (handles bracketed `<...>` or unbracketed forms).
  - Calls `parseLinkTitle` to extract titles wrapped in `"..."`, `'...'`, or `(...)`.
  - Expects closing `)`.

- **`parseReferenceLink(parent ast.Node, last *linkLabelState, block text.Reader, pc Context) (*ast.Link, bool)`**
  Parses full `[text][label]` or collapsed `[text][]` reference links using `findClosure(block, '[', ']')`. Lookups are evaluated via `pc.LinkDefinition(...)`.

- **`parseLinkDestination(block text.Reader) (text.SingleLineValue, bool)`**
  Parses link destinations:
  - Supports point-bracketed destinations `<...>`.
  - Supports standard destinations with balanced parentheses and backslash-escaped characters.

- **`parseLinkTitle(block text.Reader) (text.MultiLineValue, bool)`**
  Parses titles starting with `"`, `'`, or `(`. Uses `findClosure` to find matching quotes/closing parenthesis.

- **`findClosure(r text.Reader, opener, closer byte) (text.MultiLineValue, bool)`**
  Scans text lines to find a matching closing byte delimiter (`closer`), respecting backslash escape sequences (`\`). Restores reader position on failure.

- **`referenceLink(def LinkDefinition, kind ast.ReferenceLinkKind, refvalue text.MultiLineValue, decoder text.Decoder) *ast.Link`**
  Constructs an `*ast.Link` instance from a resolved `LinkDefinition` and reference type (`Full`, `Collapsed`, or `Shortcut`).

- **`containsLink(n ast.Node) bool`**
  Recursively inspects `n` and its siblings/children to check if an `*ast.Link` already exists inside a candidate link's text label.

---

### `CloseBlock` Method

```go
func (s *linkParser) CloseBlock(_ ast.Node, reader text.Reader, pc Context)
```

Called when the parent block closes. It cleans up unclosed link labels:
1. Resets the `linkBottom` context key to `nil`.
2. Iterates over any remaining `linkLabelState` items in `linkLabelStateKey`.
3. Removes state references via `removeLinkLabelState`.
4. Replaces unclosed label state nodes in the AST with plain `ast.Text` nodes containing the literal bracket text segments.