# Documentation: `ast_test.go`

## Overview

The `ast_test.go` file contains unit tests for the Abstract Syntax Tree (AST) features of the `goldmark` Markdown parsing library (version 2). Located within the `goldmark_test` package, this test file verifies structural properties, source positioning offset tracking, text processing (raw vs. entity-decoded text), and pretty-printing capabilities across various Markdown block and inline AST nodes.

---

## Dependencies

The test suite imports the following packages:

* `testing`: Standard Go library package for unit testing.
* `github.com/yuin/goldmark/v2/ast`: Provides AST node definitions and interfaces (`BlockNode`, `CodeSpan`, `Text`, `RawHTML`, `AutoLink`, `Link`).
* `github.com/yuin/goldmark/v2/parser`: Provides the Markdown parsing interface and configuration options (`New`, `Parse`, `NewContext`, `WithContext`, `WithPrettyPrint`).

---

## Key Test Functions

### 1. `TestHasBlankPreviousLines(t *testing.T)`

#### Purpose
Verifies the `HasBlankPreviousLines()` method on `ast.BlockNode`. This method determines whether a block node is preceded by one or more blank lines within various nested block constructs (e.g., blockquotes, lists).

#### Mechanism
Uses a table-driven test approach (`cases` struct slice). Each case defines:
* `Name`: Test description.
* `Source`: Markdown input string containing nested structures.
* `Node`: A closure function that navigates the parsed AST node hierarchy starting from the root node `n` to isolate the target node.
* `Expected`: Boolean expectation of `HasBlankPreviousLines()`.

#### Test Cases Covered

| Test Case Name | Target Node Structural Hierarchy | Expected Result |
| :--- | :--- | :--- |
| **nesting paragraphs in blockquotes** | Paragraph preceded by a blank line inside a blockquote | `true` |
| **nesting HTML blocks in blockquotes** (with blank line) | HTML block preceded by a blank line inside a blockquote | `true` |
| **nesting HTML blocks in blockquotes** (without blank line) | HTML block directly following another HTML block inside a blockquote | `false` |
| **nesting loose lists in blockquotes** | List item in a loose list inside a blockquote | `true` |
| **nesting tight lists in blockquotes** | List item in a tight list inside a blockquote | `false` |
| **nesting paragraphs in lists** | Paragraph inside a list item separated by blank lines | `true` |

---

### 2. `TestInlinePos(t *testing.T)`

#### Purpose
Validates that inline AST nodes record the correct starting byte offset position (`Pos()`) within the original source byte slice.

#### Mechanism
Parses a Markdown source string containing link reference definitions, inline links, emphasis, and image syntax. It traverses the AST child nodes sequentially and asserts expected byte positions (`Pos()`).

#### Validated Positions

* **1st link reference**: Position `0`
* **2nd link reference**: Position `9`
* **3rd link reference**: Position `21`
* **1st inline link**: Position `28`
* **1st emphasis (`**b**`)**: Position `60`
* **1st image (`![aaa](...)`)**: Position `68`

---

### 3. `TestBlockPos(t *testing.T)`

#### Purpose
Validates that top-level and nested block AST nodes accurately record their starting byte offset position (`Pos()`) within the source input.

#### Mechanism
Parses a multi-element Markdown document containing various block types. It steps through top-level sibling nodes using `NextSibling()` and verifies the value returned by `Pos()`.

#### Validated Block Nodes and Positions

| Block Node Type | Verified Position (`Pos()`) |
| :--- | :--- |
| **Paragraph** | `0` |
| **Heading** | `16` |
| **Thematic Break** | `28` |
| **Blockquote** | `33` |
| **Unordered List** | `47` |
| **Ordered List** | `60` |
| **Indented Code Block / Fenced Code Block** | `91` |
| **HTML Block** | `114` |
| **Link Reference Definition** | `135` |

---

### 4. `TestRawText(t *testing.T)`

#### Purpose
Tests text evaluation and entity unescaping behavior across different AST node types using `.Value.Value(source)`. This confirms whether specific nodes preserve raw input strings (such as HTML entities) or perform decoding.

#### Tested Node Behaviors

1. **`ast.CodeSpan`**: Retains raw text verbatim (`inline &amp; value`).
2. **`ast.Text`** (Standard inline text): Automatically decodes entities (converts `inline &amp; value` to ` inline & value `).
3. **`ast.RawHTML`** (Opening tag): Retains raw string (`<code class="&AElig;">`).
4. **`ast.Text`** (Inside raw HTML element): Decodes HTML entities (converts `inline &AElig; value` to `inline Æ value`).
5. **`ast.RawHTML`** (Closing tag): Retains raw string (`</code>`).
6. **`ast.AutoLink`**: Destination, Label, and Text attributes retain raw formatting (`http://www.example.com/&AElig;` and `<http://www.example.com/&AElig;>`).
7. **`ast.Link`**: Link destination field decodes HTML entities (`http://www.example.com/Æ`).

---

### 5. `TestPP(_ *testing.T)`

#### Purpose
Tests AST pretty-printing option (`parser.WithPrettyPrint(ast.WithSource(false))`) to ensure AST nodes can be parsed and formatted via the pretty printer without causing runtime panics or errors.