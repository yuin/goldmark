## Breaking changes in v2

### Module path

The module path has changed from `github.com/yuin/goldmark` to `github.com/yuin/goldmark/v2`.

### Top-level `goldmark` package

The `goldmark.Markdown` interface, `goldmark.New()`, `goldmark.Convert()`, and the `goldmark.Extender` interface have been removed.
Use `parser.New()` and `html.New()` (or another renderer) directly.

```go no-run
// v1
import "github.com/yuin/goldmark"
md := goldmark.New(goldmark.WithExtensions(...))
md.Convert(source, &buf)

// v2
import (
    "github.com/yuin/goldmark/v2/parser"
    "github.com/yuin/goldmark/v2/renderer/html"
)
p := parser.New(parser.WithExtensions(...))
r := html.New(html.WithExtensions(...))
doc := p.Parse(source)
r.Render(&buf, source, doc)
```

### Extension pattern

In v1, extensions implemented the `goldmark.Extender` interface with a single `Extend(goldmark.Markdown)` method that configured both the parser and renderer.

In v2, parser extensions implement `parser.Extension` (returns `[]parser.Option`) and renderer extensions implement `renderer.Extension[C]` (returns `[]renderer.Option[C]`). These are passed separately to `parser.New()` and `html.New()`.

```go no-run
// v1
type MyExtension struct{}
func (e *MyExtension) Extend(m goldmark.Markdown) {
    m.Parser().AddOptions(...)
    m.Renderer().AddOptions(...)
}

// v2: split into parser extension and renderer extension
type MyParserExtension struct{}
func (e *MyParserExtension) ParserOptions(c *parser.Config) []parser.Option { ... }

type MyHTMLRendererExtension struct{}
func (e *MyHTMLRendererExtension) RendererOptions(c *html.Config) []html.Option { ... }
```

### `renderer` package

The renderer is now generic over the writer type. The main interfaces are now `renderer.Renderer[W any]` and `renderer.NodeRenderer[W any]`.

In v1, `NodeRenderer` implemented `RegisterFuncs(NodeRendererFuncRegisterer)` to register `NodeRendererFunc` callbacks. In v2, use `renderer.WithNodeRenderer(kind, nodeRenderer)` or `renderer.WithNodeRenderers(map[ast.NodeKind]NodeRenderer)` options directly.

The v1 signature `NodeRendererFunc func(writer util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error)` is replaced by a generic `renderer.NodeRendererFunc[W any]`. For HTML rendering, `W` is `io.Writer`.

`renderer.Renderer[W].Render` now takes a `renderer.RenderOption`s.

A `renderer.NodeRendererDecorator[W any]` is new in v2, allowing you to run code before and after the render pass.

### `renderer/html` package

`html.NewRenderer(opts ...Option) renderer.NodeRenderer` has been replaced by `html.New(opts ...Option) Renderer`.

`html.RenderAttributes` is still a free function, but its signature changed from v1's `RenderAttributes(w util.BufWriter, node ast.Node, filter util.BytesFilter)` to v2's `RenderAttributes(writer io.Writer, source []byte, node ast.Node, filter util.BytesFilter, rc renderer.Context)` — it now takes `source` explicitly, since attribute values are resolved from it rather than pre-decoded.

`html.WithEastAsianLineBreaks` has been removed. Use `html.WithLineBreakStrategy` instead.

### `ast.Node` interface

- `Type() NodeType` and the `NodeType` type (with constants `TypeBlock`, `TypeInline`, `TypeDocument`) have been removed. Use type assertions to `ast.BlockNode` or `ast.InlineNode` instead.
- `Text(source []byte) []byte` (was already deprecated in v1) has been removed.
- `HasBlankPreviousLines()`, `SetBlankPreviousLines()`, and `Lines()`/`SetLines()` (renamed `Source()`/`SetSource()`, plus a new `AppendSource()`) have been removed from `Node` and moved to the new `BlockNode` interface (see below).
- `IsRaw() bool` has been removed entirely, with no replacement on any interface. Raw/unparsed block content (e.g. `HTMLBlock`, `CodeBlock`) is now identified purely by node kind, not by a marker method.
- Tree mutation methods (`AppendChild`, `RemoveChild`, `RemoveChildren`, `InsertBefore`, `InsertAfter`, `ReplaceChild`) no longer take a `self Node` as their first argument.
- `Dump` now returns `*NodeDump`. `Dump(source []byte) *NodeDump` is the new signature.
- `ast.Attribute` uses `string` names instead of `[]byte`.
  - `SetAttributeString` and `AttributeString` has been removed.
  - `SetAttribute` and `Attribute` now take `string` names instead of `[]byte`.
