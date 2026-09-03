# Technical Documentation: `parser/emphasis.go`

## Overview

The `parser/emphasis.go` file provides functionality for parsing Markdown emphasis (typically rendered as `<em>`) and strong emphasis (typically rendered as `<strong>`) constructs. It defines an inline parser (`emphasisParser`) triggered by delimiter bytes (`*` and `_`) and a delimiter processor (`emphasisDelimiterProcessor`) that determines delimiter matching rules and generates the corresponding AST nodes.

---

## Constants and Package Variables

### `defaultEmphasisDelimiterProcessor`
```go
var defaultEmphasisDelimiterProcessor = &emphasisDelimiterProcessor{}
```
An internal package-level instance of `emphasisDelimiterProcessor` used as the default delimiter processor during emphasis parsing.

### `defaultEmphasisParser`
```go
var defaultEmphasisParser = &emphasisParser{}
```
An internal package-level instance of `emphasisParser` used to serve requests created by `NewEmphasisParser()`.

---

## Structs and Interface Implementations

### 1. `emphasisDelimiterProcessor`

`emphasisDelimiterProcessor` handles the delimiter processing logic for matching opening and closing emphasis characters (`*` and `_`).

#### Methods

##### `IsDelimiter(b byte) bool`
* **Parameters**: `b byte` — The character byte to test.
* **Returns**: `bool` — `true` if the byte is an asterisk (`'*'`) or an underscore (`'_'`), otherwise `false`.
* **Purpose**: Identifies whether a given byte is a valid emphasis delimiter character.

##### `CanOpenCloser(opener, closer *Delimiter) bool`
* **Parameters**:
  * `opener *Delimiter` — The opening delimiter candidate.
  * `closer *Delimiter` — The closing delimiter candidate.
* **Returns**: `bool` — `true` if the opening character matches the closing character (`opener.Char == closer.Char`), otherwise `false`.
* **Purpose**: Ensures that an opening delimiter and a closing delimiter can form a pair only if they use the same character byte (e.g., `*` with `*`, or `_` with `_`).

##### `OnMatch(consumes int) ast.Node`
* **Parameters**: `consumes int` — The number of delimiter characters consumed in the match.
* **Returns**: `ast.Node` — An AST node representing the matched emphasis structure.
* **Logic**:
  * If `consumes == 1`, it creates and returns a standard emphasis node (`ast.NewEmphasis()`).
  * For any other consumption count (e.g., `consumes == 2`), it creates and returns a strong emphasis node (`ast.NewStrong()`).

---

### 2. `emphasisParser`

`emphasisParser` implements the `InlineParser` interface to process emphasis triggers within a block of text.

#### Methods

##### `Trigger() []byte`
* **Returns**: `[]byte` — A slice containing the trigger bytes: `[]byte{'*', '_'}`.
* **Purpose**: Registers the parser to be invoked whenever an asterisk (`*`) or underscore (`_`) byte is encountered in the input text stream.

##### `Parse(parent ast.Node, block text.Reader, pc Context) ast.Node`
* **Parameters**:
  * `_ ast.Node` — The parent AST node (unused in this method).
  * `block text.Reader` — The text reader positioned at the current input block.
  * `pc Context` — The parser context.
* **Returns**: `ast.Node` — The AST node produced by delimiter parsing.
* **Logic**: Delegates the parsing execution to `ParseDelimiter`, passing the reader, a minimum delimiter length of `1`, the shared `defaultEmphasisDelimiterProcessor`, and the parser context `pc`.

---

## Constructor Functions

### `NewEmphasisParser()`

```go
func NewEmphasisParser() InlineParser
```

* **Returns**: `InlineParser` — Returns the package singleton `defaultEmphasisParser`.
* **Purpose**: Provides a public constructor function to obtain an `InlineParser` capable of parsing standard Markdown emphasis and strong emphasis.

---

## How It Works (Execution Flow)

1. **Triggering**: During inline parsing, when the parser encounters a byte defined in `Trigger()` (`*` or `_`), `emphasisParser.Parse` is invoked.
2. **Delimiter Parsing**: `emphasisParser.Parse` calls `ParseDelimiter`, passing `defaultEmphasisDelimiterProcessor` to handle delimiter evaluation.
3. **Delimiter Verification**:
   - `IsDelimiter` validates whether target bytes are valid emphasis characters (`*` or `_`).
   - `CanOpenCloser` ensures closing delimiters match their opening counterpart's character.
4. **Node Generation**:
   - When a valid delimiter pair match is made, `OnMatch` determines the node type based on `consumes`:
     - `1` consumed character $\rightarrow$ `ast.NewEmphasis()`
     - Otherwise $\rightarrow$ `ast.NewStrong()`