# Technical Documentation: `parser/blockquote.go`

## Overview

The `parser/blockquote.go` file implements the parsing logic for Markdown blockquotes as part of the `parser` package (utilizing Goldmark v2 APIs). It defines a struct `blockquoteParser` that implements the `BlockParser` interface, detecting, opening, continuing, and closing standard Markdown blockquote nodes (`>`).

---

## Constants and Package Variables

* **`defaultBlockquoteParser`**: A package-private singleton instance of `blockquoteParser`.

---

## Types and Functions

### `NewBlockquoteParser`

```go
func NewBlockquoteParser() BlockParser
```

Returns the package-level default `BlockParser` implementation (`defaultBlockquoteParser`) for blockquotes.

---

## Methods of `blockquoteParser`

### `process`

```go
func (b *blockquoteParser) process(reader text.Reader) bool
```

`process` is an internal helper method that inspects the current line in the `text.Reader` to determine if it contains a valid blockquote marker (`>`).

#### Step-by-Step Logic:
1. **Peeks current line**: Reads the current line using `reader.PeekLine()`.
2. **Indentation Check**: Computes indentation width `w` and character offset `pos` using `util.IndentWidth(line, reader.LineOffset())`.
   * If indentation `w > 3`, or `pos` exceeds the line length, or the character at `pos` is not `'>'`, it returns `false`.
3. **Marker Consumption**: Increments `pos` past the `>` marker.
4. **End of Line / Empty Line Handling**:
   * If `pos` reaches or exceeds the end of the line, or if the next character is `\n`, the reader advances by `pos` and returns `true`.
5. **Space/Tab Handling**:
   * Advances the reader up to the `>` character.
   * Checks if the character immediately following `>` is a space `' '` or tab `'\t'`.
   * If it is a tab `'\t'`, calculates `padding = util.TabWidth(reader.LineOffset()) - 1`.
   * Calls `reader.AdvanceAndSetPadding(1, padding)` to consume the space/tab and set any required tab expansion padding.
6. Returns `true` indicating a blockquote marker was matched and consumed.

---

### `Trigger`

```go
func (b *blockquoteParser) Trigger() []byte
```

Returns the trigger byte slice containing `[]byte{'>'}`. This tells the block parser registry which byte triggers this parser.

---

### `Open`

```go
func (b *blockquoteParser) Open(_ ast.Node, reader text.Reader, _ Context) (ast.Node, State)
```

Invoked when entering a potential new blockquote block.

* **Behavior**:
  * Calls `b.process(reader)`.
  * If `process` succeeds (`true`), it creates and returns a new AST blockquote node (`ast.NewBlockquote()`) along with state `HasChildren`.
  * If `process` fails (`false`), it returns `nil` and state `NoChildren`.

---

### `Continue`

```go
func (b *blockquoteParser) Continue(_ ast.Node, reader text.Reader, _ Context) State
```

Invoked on subsequent lines to check if the current blockquote continuation condition holds.

* **Behavior**:
  * Calls `b.process(reader)`.
  * If `process` succeeds (`true`), it returns the state `Continue | HasChildren`.
  * If `process` fails (`false`), it returns the state `Close`.

---

### `Close`

```go
func (b *blockquoteParser) Close(_ ast.Node, _ text.Reader, _ Context)
```

Invoked when closing a blockquote block. This implementation performs no cleanup operations (empty body).

---

### `CanInterruptParagraph`

```go
func (b *blockquoteParser) CanInterruptParagraph() bool
```

Returns `true`, indicating that a blockquote marker can interrupt an active paragraph block without requiring a blank line preceding it.

---

### `CanAcceptIndentedLine`

```go
func (b *blockquoteParser) CanAcceptIndentedLine() bool
```

Returns `false`, indicating that an indented line (4 or more spaces of indentation) cannot be implicitly accepted as part of a blockquote.