- Attributes other than string type are no longer supported.
  - `goldmark_v1_attribute` build tag allows using v1-compatible attributes.
    - Under this build tag, the `text.MultiLineValue` returned by `Node.Attribute(name)` gains an `Any(source []byte) any` method to get the parsed value:
      ```go no-run
      attr, ok := node.Attribute("data-count")
      v := attr.Any(source) // returns float64 if the attribute value is a number
      ```

### New `BlockNode` and `InlineNode` interfaces

`ast.BlockNode` extends `Node` with block-specific behaviour:
- `HasBlankPreviousLines() bool` / `SetBlankPreviousLines(bool)`
- `Source() []text.Segment` / `SetSource([]text.Segment)` (replaces `Lines() *text.Segments`)

`ast.InlineNode` extends `Node` as a marker interface for inline nodes.

### `ast.BaseNode.Init()` must be called in every custom node constructor

`BaseNode` now stores a `self` reference to support argument-free tree mutation methods. Call `n.Init(n)` in every node constructor, including those of custom extension nodes.

### Removed and merged AST nodes

| Removed (v1) | Replacement (v2) |
|---|---|
| `ast.TextBlock` | Removed (was only used internally by the parser) |
| `ast.FencedCodeBlock` | Merged into `ast.CodeBlock`; distinguish via `CodeBlock.CodeBlockKind` (`CodeBlockKindIndented` / `CodeBlockKindFenced`) |
| `extension/ast.TaskCheckBox` | Removed; task state is stored as an attribute on `ListItem` |

`KindTextBlock` and `KindFencedCodeBlock` no longer exist.

### Changed AST node fields and constructors

**`ast.Text`**

