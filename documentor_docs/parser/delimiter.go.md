# Technical Documentation: `parser/delimiter.go`

## Overview

The `parser/delimiter.go` file provides data structures and functions for processing Markdown inline delimiters (such as `*` or `_` used for emphasis and strong emphasis) according to CommonMark parsing specifications. It defines the interface for custom delimiter logic, a AST node representation for delimiters, helper functions for flanking rules, and core algorithms for scanning and resolving delimiter pairs.

---

## Key Interfaces

### `DelimiterProcessor`

`DelimiterProcessor` defines the contract for handling specific delimiter characters during parsing.

```go
type DelimiterProcessor interface {
    IsDelimiter(byte) bool
    CanOpenCloser(opener, closer *Delimiter) bool
    OnMatch(consumes int) ast.Node
}
```

*   **`IsDelimiter(byte) bool`**: Returns `true` if the given byte character is recognized as a valid delimiter by this processor.
*   **`CanOpenCloser(opener, closer *Delimiter) bool`**: Determines whether a given `opener` delimiter node can be paired with a given `closer` delimiter node.
*   **`OnMatch(consumes int) ast.Node`**: Callback executed when a valid opener/closer pair is matched. It takes the number of consumed characters and returns the newly constructed AST node (e.g., an emphasis or strong emphasis node).

---

## Data Structures

### `Delimiter`

`Delimiter` represents an inline delimiter node within the AST and the active delimiter list. It embeds `ast.BaseInline`.

```go
type Delimiter struct {
    ast.BaseInline

    value text.Segment
    decoder text.Decoder

    CanOpen bool
    CanClose bool
    Length int
    OriginalLength int
    Char byte

    PreviousDelimiter *Delimiter
    NextDelimiter *Delimiter

    Processor DelimiterProcessor
}
```

#### Struct Fields

| Field | Type | Description |
| :--- | :--- | :--- |
| `value` | `text.Segment` | Text segment bounds of the delimiter in the source buffer. |
| `decoder` | `text.Decoder` | Character decoder associated with the text reader. |
| `CanOpen` | `bool` | `true` if this delimiter run can open a span. |
| `CanClose` | `bool` | `true` if this delimiter run can close a span. |
| `Length` | `int` | Remaining unconsumed character count of the delimiter run. |
| `OriginalLength` | `int` | Initial character count of the delimiter run before any matches. |
| `Char` | `byte` | The delimiter byte character (e.g., `'*'` or `'_'`). |
| `PreviousDelimiter`| `*Delimiter` | Pointer to the previous sibling delimiter in the active delimiter chain. |
| `NextDelimiter` | `*Delimiter` | Pointer to the next sibling delimiter in the active delimiter chain. |
| `Processor` | `DelimiterProcessor` | The processor instance associated with this delimiter. |

---

## Methods on `Delimiter`

### `Inline()`
```go
func (d *Delimiter) Inline()
```
An empty method satisfying the `ast.Inline` interface.

### `Dump(_ []byte) *ast.NodeDump`
```go
func (d *Delimiter) Dump(_ []byte) *ast.NodeDump
```
Returns an AST node dump containing `CanOpen`, `CanClose`, `OriginalLength`, and `Char` attributes for debugging.

### `Kind() ast.NodeKind`
```go
func (d *Delimiter) Kind() ast.NodeKind
```
Returns the AST node kind (`kindDelimiter`), defined internally as `ast.NewNodeKind("Delimiter")`.

### `ConsumeCharacters(n int)`
```go
func (d *Delimiter) ConsumeCharacters(n int)
```
Decrements the delimiter's remaining `Length` by `n` and updates the end boundary (`Stop`) of its internal text segment `value`.

### `CalcConsumption(closer *Delimiter) int`
```go
func (d *Delimiter) CalcConsumption(closer *Delimiter) int
```
Calculates how many characters should be consumed from the opener (`d`) and `closer` to form a span.
*   **Rule 1**: Returns `0` if `(d.CanClose || closer.CanOpen)` is true AND `(d.OriginalLength + closer.OriginalLength) % 3 == 0` AND `closer.OriginalLength % 3 != 0`. (Prevents invalid matches for multiple of 3 runs according to the CommonMark spec).
*   **Rule 2**: Returns `2` if both `d.Length >= 2` and `closer.Length >= 2`.
*   **Rule 3**: Otherwise, returns `1`.

---

## Constructors and Helper Functions

### `NewDelimiter(...)`
```go
func NewDelimiter(canOpen, canClose bool, length int, char byte, processor DelimiterProcessor) *Delimiter
```
Allocates and initializes a new `Delimiter` node with `Length` and `OriginalLength` set to `length`, `CanOpen`, `CanClose`, `Char`, and its associated `Processor`.

