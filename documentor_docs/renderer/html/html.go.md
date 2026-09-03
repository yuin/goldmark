# Technical Documentation: `renderer/html/html.go`

## Overview

The `html` package implements an HTML/XHTML renderer for AST (Abstract Syntax Tree) nodes produced by Goldmark. It translates Markdown AST representations into HTML or XHTML markup while providing features such as configurable line-break strategies, security filtering for raw HTML and URLs, custom attribute filtering, and tight-block paragraph detection.

---

## Architecture & Context

The rendering system relies on context-bound writer wrappers to sanitize and format output depending on what is being rendered (e.g., HTML text, URL attributes, or raw HTML content).

```
                      +-------------------+
                      |   htmlRenderer    |
                      +---------+---------+
                                |
                   +------------+------------+
                   |                         |
        +----------v----------+    +---------v---------+
        | Context Initialization |    |  commonMark Ext.  |
        +----------+----------+    +---------+---------+
                   |                         |
       +-----------+-----------+             | Custom AST Node Renderers
       |           |           |             | (Heading, Link, Image, etc.)
 +-----v----+ +----v-----+ +---v----+        |
 |htmlWriter| |textWriter| |linkURL |        | Filtered Attributes &
 +----------+ +----------+ | Writer |        | Security Checks
                           +--------+        v
                                       +-----+-----+
                                       | Output Stream |
                                       +-----------+
```

---

## Core Types and Configurations

### Configuration Structs

#### `Config`
Holds the configuration parameters for the HTML renderer:
- `HardWraps` (`bool`): When `true`, soft line breaks are rendered as `<br>` or `<br />`.
- `LineBreakStrategy` (`LineBreakStrategy`): Strategy used to determine whether soft line breaks produce a newline character.
- `XHTML` (`bool`): When `true`, self-closing elements render with XML-style closing tags (e.g., `<br />`, `<hr />`, `<img ... />`).
- `Unsafe` (`bool`): When `false` (default), raw HTML blocks and potentially dangerous URLs are suppressed or sanitized. When `true`, raw content is output directly.
- `Paragraph` (`ParagraphConfig`): Holds paragraph-specific configuration options.

#### `ParagraphConfig`
- `IsInTightBlockFunc` (`IsInTightBlockFunc`): Function used to determine whether a paragraph resides in a tight block (e.g., a tight list item) where `<p>` tags should be omitted.

### Default Configuration
`Config.Default()` initializes default settings:
- `HardWraps`: `false`
- `LineBreakStrategy`: `nil`
- `XHTML`: `false`
- `Unsafe`: `false`
- `Paragraph.IsInTightBlockFunc`: `IsInTightBlock`

---

## Type Aliases & Functional Options

The package provides standard aliases wrapping generic types from the underlying `renderer` package:

| Alias | Target Type |
| :--- | :--- |
| `Renderer` | `renderer.Renderer[io.Writer]` |
| `NodeRenderer` | `renderer.NodeRenderer[io.Writer]` |
| `NodeRendererDecorator` | `renderer.NodeRendererDecorator[io.Writer]` |
| `Extension` | `renderer.Extension[Config]` |
| `Option` | `renderer.Option[Config]` |

### Functional Options

- `WithNodeRenderers(nodeRenderers map[ast.NodeKind]NodeRenderer) Option`
- `WithNodeRenderer(kind ast.NodeKind, nodeRenderer NodeRenderer) Option`
- `WithNodeRendererDecorators(decorators map[ast.NodeKind]NodeRendererDecorator) Option`
- `WithNodeRendererDecorator(kind ast.NodeKind, decorator NodeRendererDecorator) Option`
- `WithExtensions(extensions ...Extension) Option`
- `WithHardWraps() Option`: Enables `HardWraps`.
- `WithLineBreakStrategy(strategy LineBreakStrategy) Option`: Sets a custom soft-line break strategy.
- `WithXHTML() Option`: Enables XHTML output mode.
- `WithUnsafe() Option`: Enables rendering of raw HTML and untrusted URLs.
- `WithIsInTightBlockFunc(fn IsInTightBlockFunc) Option`: Overrides the default tight-block checking function.

---

## Tight Block Detection

```go
type IsInTightBlockFunc func(n ast.Node) bool
```

