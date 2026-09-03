# Technical Documentation: `extension/typographer.go`

## Overview

The `extension/typographer.go` file implements a Goldmark markdown parser extension that transforms standard ASCII punctuation into typographic HTML entities (commonly known as "smart typography" or "smart quotes"). 

It provides inline parsing capability for single quotes, double quotes, hyphens/dashes, ellipses, and angle quotes, while maintaining contextual state to accurately distinguish between left/right opening/closing marks, apostrophes, contractions, and special edge cases (e.g., decade abbreviations like `'90s`).

---

## Key Types and Constants

### `TypographicPunctuation`

An integer-based enumeration (`int`) representing the specific punctuation marks eligible for typographic substitution:

| Constant | Description | Default Replacement Entity |
| :--- | :--- | :--- |
| `LeftSingleQuote` | Left single quotation mark (`'`) | `&lsquo;` |
| `RightSingleQuote` | Right single quotation mark (`'`) | `&rsquo;` |
| `LeftDoubleQuote` | Left double quotation mark (`"`) | `&ldquo;` |
| `RightDoubleQuote` | Right double quotation mark (`"`) | `&rdquo;` |
| `EnDash` | En dash (`--`) | `&ndash;` |
| `EmDash` | Em dash (`---`) | `&mdash;` |
| `Ellipsis` | Ellipsis (`...`) | `&hellip;` |
| `LeftAngleQuote` | Left guillemet (`<<`) | `&laquo;` |
| `RightAngleQuote` | Right guillemet (`>>`) | `&raquo;` |
| `Apostrophe` | Apostrophe (`'`) | `&rsquo;` |

`typographicPunctuationMax` is an unexported sentinel constant defining the bounds for array/slice allocations of substitutions.

### `TypographicSubstitutions`

```go
type TypographicSubstitutions map[TypographicPunctuation]string
```

A map alias connecting `TypographicPunctuation` constant keys to custom replacement string values.

### State Tracking Types

* **`unclosedCounter`**: A struct tracking unmatched open quotation marks within the current parsing context.
  * `Single int`: Count of open single quotes.
  * `Double int`: Count of open double quotes.
  * `Reset()`: Method to reset both counters back to zero.
* **`uncloseCounterKey`**: Internal context key (`parser.NewContextKey()`) used to store and retrieve `unclosedCounter` instances within `parser.Context`.
* **`getUnclosedCounter(pc parser.Context) *unclosedCounter`**: Helper function that fetches or initializes the `unclosedCounter` attached to the current context.

---

## Configuration and Functional Options

### `typographerConfig`

An internal struct holding the active replacement substitutions as a slice of strings indexed by `TypographicPunctuation`.

```go
type typographerConfig struct {
    Substitutions []string
}
```

### Options and Constructors

#### `newDefaultSubstitutions() []string`
Returns a slice pre-populated with default HTML entity strings for each `TypographicPunctuation` constant.

#### `TypographerParserOption`
A function signature type `func(*typographerConfig)` used to configure the parser.

#### `WithTypographicSubstitutions[T []byte | string](values map[TypographicPunctuation]T) TypographerParserOption`
A generic functional option allowing callers to override specific default punctuation substitutions using a map. Accepts map values typed as either `string` or `[]byte`.

---

## Inline Parser Implementation

The internal `typographerParser` type implements the Goldmark `parser.InlineParser` interface.

```go
type typographerParser struct {
    typographerConfig
}
```

### Trigger Characters

The parser registers to trigger on the following ASCII byte characters:
```go
func (s *typographerParser) Trigger() []byte {
    return []byte{'\'', '"', '-', '.', ',', '<', '>', '*', '['}
}
```

### `Parse` Logic Breakdown

`Parse(_ gast.Node, block text.Reader, pc parser.Context) gast.Node` evaluates characters at the reader's current position and determines if a substitution applies:

1. **Multi-Character Patterns (Length >= 3)**:
   * `---` $\rightarrow$ Evaluates to `EmDash` (`&mdash;`), advances `block` by 3 bytes.
   * `...` $\rightarrow$ Evaluates to `Ellipsis` (`&hellip;`), advances `block` by 3 bytes.