### `IsLeftFlankingDelimiterRun(before, after rune) bool`
```go
func IsLeftFlankingDelimiterRun(before, after rune) bool
```
Determines if a position is a left-flanking delimiter run according to CommonMark specification:
*   `after` is not whitespace AND
*   `after` is not punctuation, OR `before` is whitespace, OR `before` is punctuation.

### `IsRightFlankingDelimiterRun(before, after rune) bool`
```go
func IsRightFlankingDelimiterRun(before, after rune) bool
```
Determines if a position is a right-flanking delimiter run according to CommonMark specification:
*   `before` is not whitespace AND
*   `before` is not punctuation, OR `after` is whitespace, OR `after` is punctuation.

---

## Core Parsing & Processing Functions

### `ParseDelimiter`

```go
func ParseDelimiter(block text.Reader, minimum int, processor DelimiterProcessor, pc Context) *Delimiter
```

Scans the input buffer for a delimiter run using the specified `DelimiterProcessor`.

#### Process Flow:
1.  Obtains `before` rune using `block.PrecedingCharacter()`.
2.  Peeks the current line segment using `block.PeekLine()`.
3.  Checks if line is empty or if `line[0]` is not a valid delimiter byte according to `processor.IsDelimiter`. If not, returns `nil`.
4.  Counts consecutive identical delimiter characters `j`. If `j < minimum`, returns `nil`.
5.  Determines the `after` rune following the delimiter run (defaults to space `' '` if run extends to end of line).
6.  Computes `isLeft` via `IsLeftFlankingDelimiterRun` and `isRight` via `IsRightFlankingDelimiterRun`.
7.  Determines `canOpen` and `canClose`:
    *   If character is `'_'` (underscore):
        *   `canOpen = isLeft && (!isRight || beforeIsPunctuation)`
        *   `canClose = isRight && (!isLeft || afterIsPunctuation)`
    *   Otherwise:
        *   `canOpen = isLeft`
        *   `canClose = isRight`
8.  Constructs a `Delimiter` node via `NewDelimiter`.
9.  Sets segment start/stop on the delimiter node and stores reader decoder.
10. Advances reader by `j` bytes (`block.Advance(j)`).
11. Pushes the delimiter onto the delimiter stack context via `pc.PushDelimiter(node)`.
12. Returns the created `Delimiter`.

---

### `ProcessDelimiters`

```go
func ProcessDelimiters(bottom ast.Node, pc Context)
```

Processes accumulated delimiters within the context `pc` from `bottom` upwards to match opening and closing delimiters and construct AST inline spans.

#### Algorithm Steps:

1.  Retrieves `lastDelimiter` from `pc.LastDelimiter()`. If `nil`, processing terminates immediately.
2.  **Determine Starting Closer Node**:
    *   If `bottom` is provided and is not `lastDelimiter`, scans backward from `lastDelimiter.PreviousSibling()` until reaching `bottom` to set `closer` to the earliest encountered `Delimiter`.
    *   If `bottom` is `nil`, sets `closer = pc.FirstDelimiter()`.
    *   If no closer is found, calls `pc.ClearDelimiters(bottom)` and returns.
3.  **Delimiter Matching Loop**:
    *   Iterates through `closer` chain via `closer.NextDelimiter`.
    *   If `!closer.CanClose`, advances to the next closer.
    *   Scans backward from `closer.PreviousDelimiter` looking for a matching `opener` until reaching `bottom`.
    *   Checks if `opener.CanOpen` is `true` and `opener.Processor.CanOpenCloser(opener, closer)` returns `true`.
    *   Calculates character consumption: `consume = opener.CalcConsumption(closer)`.
    *   **If a match is found (`consume > 0`)**:
        1.  Calls `ConsumeCharacters(consume)` on both `opener` and `closer`.
        2.  Creates new AST node via `opener.Processor.OnMatch(consume)` and sets its position to `opener.value.Start`.
        3.  Moves all AST child nodes located strictly between `opener` and `closer` into the new node (`node.AppendChild`).
        4.  Inserts the new `node` into the AST parent immediately after `opener` (`parent.InsertAfter(opener, node)`).
        5.  Removes any intermediate delimiters remaining between `opener` and `closer` from the context using `pc.RemoveDelimiter`.
        6.  If `opener.Length == 0`, removes `opener` from `pc`.
        7.  If `closer.Length == 0`, removes `closer` from `pc` and updates `closer` to `closer.NextDelimiter`.
    *   **If no match is found**:
        *   If `closer` cannot open (`!closer.CanOpen`) and no potential opener was found (`!maybeOpener`), removes `closer` from `pc`.
        *   Advances `closer` to `closer.NextDelimiter`.
4.  Clears all remaining unused delimiters up to `bottom` via `pc.ClearDelimiters(bottom)`.