### `IsInTightBlock(n ast.Node) bool`
Determines if a paragraph AST node is a direct child of a list item within a tight list.
- Checks if `n.Parent()` is non-nil.
- Checks if the parent's parent (`gp`) is of type `*ast.List`.
- Returns `list.IsTight`.

---

## Line Break Strategies

### `LineBreakStrategy` Interface
```go
type LineBreakStrategy interface {
    SoftLineBreak(thisLastRune rune, siblingFirstRune rune) bool
}
```

### Strategy Implementations

#### 1. `SimpleEastAsianLineBreakStrategy` (`var SimpleEastAsianLineBreakStrategy`)
Based on Pandoc’s `east_asian_line_breaks`.
- Suppresses soft line breaks if **both** the preceding rune (`thisLastRune`) and the succeeding rune (`siblingFirstRune`) are East Asian Wide runes (`util.IsEastAsianWideRune`).

#### 2. `CSSText3LineBreakStrategy` (`var CSSText3LineBreakStrategy`)
Implements segment break transformation rules from CSS Text Module Level 3 with enhancements for CJK text:
1. **Rule 1**: If either rune is a Zero-Width Space (`U+200B`), line break is removed (`returns false`).
2. **Rule 2**: If both runes have an East Asian Width property of "F", "W", or "H", and neither side is Hangul script, line break is removed unless either character is Hangul (`returns unicode.Is(unicode.Hangul, ...)`).
3. **Rule 3**: If either rune is in the space-discarding set, Unicode Punctuation (`P*`), or `U+3000` (Ideographic Space), line break is removed (`returns false`).
4. **Rule 4**: Otherwise, converts break to space (`returns true`).

---

## Special Output Writers

The rendering system wraps `util.BufWriter` instances to apply contextual encoding during rendering.

```
       [Raw Stream Data]
               |
  +------------+------------+
  |                         |
  v                         v
[htmlWriter]          [textWriter]
 Replaces '\u0000'     Applies HTML escaping
 with '\ufffd'         via util.EscapeHTMLByte
```

### 1. `htmlWriter`
Wraps output and replaces NULL bytes (`\u0000`) with Unicode replacement characters (`\ufffd`).

### 2. `textWriter`
Performs HTML escaping on input bytes using `util.EscapeHTMLByte` (`<`, `>`, `&`, `"`).

### 3. `linkURLWriter`
Performs URL escaping using `util.URLEscape`.

### Context Retrievers
Functions used inside `NodeRenderer` functions to access the initialized writers from `renderer.Context`:
- `ContextHTMLWriter(rc renderer.Context) util.BufWriter`
- `ContextTextWriter(rc renderer.Context) util.BufWriter`
- `ContextLinkURLWriter(rc renderer.Context) util.BufWriter`

*Note: These functions panic if the corresponding writer is not present in the context.*

---

## Main Renderer Initialization and Invocation

### `New(opts ...Option) Renderer`
1. Instantiates a CommonMark extension (`NewCommonMark()`).
2. Appends `WithExtensions(cm)` to options.
3. Sets up an `OnBeforeRender` hook to initialize `htmlWriter`, `textWriter`, and `linkURLWriter` inside the execution context (`renderer.Context`).
4. Constructs and returns `htmlRenderer`.

### Rendering Methods
- `Render(w io.Writer, source []byte, n ast.Node, opts ...renderer.RenderOption) error`: Ensures `w` implements `util.ErrorBufWriter` (allocates buffer `len(source)*3` if needed) and triggers helper execution.
- `RenderStringSource(w io.Writer, source string, n ast.Node, opts ...renderer.RenderOption) error`: Converts string source to byte slice and invokes `Render`.

---

## CommonMark Extension & AST Node Handlers

The default HTML rendering rules are registered via `NewCommonMark(opts ...Option) Extension`.

### Registered Node Kind Handlers

