# Technical Documentation: `extension/tasklist.go`

## Overview

The `extension/tasklist.go` file implements support for Markdown task lists (checkboxes inside list items, e.g., `- [ ] item` or `- [x] item`) for the `goldmark` (v2) Markdown parser engine. 

It provides components for:
1. **Parsing**: Identifying checkbox syntax at the start of list items and attaching task status metadata (`task-status`) to the underlying Abstract Syntax Tree (AST) list item nodes.
2. **Rendering**: Generating HTML checkbox `<input>` elements for task list items during HTML output generation, supporting both tight and loose paragraph blocks as well as XHTML output format.

---

## Constants and Types

### Task Status Types

```go
type TaskStatus string

const (
    TaskStatusActive    TaskStatus = "active"
    TaskStatusCompleted TaskStatus = "completed"
)
```

- `TaskStatus`: Custom type representing the state of a task list item.
- `TaskStatusActive`: Indicates an unchecked/open task (`[ ]`).
- `TaskStatusCompleted`: Indicates a checked/completed task (`[x]` or `[X]`).

### Internal Attribute Constants

- `taskStatusAttributeName = "task-status"`: The AST attribute key attached to a `gast.ListItem` node storing its `TaskStatus`.

---

## Helper Functions

### `IsTask`

```go
func IsTask(node gast.Node) bool
```

- **Purpose**: Checks whether a given AST node is a task list item.
- **Logic**: Returns `true` if `node` is a `*gast.ListItem` and contains the `task-status` attribute.

### `TaskStatusOf`

```go
func TaskStatusOf(node gast.Node) (TaskStatus, bool)
```

- **Purpose**: Retrieves the `TaskStatus` of a given AST node.
- **Logic**: 
  1. Verifies the node is a `*gast.ListItem`.
  2. Looks up the `task-status` attribute on the node.
  3. Returns the extracted `TaskStatus` and `true` if found; otherwise returns `""` and `false`.

---

## Parsing Architecture

The task list extension handles checkbox inline syntax through an inline parser attached to Goldmark's parsing pipeline.

### Regular Expression

```go
var taskCheckboxRegexp = regexp.MustCompile(`^\[([\sxX])\]\s*`)
```
Matches checkbox markers at the start of a line:
- `[` followed by whitespace (`\s`), lower `x`, or upper `X`, followed by `]` and optional trailing whitespace.

### `taskListItemParser`

Internal struct implementing `parser.InlineParser`.

#### Methods:
* **`Trigger() []byte`**: Returns `[]byte{'['}`. The parser triggers when encountering an opening square bracket.
* **`Parse(parent gast.Node, block text.Reader, _ parser.Context) gast.Node`**:
  * **AST Context Validation**: Validates that the current `parent` node is the first child paragraph of a `*gast.ListItem`, that `parent` currently has no children, and that the `ListItem` does not already have a `task-status` attribute.
  * **Pattern Matching**: Peeks at the line buffer using `taskCheckboxRegexp`.
  * **State Attribution**:
    * Advances the text reader past the matched checkbox text.
    * Sets `TaskStatusCompleted` on the `ListItem` if the matched character is `'x'` or `'X'`.
    * Sets `TaskStatusActive` on the `ListItem` if the matched character is whitespace.
  * Returns `parser.Nil`.
* **`CloseBlock(_ gast.Node, _ parser.Context)`**: No-op implementation required by interface.

### `taskListItemParserExtension`

Struct implementing `parser.Extension`.

- **`NewTaskListItemParser() parser.Extension`**: Constructor function returning a new parser extension.
- **`ParserOptions(_ *parser.Config) []parser.Option`**: Registers `newTaskListItemParser()` as an inline parser with priority `0`.
- **`TaskListItemParser`**: Global exported instance initialized with `NewTaskListItemParser()`.

---

## Rendering Architecture

Task list items require overriding paragraph node rendering to prepend disabled HTML checkbox input elements before paragraph content.

### Configuration and Options

- **`taskListItemHTMLRendererConfig`**:
  - `XHTML bool`: Controls whether `<input>` tags are self-closing (`/>`).
  - `IsInTightBlockFunc html.IsInTightBlockFunc`: Function determining if a paragraph block is inside a tight list.
- **`TaskListItemHTMLRendererOption`**: Interface for configuring the HTML renderer extension.

### `taskListItemHTMLRendererExtension`

Struct implementing `html.Extension`.

- **`NewTaskListItemHTMLRenderer(opts ...TaskListItemHTMLRendererOption) html.Extension`**: Constructor configuring and returning an HTML renderer extension.
- **`RendererOptions(c *html.Config) []html.Option`**:
  - Copies XHTML configuration and tight-block evaluation functions from the provided `html.Config`.
  - Registers `renderParagraph` as the custom renderer for `gast.KindParagraph`.
- **`TaskListItemHTMLRenderer`**: Global exported instance initialized with `NewTaskListItemHTMLRenderer()`.

### Rendering Logic

#### `renderInputTag`
Writes the `<input>` HTML element into the buffer:
- If `status == TaskStatusCompleted`: Writes `<input checked="" disabled="" type="checkbox"`
- If `status == TaskStatusActive`: Writes `<input disabled="" type="checkbox"`
- Closes the tag with ` /> ` if `XHTML` configuration is `true`, or standard `> ` otherwise.

#### `renderParagraph`
Replaces default paragraph rendering to handle both standard paragraphs and task-item paragraphs:

```go
func (r *taskListItemHTMLRendererExtension) renderParagraph(
	writer io.Writer, source []byte, node gast.Node, entering bool, rc renderer.Context) (gast.WalkStatus, error)
```

1. Checks if the current paragraph node (`n`) is the first child of a task list item (retrieves status via `TaskStatusOf`).
2. **If NOT a task item**:
   - Renders standard `<p>` and `</p>` tags (or suppresses outer tags if `inTight` is true).
3. **If a task item**:
   - **Tight List (`inTight == true`)**:
     - When entering: Calls `renderInputTag` directly (no `<p>` tags rendered).
     - When exiting: Emits a newline `\n` if there is a next sibling and child content.
   - **Loose List (`inTight == false`)**:
     - When entering: Renders standard `<p>` tag (including attributes if present), then calls `renderInputTag`.
     - When exiting: Renders `</p>\n`.

---

## Workflow Summary

```
Markdown Input ("- [x] Done")
          │
          ▼
   [Goldmark Parser]
          │
  taskListItemParser triggers on '['
  - Validates parent is ListItem -> Paragraph
  - Matches regex ^\[([\sxX])\]\s*
  - Applies "task-status" = "completed" to ListItem
          │
          ▼
    [AST Created]
    ListItem (attr: task-status="completed")
      └─ Paragraph
          └─ Text ("Done")
          │
          ▼
  [Goldmark HTML Renderer]
          │
  taskListItemHTMLRendererExtension intercepts Paragraph node
  - Detects parent is Task ListItem
  - Renders <input checked="" disabled="" type="checkbox">
  - Renders inline text and optional <p> wrappers based on tight/loose state
          │
          ▼
     HTML Output
     <li><input checked="" disabled="" type="checkbox"> Done</li>
```