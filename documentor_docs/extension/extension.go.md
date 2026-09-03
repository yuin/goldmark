# Technical Documentation: `extension/extension.go`

## Overview

The `extension/extension.go` file defines functional options used to configure HTML rendering behavior for standard Goldmark markdown extensions (such as tables, task lists, and footnotes).

It uses Go's functional options pattern to modify renderer configurations (`tableHTMLRendererConfig`, `taskListItemHTMLRendererConfig`, and `footnoteHTMLRendererConfig`) without exposing internal state directly.

---

## Dependencies

*   **`github.com/yuin/goldmark/v2/renderer/html`**: Provides the `html.IsInTightBlockFunc` type used to determine block context during HTML rendering.

---

## Key Components

### 1. `WithXHTML`

#### Function Signature
```go
func WithXHTML() interface {
    TableHTMLRendererOption
    TaskListItemHTMLRendererOption
    FootnoteHTMLRendererOption
}
```

#### Description
`WithXHTML` is a constructor function that returns a functional option indicating that table, task list, and footnote elements should be rendered using XHTML self-closing tags/formatting.

#### Return Value
Returns a pointer to a `withXHTML` struct (`&withXHTML{value: true}`) which satisfies an anonymous interface composed of three option interfaces:
*   `TableHTMLRendererOption`
*   `TaskListItemHTMLRendererOption`
*   `FootnoteHTMLRendererOption`

---

### 2. `withXHTML` Struct and Methods

#### Struct Definition
```go
type withXHTML struct {
    value bool
}
```

#### Methods

*   **`applyTableHTMLRendererOption(c *tableHTMLRendererConfig)`**
    *   **Behavior**: Sets the `XHTML` field on the given `tableHTMLRendererConfig` to `o.value`.
*   **`applyTaskListItemHTMLRendererOption(c *taskListItemHTMLRendererConfig)`**
    *   **Behavior**: Sets the `XHTML` field on the given `taskListItemHTMLRendererConfig` to `o.value`.
*   **`applyFootnoteHTMLRendererOption(c *footnoteHTMLRendererConfig)`**
    *   **Behavior**: Sets the `XHTML` field on the given `footnoteHTMLRendererConfig` to `true`.

---

### 3. `WithIsInTightBlockFunc`

#### Function Signature
```go
func WithIsInTightBlockFunc(f html.IsInTightBlockFunc) interface {
    TaskListItemHTMLRendererOption
}
```

#### Description
`WithIsInTightBlockFunc` is a constructor function that configures a custom function used to check if a task list item resides within a tight block context.

#### Parameters
*   `f html.IsInTightBlockFunc`: A function matching Goldmark's HTML renderer signature for evaluating tight block status.

#### Return Value
Returns a pointer to a `withIsInTightBlockFunc` struct (`&withIsInTightBlockFunc{IsInTightBlockFunc: f}`) which satisfies an anonymous interface containing:
*   `TaskListItemHTMLRendererOption`

---

### 4. `withIsInTightBlockFunc` Struct and Methods

#### Struct Definition
```go
type withIsInTightBlockFunc struct {
    IsInTightBlockFunc html.IsInTightBlockFunc
}
```

#### Methods

*   **`applyTaskListItemHTMLRendererOption(c *taskListItemHTMLRendererConfig)`**
    *   **Behavior**: Sets the `IsInTightBlockFunc` field on the target `taskListItemHTMLRendererConfig` to `o.IsInTightBlockFunc`.

---

## How It Works

1. **Option Instantiation**: Callers invoke public options like `WithXHTML()` or `WithIsInTightBlockFunc(...)`.
2. **Interface Satisfaction**: The functions instantiate internal struct pointers (`&withXHTML{...}` or `&withIsInTightBlockFunc{...}`).
3. **Configuration Application**: When passed to extension renderers, the renderer calls the respective `apply*Option` methods to set internal fields (`XHTML` or `IsInTightBlockFunc`) on configuration structs (`tableHTMLRendererConfig`, `taskListItemHTMLRendererConfig`, or `footnoteHTMLRendererConfig`).