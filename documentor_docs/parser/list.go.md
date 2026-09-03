# Technical Documentation: `parser/list.go`

## Overview

The `parser/list.go` file provides block parsing logic for Markdown lists (both bullet lists and ordered lists) in the Goldmark AST parsing pipeline. It implements the `BlockParser` interface to detect list markers, create list AST nodes (`ast.List`), handle list continuity across lines, manage paragraph interruptions, and determine whether a list is tight or loose.

---

## Types and Constants

### `listItemType`
An internal integer type used to classify list item markers.

```go
type listItemType int

const (
	notList listItemType = iota // 0: Not a list item
	bulletList                  // 1: Bullet list item ('-', '*', '+')
	orderedList                 // 2: Ordered list item ('1.', '1)', etc.)
)
```

### Package Variables & Context Keys

* **`skipListParserKey`**: `ContextKey` used in the parser context (`Context`) to signal that the list parser should be bypassed for the current iteration.
* **`emptyListItemWithBlankLinesKey`**: `ContextKey` used to track if an empty list item was followed by blank lines.
* **`listItemFlagValue`**: `any` set to `true`, used as the value when writing `emptyListItemWithBlankLinesKey` to context.
* **`defaultListParser`**: Singleton instance of `*listParser`.

---

## Helper Functions

### `lastOffset(node ast.Node) int`
Calculates the indent offset of the last child of a given AST node.

* **Parameters**: `node` (`ast.Node`) - The parent node (typically an `ast.List`).
* **Returns**: `int` - Returns the `Offset()` of the last `ast.ListItem` child if it exists; otherwise returns `0`.

---

### `parseListItem(line []byte) ([6]int, listItemType)`
Parses a line byte slice to check if it matches a bullet or ordered list item marker. It is equivalent to matching the regular expressions:
* Bullet: `^(([ ]*)([\-\*\+]))(\s+.*)?\n?$`
* Ordered: `^(([ ]*)(\d{1,9}[\.\)]))(\s+.*)?\n?$`

#### Execution Logic:
1. **Indent Check**: Skips up to 3 leading spaces. If a tab (`\t`) is encountered during leading space checking, or if leading spaces exceed 3, it returns `notList`.
2. **Marker Identification**:
   * **Bullet List**: Matches `-`, `*`, or `+`.
   * **Ordered List**: Matches 1 to 9 digits (`0-9`) followed by standard delimiters (`.` or `)`). Returns `notList` if there are no digits or if digit count exceeds 9.
3. **Delimiter Spacing Validation**: Validates that content following the marker is preceded by appropriate whitespace or a newline.
4. **Returns**:
   * A 6-element index array `[6]int` specifying submatch offsets:
     * `[0], [1]`: Line start to end of indent.
     * `[2], [3]`: Start and end indices of the list marker.
     * `[4], [5]`: Content start index (`-1` if empty) and line content end index.
   * `listItemType`: `bulletList`, `orderedList`, or `notList`.

---

### `calcListOffset(line []byte, match [6]int) int`
Calculates the content indent offset for a list item based on the parsed match ranges.

#### Execution Logic:
* If the line content range is empty (`match[4] < 0`) or blank, returns `1`.
* Otherwise, calculates the width using `util.IndentWidth(line[match[4]:], match[4])`. If the calculated offset exceeds 4 (indicating an indented code block within the item), returns `1`.

---

## Struct `listParser`

`listParser` is the central block parser struct that manages the lifecycle of an `ast.List` node.

```go
type listParser struct{}
```

### Factory Function

#### `NewListParser() BlockParser`
Returns the default singleton instance (`defaultListParser`) of `listParser`.

---

## `listParser` Methods (`BlockParser` Implementation)

### `Trigger() []byte`
Returns the set of characters that can trigger list parsing:
* Bullet markers: `'-'`, `'+'`, `'*'`
* Ordered list digits: `'0'`, `'1'`, `'2'`, `'3'`, `'4'`, `'5'`, `'6'`, `'7'`, `'8'`, `'9'`

---

### `Open(parent ast.Node, reader text.Reader, pc Context) (ast.Node, State)`
Attempts to open a new list block node (`ast.List`) at the current line position.

#### Execution Logic:
1. **Context Check**: Checks if the last opened block is already an `ast.List` or if `skipListParserKey` is present in the context. If so, clears `skipListParserKey` and returns `nil, NoChildren`.
2. **Line Parsing**: Calls `parseListItem()` on the peeked line. If `notList`, returns `nil, NoChildren`.
3. **Ordered List Start Number**: Extracts digits from an ordered list marker and converts them to an integer using `strconv.Atoi`.
4. **Paragraph Interruption Rules**: If the last opened block is a paragraph and is a child of `parent`:
   * Ordered lists can only interrupt a paragraph if the start number is `1`.
   * Empty list items cannot interrupt a paragraph.
5. **Node Construction**: Creates an `ast.List` with the marker character (`line[match[3]-1]`). Sets `node.Start` if ordered.
6. **State Context Cleanup**: Resets `emptyListItemWithBlankLinesKey` to `nil` in the context and returns the node with state `HasChildren`.

---

### `Continue(node ast.Node, reader text.Reader, pc Context) State`
Evaluates whether the current list block should continue processing subsequent lines or close.

#### Execution Logic:
1. **Blank Line Check**:
   * If the line is blank and the last item in the list contains no child nodes, sets `emptyListItemWithBlankLinesKey` in context.
   * Returns `Continue | HasChildren`.
2. **Indent and Marker Continuity**:
   * Evaluates line indentation against `lastOffset(node)` and checks if the last list item is empty.
   * If indentation is less than 4 and less than offset:
     * Re-runs `parseListItem()`.
     * If valid, checks `list.CanContinue(marker, isOrdered)`. Returns `Close` if types/delimiters do not match.
     * Checks if the line is a thematic break or setext heading bar. Returns `Close` if a thematic break takes precedence.
3. **Empty Item Blank Line Restriction**:
   * If `emptyListItemWithBlankLinesKey` is present in context and non-empty list item text appears, returns `Close` (prevents non-empty items from attaching directly to an empty item separated by blank lines).
4. **Default Action**: Returns `Continue | HasChildren`.

---

### `Close(node ast.Node, _ text.Reader, _ Context)`
Finalizes the list node when processing of the list block ends. Determines list tightness (`list.IsTight`).

#### Tightness Algorithm:
1. Iterates over all child list items (`ast.ListItem`) of the list node.
2. Checks if any child block nodes inside a list item have preceding blank lines (`HasBlankPreviousLines()`).
3. Checks if subsequent list items (after the first child item) have preceding blank lines.
4. If blank lines are detected between items or blocks within items, `list.IsTight` is set to `false`.

---

### `CanInterruptParagraph() bool`
Returns `true`. Indicates that lists can interrupt an active paragraph under valid conditions (handled in `Open`).

---

### `CanAcceptIndentedLine() bool`
Returns `false`. Indicates that indented lines should not be automatically accepted without offset and marker validation.