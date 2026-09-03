# Technical Documentation: `parser/list_item.go`

## Overview

The `parser/list_item.go` file contains the implementation of the `listItemParser` in Goldmark, which is responsible for detecting, opening, continuing, and closing Markdown list items (represented as `ast.ListItem` nodes). It implements the `BlockParser` interface.

---

## Constants and Package Variables

* **`listItemParser`**: An unexported struct that implements the `BlockParser` interface.
* **`defaultListItemParser`**: A package-level instance of `listItemParser` used as a singleton.

---

## Constructor

### `NewListItemParser() BlockParser`
Returns the `defaultListItemParser` instance as a `BlockParser` interface type.

---

## Method Details

### `Trigger() []byte`
Returns a slice of bytes containing characters that can trigger the execution of the list item parser:
* Bullet markers: `-`, `+`, `*`
* Ordered list digits: `0`, `1`, `2`, `3`, `4`, `5`, `6`, `7`, `8`, `9`

---

### `Open(parent ast.Node, reader text.Reader, pc Context) (ast.Node, State)`

Determines whether a new list item block should be opened on the current line.

#### Execution Flow:
1. **Parent Check**: Ensures `parent` is of type `*ast.List`. If the parent is not a list, it returns `nil, NoChildren` because a list item must be a child of a list node.
2. **Offset Check**: Retrieves the offset of the parent list using `lastOffset(list)`.
3. **Line Inspection**: Peeks at the current line using `reader.PeekLine()` and parses it using `parseListItem(line)`.
4. **List Item Validation**:
   * If `typ == notList`, returns `nil, NoChildren`.
   * If the relative indentation of the marker exceed 3 spaces (`match[1] - offset > 3`), returns `nil, NoChildren`.
5. **Context Reset**: Resets the context key `emptyListItemWithBlankLinesKey` to `nil`.
6. **Node Creation**:
   * Calculates the item offset using `calcListOffset(line, match)`.
   * Creates a new `ast.ListItem` node (`ast.NewListItem()`).
   * Sets the node's offset to `match[3] + itemOffset`.
7. **Content Inspection**:
   * Checks if the list item content segment is non-existent (`match[4] < 0`) or blank (`util.IsBlank(line[match[4]:match[5]])`). If so, returns `node, NoChildren`.
   * If content is present, calculates child position and padding using `util.IndentPosition(line[match[4]:], match[4], itemOffset)`.
   * Advances the reader using `reader.AdvanceAndSetPadding(child, padding)`.
   * Returns `node, HasChildren`.

---

### `Continue(node ast.Node, reader text.Reader, pc Context) State`

Determines whether an existing list item continues onto the current line.

#### Execution Flow:
1. **Blank Line Processing**:
   * Checks if the line is blank via `util.IsBlank(line)`.
   * If blank, advances the reader to the end of the line (`reader.AdvanceToEOL()`) and returns `Continue | HasChildren`.
2. **Indentation and Closure Checks**:
   * Retrieves the parent node's offset via `lastOffset(node.Parent())`.
   * Checks if the node is empty (`isEmpty`), defined as `node.ChildCount() == 0` and `pc.Get(emptyListItemWithBlankLinesKey) != nil`.
   * Calculates line indentation width via `util.IndentWidth(line, reader.LineOffset())`.
   * If `(isEmpty || indent < offset) && indent < 4`:
     * Checks if the line starts a new list item using `parseListItem(line)`.
     * If a new list item is found (`typ != notList`), sets `pc.Set(skipListParserKey, listItemFlagValue)` and returns `Close`.
     * If not a new list item and `!isEmpty`, returns `Close`.
3. **Continuing Content**:
   * Computes position and padding using `util.IndentPosition(line, reader.LineOffset(), offset)`.
   * Advances the reader via `reader.AdvanceAndSetPadding(pos, padding)`.
   * Returns `Continue | HasChildren`.

---

### `Close(_ ast.Node, _ text.Reader, _ Context)`

Performs closing operations for a list item node. This method is a no-op (contains no implementation logic).

---

### `CanInterruptParagraph() bool`

Returns `true`, indicating that a list item marker can interrupt an ongoing paragraph block.

---

### `CanAcceptIndentedLine() bool`

Returns `false`, indicating that the list item parser does not accept standard 4-space indented lines as list item triggers.

---

## Context Keys and Helper Dependencies Used

The code references the following package-level functions, variables, and context keys defined elsewhere in the `parser` package or dependencies:

* **Context Keys**:
  * `emptyListItemWithBlankLinesKey`
  * `skipListParserKey`
  * `listItemFlagValue`
* **Package Helper Functions**:
  * `lastOffset(node ast.Node)`
  * `parseListItem(line []byte)`
  * `calcListOffset(line []byte, match []int)`
  * `notList` (type constant returned by `parseListItem`)
* **AST & Utility Imports**:
  * `ast.List`, `ast.ListItem`, `ast.NewListItem()`
  * `util.IsBlank()`, `util.IndentPosition()`, `util.IndentWidth()`