2. **Two-Character Patterns (Length >= 2)**:
   * `<<` $\rightarrow$ Evaluates to `LeftAngleQuote` (`&laquo;`), advances `block` by 2 bytes.
   * `>>` $\rightarrow$ Evaluates to `RightAngleQuote` (`&raquo;`), advances `block` by 2 bytes.
   * `--` $\rightarrow$ Evaluates to `EnDash` (`&ndash;`), advances `block` by 2 bytes.

3. **Quote Characters (`'` and `"`)**:
   Flanking checks (`parser.IsLeftFlankingDelimiterRun`, `parser.IsRightFlankingDelimiterRun`) determine whether the quote can act as an opening or closing delimiter.

   * **Single Quote (`'`) Processing**:
     * **Decade Abbreviations**: Checks for formats like `'90s` (opening single quote followed by two digits and `'s'`). Replaced with `Apostrophe`.
     * **Prefix Contractions**: Checks for words starting with contractions such as `'twas`, `'em`, `'net`, or `'l`. Replaced with `Apostrophe`.
     * **Word-Internal Apostrophes**: Checks if surrounded by alphanumeric/letter characters (e.g., `don't`, `it's`). Replaced with `Apostrophe`.
     * **Left Single Quotes**: Applied if `canOpen && !canClose`. Evaluates potential right-quote exceptions for common word endings (`'s`, `'m`, `'t`, `'d`, `'ve`, `'ll`, `'re`). If evaluated as a true left single quote, increments `counter.Single`.
     * **Plural Possessives & Word-End Abbreviations**: Formats like `Smiths'` or `doin'`. Evaluated to `RightSingleQuote`.
     * **Right Single Quotes**: Applied if `counter.Single > 0` and closing criteria are met (`canClose`). Decrements `counter.Single`.

   * **Double Quote (`"`) Processing**:
     * **Left Double Quote**: Applied if `canOpen && !canClose`. Increments `counter.Double`.
     * **Right Double Quote**: Applied if `counter.Double > 0` and closing conditions (`canClose`) are satisfied. 
       * *Edge Case*: Ignores double quotes directly preceded by digits followed immediately by another quote (e.g., `"Monitor 21""`).
       * On replacement, decrements `counter.Double`.

If a substitution match is made, `Parse` advances the `block` reader by the consumed character count and returns a new AST text node (`gast.NewText`) containing the substitution string decoded via `block.Decoder()`. Otherwise, it returns `nil`.

### Block Lifetime Management (`CloseBlock`)

```go
func (s *typographerParser) CloseBlock(_ gast.Node, pc parser.Context) {
    getUnclosedCounter(pc).Reset()
}
```
Resets the `unclosedCounter` in `parser.Context` at the conclusion of parsing a block node to ensure quote state tracking does not bleed across block boundaries.

---

## Extension Interface and Registration

### `typographerParserExtension`

Implements Goldmark's `parser.Extension` interface.

```go
type typographerParserExtension struct {
    options []TypographerParserOption
}
```

* **`ParserOptions(_ *parser.Config) []parser.Option`**: Registers `typographerParser` as an inline parser with high priority (`9999`) using `util.Prioritized`.

### Constructors and Default Instance

* **`NewTypographerParser(opts ...TypographerParserOption) parser.Extension`**: Returns a new `parser.Extension` initialized with any provided functional options.
* **`TypographerParser`**: A pre-initialized package-level variable instance of `parser.Extension` using default settings.

---

## Usage Example

```go
import (
    "github.com/yuin/goldmark/v2"
    "github.com/yuin/goldmark/v2/extension"
)

// Using the default TypographerParser instance
markdown := goldmark.New(
    goldmark.WithExtensions(
        extension.TypographerParser,
    ),
)

// Or initializing with custom substitutions
customTypographer := extension.NewTypographerParser(
    extension.WithTypographicSubstitutions(map[extension.TypographicPunctuation]string{
        extension.LeftDoubleQuote:  "«",
        extension.RightDoubleQuote: "»",
    }),
)

markdownWithCustom := goldmark.New(
    goldmark.WithExtensions(
        customTypographer,
    ),
)
```