- `Segment text.Segment` → `Value text.SingleLineValue`
- All constructors are replaced by a single `NewText(v text.SingleLineValue) *Text`. Build the value first with the `text` package constructors (e.g. `text.NewSingleLineValueFromSegment(seg, decoder)`, `text.NewSingleLineValueFromString(s, decoder)`), then pass it to `NewText`.
- `SoftLineBreak()`/`SetSoftLineBreak(bool)` and `HardLineBreak()`/`SetHardLineBreak(bool)` are unchanged. `IsRaw()`/`SetRaw(bool)` are removed (see [`ast.Node` interface](#astnode-interface)); "raw" text is now expressed by binding `text.IdentityDecoder`(or `text.CodeSpanDecoder`) when constructing the node's `text.SingleLineValue`.

**`ast.String` (inline node) — removed**

Use `ast.NewText(text.NewSingleLineValueFromString(s, decoder))` instead, or the `ast.N(...)` builder helper (see [Usage](#usage)) for constructing literal-string trees.

**`ast.Emphasis`**

- `Level int` field removed. `*Emphasis` always represents single emphasis (`*`/`_`), `*Strong` always represents strong emphasis (`**`/`__`). They are now separate types.
- `NewEmphasis(level int)` → `NewEmphasis()`

**`ast.CodeSpan`**

- No longer holds `Text` child nodes. Now has `Value text.MultiLineValue`.
- `NewCodeSpan()` → `NewCodeSpan(value text.MultiLineValue)`

**`ast.RawHTML`**

- Now has `Value text.MultiLineValue` (no children).
- `NewRawHTML()` → `NewRawHTML(value text.MultiLineValue)`

**`ast.Heading`**

- Added `HeadingKind HeadingKind` (`HeadingKindATX` or `HeadingKindSetext`).
- `NewHeading(level int)` → `NewHeading(level int, kind HeadingKind)`

**`ast.CodeBlock` (fenced)**

- `FencedCodeBlock` is merged; `NewFencedCodeBlock(info *Text)` is gone.
- `Info` is now `text.SingleLineValue` (not `*Text`); `Value text.Lines` holds the code body.
- `NewCodeBlock(kind CodeBlockKind)` → `NewCodeBlock(kind CodeBlockKind, value text.Lines, opts ...CodeBlockOption)`
- Optional info string: `ast.WithCodeBlockInfo(info)`
- `Language(source []byte) []byte` → `Language(source []byte) (string, bool)` (the returned `bool` reports whether a non-empty language token was found in the info string, distinguishing "no language" from "language is the empty string")

**`ast.Link` and `ast.Image`**

- `Destination []byte` → `Destination text.SingleLineValue`
- `Title []byte` → `Title text.MultiLineValue`
- `NewLink()` → `NewLink(destination text.SingleLineValue, opts ...LinkOption)`
- `NewImage(link *Link)` → `NewImage(destination text.SingleLineValue, opts ...LinkOption)`
- Optional title: `ast.WithLinkTitle(title)`
- Optional reference: `ast.WithLinkReference(kind, value)`

**`ast.AutoLink`**

- `AutoLinkType AutoLinkType`, `Protocol []byte`, and `value *Text` removed.
- `URL(source []byte) []byte` and `Label(source []byte) []byte` methods removed.
- Now has three `text.SingleLineValue` fields: `Destination` (full href, email includes `mailto:`), `Label` (display text), `Text` (raw source text).
- `NewAutoLink(typ AutoLinkType, value *Text)` → `NewAutoLink(destination, label text.SingleLineValue, opts ...AutoLinkOption)`
- Optional source text: `ast.WithAutoLinkText(text)`

**`ast.ReferenceLink`**

- `Type ReferenceLinkType` → `ReferenceLinkKind ReferenceLinkKind`
- Constants renamed: `ReferenceLinkFull/Collapsed/Shortcut` → `ReferenceLinkKindFull/Collapsed/Shortcut`
- `Value []byte` → `Value text.MultiLineValue`

**`ast.HTMLBlock`**

- `HTMLBlockType` type renamed to `HTMLBlockKind`.
- Constants renamed: `HTMLBlockType1`..`HTMLBlockType7` → `HTMLBlockKind1`..`HTMLBlockKind7`.
- `ClosureLine text.Segment` field removed. The closing delimiter line is now folded into the unified `Value text.Lines` field along with the rest of the block's content.
- `HasClosure() bool` and `IsRaw() bool` methods removed (the latter follows the general `IsRaw()` removal — see [`ast.Node` interface](#astnode-interface)).
- `NewHTMLBlock(typ HTMLBlockType)` → `NewHTMLBlock(kind HTMLBlockKind)`

**`ast.ListItem`**

- `Offset int` (public field) → unexported; access via `Offset() int` / `SetOffset(int)`.
- `NewListItem(offset int)` → `NewListItem()`

**`ast.LinkReferenceDefinition`**

- `Label/Destination/Title []byte` → `Label text.MultiLineValue`, `Destination text.SingleLineValue`, `Title text.MultiLineValue`
- `NewLinkReferenceDefinition(label, destination, title []byte)` → `NewLinkReferenceDefinition(label text.MultiLineValue, destination text.SingleLineValue, opts ...LinkReferenceDefinitionOption)`
- Title is now optional and passed as an option: `ast.WithLinkTitle(title)` — the same generic helper used for `Link`/`Image` — also satisfies `LinkReferenceDefinitionOption`

**`extension/ast.DefinitionList`**

- `Offset int` and `TemporaryParagraph *Paragraph` (public fields) → unexported; access via `Offset()` / `SetOffset()` / `TemporaryParagraph()` / `SetTemporaryParagraph()`.
- `NewDefinitionList(offset int, para *Paragraph)` → `NewDefinitionList()`

**`extension/ast.TableCell`**

- `NewTableCell()` → `NewTableCell(alignment Alignment)` (alignment is now a required argument)

**`extension/ast.TableHeader`**

- `NewTableHeader(row *TableRow)` → `NewTableHeader()` (child nodes must be moved manually)

**`extension/ast.Table`, `extension/ast.TableRow`, `extension/ast.TableHeader`**

- The `Alignments []Alignment` field is removed from all three types. Column alignment is now tracked purely per-cell via `TableCell.Alignment` (see `NewTableCell(alignment Alignment)` above) — there is no longer a table- or row-level alignment list to keep in sync.

**`extension/ast.TableBody` — new**

- New node type wrapping the non-header rows of a table, with kind `KindTableBody` and constructor `NewTableBody()`. `Table`'s children are now `TableHeader` followed by a single `TableBody` (which itself holds the `TableRow` children), rather than `TableHeader` followed directly by `TableRow` siblings.

**`extension/ast` Footnotes — renamed and restructured**

| v1 | v2 | Notes |
|---|---|---|
| `FootnoteLink` / `NewFootnoteLink(index int)` | `FootnoteReference` / `NewFootnoteReference(label text.SingleLineValue)` | Gains a `Label text.SingleLineValue` field and is now constructed from that label instead of a pre-resolved index (`Index`/`RefIndex` are kept, `RefCount` is dropped) |
| `FootnoteBacklink` / `NewFootnoteBacklink(index int)` | Removed, no replacement | The backlink anchor is generated directly by the HTML renderer instead of being a distinct AST node |
| `Footnote` / `NewFootnote(ref []byte)` | `FootnoteDefinition` / `NewFootnoteDefinition(label text.SingleLineValue)` | |
| `FootnoteList` / `NewFootnoteList()` | Removed, no replacement | Footnote definitions are tracked via the new `extension.Footnotes` parser-context interface (`extension.ContextFootnotes(pc)`) instead of being collected under a dedicated list node |

### `text` package

The `text.Segments` type (`*Segments` holding `[]Segment`) is no longer part of the public `Node` API.

New types for representing text values:
- `text.Value` — an interface for a single-line value or a multi-line value
- `text.SingleLineValue` — a single contiguous source span or a literal string
- `text.Index` — a raw `(Start, Stop)` index pair
- `text.MultiLineValue` — a value that may span multiple source lines
- `text.Lines` — a list of source `Segment`s for block-level content (e.g. code blocks)

In v1, `ast.Text`/`ast.String` and friends held a raw `[]byte`/`Segment` pointing at the source, and decoding (backslash escapes, numeric references, entity names) happened ad hoc wherever a renderer wrote that value out — e.g. `util.UnescapePunctuations`, `util.ResolveNumericReferences`, and `util.ResolveEntityNames` were called directly from renderer code, mixed together with HTML-escaping in `html.Writer.Write`.

In v2, decoding is a first-class, pluggable step performed once, at AST-construction time, via the new `text.Decoder` interface — not at render time, and not via those removed `util` functions:
- `text.NewSingleLineValue`/`text.NewMultiLineValue` and their `...FromIndex`/`...FromIndices`/`...FromString` variants take a `text.Decoder` argument, which is applied when `Value.Value(source []byte) string` is later called. `text.ValueBuilder.Decoder(d Decoder) *ValueBuilder` sets the decoder used by `BuildSingleLine`/`BuildMultiLine`/`Build` (defaults to `text.IdentityDecoder` if never called).
- `text.IdentityDecoder` is a decoder that returns its input unchanged; bind it explicitly when constructing raw/undecoded values (e.g. raw HTML content).
- `text.Reader`/`text.BlockReader` hold the `text.Decoder` used for parsing; `NewReader`/`NewBlockReader` take a `decoder Decoder` argument, and `Decoder() Decoder` returns it — this is what block/inline parsers pass into the `text.Value` constructors above so that node values are already bound to the right decoder.
- Rendering no longer decodes at all: it only has to choose the right HTML-safety writer for an already-decoded `text.Value` (see [Writing text values safely](#writing-text-values-safely-textvalue-and-context-writers)).

`text.Reader.FindClosure()` and `text.FindClosureOptions` have been removed (they were moved to parser-internal use only).

### `parser` package

`parser.Parser.Parse(reader, opts ...ParseOption)` has been simplified. `reader` is now `source []byte`.

- `parser.NewParser(options ...Option) Parser` → `parser.New(options ...Option) Parser`.
- `parser.Reference` / `parser.NewReference(label, destination, title []byte) Reference` → `parser.LinkDefinition` / `parser.NewLinkDefinition(label, destination, title []byte) LinkDefinition`. `parser.Context`'s `AddReference`/`Reference`/`References` methods are renamed to `AddLinkDefinition`/`LinkDefinition`/`LinkDefinitions` to match.
- `parser.IDs` was an interface in v1; it is now a concrete `*IDs` struct returned by `parser.NewIDs(opts ...IDsOption)`. Custom ID generation is now a separate `parser.IDGenerator` interface, plugged in via `parser.WithIDGenerator(gen IDGenerator)`.
- `parser.DefaultBlockParsers()`, `parser.DefaultInlineParsers()`, and `parser.DefaultParagraphTransformers()` have been removed. Default parsing behavior is now bundled into a `parser.Extension` — `parser.CommonMark` (or `parser.NewCommonMark(opts ...Option)`) — which `parser.New()` wires in automatically. Use `parser.WithDefaultParsers(false)` to opt out of it (e.g. to build a parser from scratch with only your own parsers).
- `ScanDelimiter(line []byte, before rune, minimum int, processor DelimiterProcessor) *Delimiter` is renamed and re-signatured to `ParseDelimiter(block text.Reader, minimum int, processor DelimiterProcessor, pc Context) *Delimiter` — it now advances a `text.Reader` directly instead of being handed a raw `line []byte`/`before rune`. New helper functions `IsLeftFlankingDelimiterRun`/`IsRightFlankingDelimiterRun` expose the CommonMark delimiter-run classification directly, for parsers that need it without going through a full `parser.Delimiter`.
- The `parser.Attribute`/`parser.Attributes` types (which supported `[]byte` names, and values that could be numbers, arrays, or nested attribute objects with comma-separated lists) are removed. Attributes are now always `ast.Attribute{Name string, Value text.MultiLineValue}` — string names and text values only.
  - `ParseAttributes(reader text.Reader) (Attributes, bool)` → `ParseAttributes(reader text.Reader) ([]ast.Attribute, bool)`.
  - The `goldmark_v1_attribute` build tag (in `parser/attribute_v1.go`) restores the v1-compatible typed/comma-separated behavior for projects that depend on it.
    - Under this build tag, the `text.MultiLineValue` returned by `Node.Attribute(name)` gains an `Any(source []byte) any` method to get the parsed value:
      ```go no-run
      attr, ok := node.Attribute("data-count")
      v := attr.Any(source) // returns float64 if the attribute value is a number
      ```
- New: `parser.WithPrettyPrint(opts ...ast.PrettyPrintOption) ParseOption` prints the parsed AST tree for debugging (see [Parse options](#parse-options)).

### `util` package

- `util.UnescapePunctuations`, `util.ResolveNumericReferences`, `util.ResolveEntityNames` has been removed.
  - use `text.Decoder` instead.
- `util.IsEscapedPunctuation`, `util.DedentPosition`, `util.DedentPositionPadding`, `util.FindClosure`, `util.FindURLIndex`, and `util.FindEmailIndex` have also been removed, with no direct replacement (equivalent logic now lives inside the parser package or the relevant extension). `util.IndentPosition`/`util.IndentPositionPadding` are unaffected and remain unchanged.
- Second argument of `util.URLEscape` has been removed.
  - If you need to decode values before escaping, use `text.Decoder` to decode first, then `util.URLEscape` to escape.
- `util.BufWriter` no longer has `Available() int` and `Buffered() int` methods; it is now just `io.Writer` plus `WriteByte`, `WriteRune`, `WriteString`, and `Flush`.
- `util.PrioritizedValue`/`util.PrioritizedSlice` are now generic: `util.PrioritizedValue[T any]{Value T; Priority int}` and `util.PrioritizedValues[T comparable]`, with `.Sort()`/`.Remove(v T)` methods. `util.Prioritized(v T, priority int)` remains the constructor.
- New: `util.BytesFilter` gained `AddString(st string)` and `ContainsString(st string) bool` methods, for filters keyed by string instead of `[]byte`.

### Task list extension

The `extension/ast.TaskCheckBox` inline node no longer exists. Task state is stored as a `text.MultiLineValue` attribute on the `ListItem` node. Use `extension.IsTask(node)` and `extension.TaskStatusOf(node)` to inspect task items.

### New in v2

This section is a flat index of public APIs that have **no v1 counterpart at all** — brand new packages, types, or functions. A rename, a re-signatured method, or a struct that gained/lost a field is a *change* to an existing v1 API, not a new one, so it's covered once in the package-by-package sections above and intentionally not repeated here.

**`ast`**
- `ast.N(node Node, children ...any) Node` — builder helper that appends child nodes (or strings) to a node, useful for programmatically constructing an AST.
- `ast.BlockNode` / `ast.InlineNode` interfaces for type-safe node categorization.
- `ast.NodeDump` / `ast.NewNodeDump(node Node, properties map[string]any) *NodeDump` — the struct now returned by `Node.Dump`.
- `ast.PrettyPrintOption` — options consumed by `parser.WithPrettyPrint`.
- Functional option types used by the new node constructors: `ast.LinkOption`, `ast.AutoLinkOption`, `ast.CodeBlockOption`, `ast.LinkReferenceDefinitionOption`, and their `ast.WithLinkTitle`/`ast.WithLinkReference`/`ast.WithAutoLinkText`/`ast.WithCodeBlockInfo` constructors.

**`text`**
- `text.Value` interface, `text.SingleLineValue`, `text.MultiLineValue`, `text.Index`, and `text.Lines` — the value types described in the [`text` package](#text-package) section above.
- `text.Decoder` interface, `text.NewDecoder(opts ...DecoderOption) *DefaultDecoder`, `text.IdentityDecoder`, and `text.ValueBuilder` for constructing values with an explicit decoder.
- `text.Reader.Decoder()` / `text.BlockReader.Decoder()`.
- Under the `goldmark_v1_attribute` build tag: `text.MultiLineValue.Any(source []byte) any`, for parsing a v1-style typed attribute value.

**`parser`**
- `parser.IDGenerator` interface and `parser.WithIDGenerator(gen IDGenerator)` option, for pluggable element-ID generation (paired with the now-struct `parser.IDs`, see [`parser` package](#parser-package)).
- `parser.CommonMark` / `parser.NewCommonMark(opts ...Option)` — the default CommonMark parsing behavior, expressed as an ordinary `parser.Extension` instead of being built into the parser unconditionally — and `parser.WithDefaultParsers(bool)` to opt out of it.
- `parser.IsLeftFlankingDelimiterRun(before, after rune) bool` / `parser.IsRightFlankingDelimiterRun(before, after rune) bool` — CommonMark delimiter-run classification, exposed directly for parsers that don't need a full `parser.Delimiter`.
- `parser.WithPrettyPrint(opts ...ast.PrettyPrintOption) ParseOption` — prints the parsed AST tree for debugging (see [Parse options](#parse-options)).
- `parser.Parser.ParseStringSource(source string, opts ...ParseOption) ast.Node` convenience method.
- The `goldmark_v1_attribute` build tag (`parser/attribute_v1.go`) restoring v1-compatible attribute parsing for projects that depend on it.

**`renderer`**
- `renderer.NodeRendererDecorator[W any]` for decorating a node renderer.
- `renderer.RenderOption` and `renderer.Renderer[W].RenderStringSource(w W, source string, n ast.Node, opts ...RenderOption) error` convenience method.

**`renderer/html`**
- `html.ContextHTMLWriter(rc)` / `html.ContextTextWriter(rc)` / `html.ContextLinkURLWriter(rc)` — context-scoped `util.BufWriter`s for writing already-decoded `text.Value` content safely into HTML output (see [Writing text values safely](#writing-text-values-safely-textvalue-and-context-writers)).

**`extension`**
- `extension.WithXHTML()` and `extension.WithIsInTightBlockFunc(f)` — cross-cutting functional options that configure multiple extensions' HTML renderers at once (table, task list, and — for `WithXHTML` — footnote).
- `extension.Footnotes` / `extension.ContextFootnotes(pc)` — a parser-context-scoped interface for tracking footnote definitions/references while parsing.
- `extension/ast.TableBody` / `extension/ast.NewTableBody()` — wraps a table's body rows, sibling to `TableHeader` under `Table`.
- `extension.IsTask(node)` / `extension.TaskStatusOf(node)` — helpers for inspecting task-list items, now that `extension/ast.TaskCheckBox` is gone.

**`util`**
- `util.BytesFilter.AddString(st string)` / `.ContainsString(st string) bool`, for filters keyed by string instead of `[]byte`.