| AST Node Kind | Handling Logic |
| :--- | :--- |
| `ast.KindDocument` | Continues walk. |
| `ast.KindHeading` | Outputs `<h1-6>` and `</h1-6>\n` with filtered attributes. |
| `ast.KindBlockquote` | Outputs `<blockquote>\n` or `<blockquote [attrs]>`. |
| `ast.KindCodeBlock` | Outputs `<pre><code class="language-...">` escaping code text via `textWriter`. |
| `ast.KindHTMLBlock` | Outputs raw HTML if `Unsafe` is `true`; otherwise outputs `<!-- raw HTML omitted -->\n`. |
| `ast.KindList` | Outputs `<ul>` or `<ol>`. Sets `start` attribute if ordered and start != 1. |
| `ast.KindListItem` | Outputs `<li>`. Checks if child paragraph is tight to format newlines properly. |
| `ast.KindParagraph` | Omits `<p>` tags if inside a tight block (`IsInTightBlockFunc`). Otherwise outputs `<p>...</p>\n`. |
| `ast.KindThematicBreak` | Outputs `<hr>` or `<hr />\n` if `XHTML` is enabled. |
| `ast.KindAutoLink` | Renders `<a href="...">` checking destination safety via `IsDangerousURL`. |
| `ast.KindCodeSpan` | Outputs `<code>...</code>`, escaping code text via `textWriter`. |
| `ast.KindEmphasis` | Outputs `<em>...</em>`. |
| `ast.KindStrong` | Outputs `<strong>...</strong>`. |
| `ast.KindLink` | Outputs `<a href="..." title="...">...</a>`. Checks URL safety. |
| `ast.KindImage` | Outputs `<img src="..." alt="..." title="..." />` (or `>`). Skips children. |
| `ast.KindRawHTML` | Outputs raw content via `htmlWriter` if `Unsafe` is `true`; otherwise `<!-- raw HTML omitted -->`. |
| `ast.KindText` | Writes text via `textWriter`. Handles hard line breaks (`<br>`) and soft line breaks via `LineBreakStrategy` or `HardWraps`. |

---

## HTML Attribute Filtering

Attributes are written using `RenderAttributes(writer, source, node, filter, rc)`. Attributes are excluded unless included in the specified filter or prefixed with `data-`.

### Global & Element-Specific Attribute Filters

- **`GlobalAttributeFilter`**:
  `accesskey`, `autocapitalize`, `autofocus`, `class`, `contenteditable`, `dir`, `draggable`, `enterkeyhint`, `hidden`, `id`, `inert`, `inputmode`, `is`, `itemid`, `itemprop`, `itemref`, `itemscope`, `itemtype`, `lang`, `part`, `role`, `slot`, `spellcheck`, `style`, `tabindex`, `title`, `translate`
- **`HeadingAttributeFilter`**: Inherits `GlobalAttributeFilter`.
- **`BlockquoteAttributeFilter`**: Global + `cite`.
- **`ListAttributeFilter`**: Global + `start`, `reversed`, `type`.
- **`ListItemAttributeFilter`**: Global + `value`.
- **`ParagraphAttributeFilter`**: Inherits `GlobalAttributeFilter`.
- **`ThematicAttributeFilter`**: Global + `align`, `color`, `noshade`, `size`, `width`.
- **`LinkAttributeFilter`**: Global + `download`, `href`, `lang`, `media`, `ping`, `referrerpolicy`, `rel`, `shape`, `target`.
- **`CodeAttributeFilter`**: Inherits `GlobalAttributeFilter`.
- **`EmphasisAttributeFilter`**: Inherits `GlobalAttributeFilter`.
- **`StrongAttributeFilter`**: Inherits `GlobalAttributeFilter`.
- **`ImageAttributeFilter`**: Global + `align`, `border`, `crossorigin`, `decoding`, `height`, `importance`, `intrinsicsize`, `ismap`, `loading`, `referrerpolicy`, `sizes`, `srcset`, `usemap`, `width`.

---

## Security Utilities

### `IsDangerousURL(url string) bool`

Validates URLs when `Unsafe` mode is disabled (`false`).

- **Safe Schemes / Protocols**:
  - Standard HTTP/HTTPS, relative paths, mailto, etc. (do not match dangerous prefixes).
  - Data URIs containing images (`data:image/`) if followed specifically by `png;`, `gif;`, `jpeg;`, or `webp;`.
- **Dangerous Schemes (returns `true`)**:
  - `javascript:`
  - `vbscript:`
  - `file:`
  - `data:` (unless explicitly allowed image MIME types listed above)