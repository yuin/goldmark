## goldmark internal(for extension developers)

### Overview

goldmark's Markdown processing pipeline is outlined in the diagram below.

```
            <Markdown source ([]byte)>
                           |
                           V
            +-------- parser.Parser ---------------------------+
            | 1. Parse block elements into AST                 |
            |      For each paragraph, apply                   |
            |      ParagraphTransformers                       |
            | 2. Traverse block AST; for each block node,      |
            |    parse its Source() into inline nodes.         |
            |    At the end of each block, process             |
            |    the delimiter stack (emphasis, strong, etc.)  |
            | 3. Apply ASTTransformers to the whole AST        |
            +--------------------------------------------------+
                           |
                           V
                      <ast.Node tree>
                           |
                           V
            +-------- renderer.Renderer[W] --------------------+
            | 1. Walk AST; for each node, call the             |
            |    NodeRenderer[W] registered for its Kind       |
            +--------------------------------------------------+
                           |
                           V
                        <Output written to W>
```

An extension can hook into any of these stages by providing implementations of the interfaces described below. At a high level, building an extension requires four steps:

1. **Define AST nodes** — structs embedding `ast.BaseBlock` or `ast.BaseInline`.
2. **Write a parser** — implementing `parser.BlockParser`, `parser.InlineParser`, `parser.ParagraphTransformer`, or `parser.ASTTransformer`.
3. **Write a renderer** — a `renderer.NodeRenderer[W]` for your output format (e.g. `html.NodeRenderer` = `renderer.NodeRenderer[io.Writer]`).
4. **Package them as extensions** — implementing `parser.Extension` and/or `renderer.Extension[C]`.

### Step 1: Define AST nodes

Every custom node must embed either `ast.BaseBlock` (for block-level elements) or `ast.BaseInline` (for inline elements) and must:

- Implement `Kind() ast.NodeKind` returning a package-level `NodeKind` variable.
- Implement `Dump(source []byte) *ast.NodeDump` for debugging.
- Call `n.Init(n)` in its constructor.

```go no-run
package myext

import gast "github.com/yuin/goldmark/v2/ast"

// MyNode represents a custom inline element.
type MyNode struct {
    gast.BaseInline
    // Add fields for data that belongs to the node semantics.
    // Do NOT store parser-internal state here.
    MyField string
}

func (n *MyNode) Dump(_ []byte) *NodeDump {
    return gast.NewNodeDump(n, map[string]any {
        "MyField": n.MyField,
    })
}

var KindMyNode = gast.NewNodeKind("MyNode")

func (n *MyNode) Kind() gast.NodeKind { return KindMyNode }

func NewMyNode(field string) *MyNode {
    n := &MyNode{MyField: field}
    n.Init(n) // always required
    return n
}
```

For block nodes, embed `ast.BaseBlock`. The block's raw source text (used later for inline parsing) is stored via `AppendSource` / `Source()` rather than in a plain string field.

```go no-run
type MyBlock struct {
    gast.BaseBlock
}

func (n *MyBlock) Dump(_ []byte) *NodeDump {
    return gast.NewNodeDump(n, nil)
}

var KindMyBlock = gast.NewNodeKind("MyBlock")

func (n *MyBlock) Kind() gast.NodeKind { return KindMyBlock }

func NewMyBlock() *MyBlock {
    n := &MyBlock{}
    n.Init(n)
    return n
}
```

### Step 2: Write a parser

#### Block parser (`parser.BlockParser`)

A `BlockParser` opens and continues a block-level element line by line.

```go no-run
type BlockParser interface {
    // Trigger returns the set of first-column bytes that activate Open.
    // Return nil to be called for every line.
    Trigger() []byte

    // Open is called when the trigger byte is seen at the start of a line.
    // Return (node, HasChildren) if this line begins a new block, or (nil, NoChildren).
    Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State)

    // Continue is called for each subsequent line while the block is open.
    // Return (Continue | HasChildren), (Continue | NoChildren), or Close.
    Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State

    // Close is called when the block is finalised.
    Close(node ast.Node, reader text.Reader, pc parser.Context)

    // CanInterruptParagraph returns true if this parser may interrupt a paragraph.
    CanInterruptParagraph() bool

    // CanAcceptIndentedLine returns true if this parser may open with an indented line.
    CanAcceptIndentedLine() bool
}
```

Inside `Open` and `Continue`, use `text.Reader` to inspect and advance through the source:

| Method | Description |
|---|---|
| `reader.PeekLine()` | Returns `(line []byte, segment text.Segment)` without advancing |
| `reader.Advance(n)` | Advances the pointer by `n` bytes within the current line |
| `reader.AdvanceToEOL()` | Advances to the end of the current line |
| `reader.AdvanceLine()` | Moves to the start of the next line |
| `reader.LineOffset()` | Byte offset of the current position from the line start |
| `reader.Source()` | The full source byte slice |
| `pc.BlockOffset()` | Position of the first non-space byte on the current line (valid only in `Open`) |
| `pc.BlockIndent()` | Indentation width of the current line (valid only in `Open`) |

To store the source text that will later be parsed into inline nodes, call `node.AppendSource(segment)`:

```go no-run
func (b *myBlockParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
    line, segment := reader.PeekLine()
    if !bytes.HasPrefix(line, []byte(">>> ")) {
        return nil, parser.NoChildren
    }
    node := NewMyBlock()
    node.SetPos(segment.Start)
    reader.Advance(4) // consume ">>> "
    _, seg := reader.PeekLine()
    node.AppendSource(seg.TrimRightSpace(reader.Source()))
    reader.AdvanceToEOL()
    return node, parser.HasChildren
}
```

#### Inline parser (`parser.InlineParser`)

An `InlineParser` is triggered by a specific byte within a line and returns an inline AST node.

```go no-run
type InlineParser interface {
    // Trigger returns the bytes that activate this parser (must be punctuation or space).
    Trigger() []byte

    // Parse is called when the trigger byte is encountered.
    // It may consume beyond the current line.
    // Return nil if the trigger does not match.
    Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node
}
```

Optionally implement `parser.CloseBlocker` to receive a callback when the enclosing block is closed:

```go no-run
type CloseBlocker interface {
    CloseBlock(parent ast.Node, block text.Reader, pc parser.Context)
}
```

#### Delimiter-based inline elements

Elements like emphasis, strong, and strikethrough are based on a matching opener/closer delimiter pair. Use `parser.ParseDelimiter` together with a `parser.DelimiterProcessor`:

```go no-run
type DelimiterProcessor interface {
    IsDelimiter(byte) bool
    CanOpenCloser(opener, closer *parser.Delimiter) bool
    OnMatch(consumes int) ast.Node
}
```

`parser.ParseDelimiter(block, minimum, processor, pc)` scans the run of delimiter characters, pushes a `*Delimiter` node onto the delimiter stack in `pc`, and returns it. The matching between openers and closers is resolved later by `parser.ProcessDelimiters`. Refer to the strikethrough extension (`extension/strikethrough.go`) for a complete example.

#### Paragraph transformer (`parser.ParagraphTransformer`)

A `ParagraphTransformer` is called on every `*ast.Paragraph` after block parsing, before inline parsing. It can replace the paragraph with a different node (e.g. table, definition list). The table and definition list extensions use this hook.

```go no-run
type ParagraphTransformer interface {
    Transform(node *ast.Paragraph, reader text.Reader, pc parser.Context)
}
```

#### AST transformer (`parser.ASTTransformer`)

An `ASTTransformer` receives the fully-parsed `*ast.Document` and can make global changes.

```go no-run
type ASTTransformer interface {
    Transform(node *ast.Document, reader text.Reader, pc parser.Context)
}
```

#### Parser context (`parser.Context`)

`pc parser.Context` is a key/value store scoped to a single parse invocation. Use it to pass state between `Open`, `Continue`, and `Close` calls, or between a block parser and an AST transformer.

```go no-run
var myKey = parser.NewContextKey()

// store
pc.Set(myKey, myValue)

// retrieve
val := pc.Get(myKey)
```

### Step 3: Write a renderer

The renderer walks the AST and calls the `NodeRenderer[W]` registered for each node's `Kind`. The type parameter `W` is the writer type; for HTML output `W` is `io.Writer`.

```go no-run
// renderer.NodeRenderer[W] signature
type NodeRenderer[W any] interface {
    Render(w W, source []byte, n ast.Node, entering bool, rc renderer.Context) (ast.WalkStatus, error)
}
```

Use `renderer.NodeRendererFunc` to create a `NodeRenderer` from a plain function:

```go no-run
html.NodeRendererFunc(func(w io.Writer, source []byte, n ast.Node, entering bool, rc renderer.Context) (ast.WalkStatus, error) {
    bw := w.(util.BufWriter)
    if entering {
        _, _ = bw.WriteString("<my-element>")
    } else {
        _, _ = bw.WriteString("</my-element>")
    }
    return ast.WalkContinue, nil
})
```

For HTML output, cast `io.Writer` to `util.BufWriter` for efficient buffered writes:

```go no-run
w := writer.(util.BufWriter)
_, _ = w.WriteString("<tag>")
_ = w.WriteByte('\n')
```

To render HTML attributes attached to a node, use `html.RenderAttributes`:

```go no-run
if n.Attributes() != nil {
    _, _ = w.WriteString("<del")
    html.RenderAttributes(w, source, n, MyAttributeFilter, rc)
    _ = w.WriteByte('>')
} else {
    _, _ = w.WriteString("<del>")
}
```

`MyAttributeFilter` is a `util.BytesFilter` that controls which attribute names are allowed. Start from `html.GlobalAttributeFilter` and extend it as needed:

```go no-run
var MyAttributeFilter = html.GlobalAttributeFilter.ExtendString(`align,width`)
```

#### Writing text values safely: `text.Value` and context writers

When a `text.Value` is constructed, `text.Decoder` bound to it — a decoder (e.g. one created with `text.NewDecoder()`) resolves escapes/entities, `text.IdentityDecoder` leaves the bytes untouched. By the time a renderer sees `n.Value`, the decoding decision has already been made by whoever built the AST node.

What's left for the renderer is HTML-safety, and that's a choice between three context-scoped `util.BufWriter`s:

| Function | What it applies | Use for |
|---|---|---|
| `html.ContextTextWriter(rc)` | HTML-escapes `&`, `<`, `>`, `"` byte-by-byte | Content that must be safe inside HTML text/attributes — `Text.Value`, `CodeSpan.Value`, `CodeBlock.Value`, link/image `Title` |
| `html.ContextHTMLWriter(rc)` | Replaces NUL (`\x00`) with the replacement character (`\uFFFD`) only | Content that is already valid HTML — `RawHTML.Value`, `HTMLBlock.Value` |
| `html.ContextLinkURLWriter(rc)` | Escapes unsafe URL characters | URLs in link/image href |

Write a `text.Value` to one of these writers with `Value.WriteTo`:

```go no-run
// Render display text: HTML-escape it, decoding already happened at construction time.
tw := html.ContextTextWriter(rc)
_, _ = n.Value.WriteTo(tw, source)

// Render raw HTML that is trusted to already be valid: only NUL is replaced.
hw := html.ContextHTMLWriter(rc)
_, _ = n.Value.WriteTo(hw, source)
```

Writing a constant string (a fixed HTML tag or literal punctuation that contains no characters needing escaping) directly to the `util.BufWriter` is fine. Writing a **variable** value — anything derived from node fields or the source byte slice — must always go through one of the mechanisms above.

#### Node renderer decorator

`renderer.NodeRendererDecorator[W]` lets you run code before and after the node rendering:

```go no-run
type NodeRendererDecorator[W any] = func(next NodeRenderer[W]) NodeRenderer[W]
```

`NodeRendererDecorator` decorates a `NodeRenderer` like `net/http` middlewares.

Use `html.WithNodeRendererDecorator(s)` (or `renderer.WithNodeRendererDecorator(s)`) to decorate a node renderer. 

e.g. : You can decorate the `Document` node renderer to add required JavaScript:

```go no-run
func addMyScript(next html.NodeRenderer) html.NodeRenderer {
    return html.NodeRendererFunc(func(w io.Writer, source []byte, n ast.Node,
        entering bool, rc renderer.Context) (ast.WalkStatus, error) {
        if !entering {
            bw := w.(util.BufWriter)
            _, _ = bw.WriteString(`<script src="my-script.js"></script>`)
        }
        return next.Render(w, source, n, entering, rc)
    })
}
```

### Step 4: Package as extensions

In v2, parser and renderer extensions are separate types.

**Parser extension** implements `parser.Extension`:

```go no-run
type Extension interface {
    ParserOptions(c *parser.Config) []parser.Option
}
```

**Renderer extension** implements `renderer.Extension[C]` (e.g. `html.Extension` = `renderer.Extension[html.Config]`):

```go no-run
type Extension[C any] interface {
    RendererOptions(c *C) []renderer.Option[C]
}
```

Pass parsers and transformers with a priority using `util.Prioritized`. Lower numbers run first. Built-in parsers use priorities in the range 0–1000; use a value in the same range to interleave with them, or a larger value to run after them.

```go no-run
type myParserExtension struct{}

func NewMyParser() parser.Extension { return &myParserExtension{} }

func (e *myParserExtension) ParserOptions(_ *parser.Config) []parser.Option {
    return []parser.Option{
        parser.WithBlockParsers(
            util.Prioritized(newMyBlockParser(), 600),
        ),
        parser.WithInlineParsers(
            util.Prioritized(newMyInlineParser(), 600),
        ),
    }
}

type myHTMLRendererExtension struct{}

func NewMyHTMLRenderer() html.Extension { return &myHTMLRendererExtension{} }

func (e *myHTMLRendererExtension) RendererOptions(_ *html.Config) []html.Option {
    return []html.Option{
        html.WithNodeRenderers(map[ast.NodeKind]html.NodeRenderer{
            KindMyNode: html.NodeRendererFunc(renderMyNode),
        }),
    }
}
```

Use both extensions together when building the parser and renderer:

```go no-run
p := parser.New(parser.WithExtensions(NewMyParser()))
r := html.New(html.WithExtensions(NewMyHTMLRenderer()))
doc := p.Parse(source)
if err := r.Render(&buf, source, doc); err != nil {
    // ...
}
```

**Recommended naming convention**

- Use `myext.NewParser()` and `myext.NewHTMLRenderer()` for the extension constructors, and `KindMyExt` for the node kind variable.
- Use `var myext.Parser` and `var myext.HTMLRenderer` for default extension values that do not require options.

### Setting `Pos` on nodes

Every AST node stores a `Pos() int` value that records the byte offset of the node's start in the source. goldmark uses this for features such as source mapping and LSP support.

**Automatic setting**: goldmark sets `Pos` automatically in most cases.

- After `BlockParser.Open` returns, the parser sets `Pos` to the position of the first non-space character on the opening line (`blockPos.Start + BlockOffset()`).
- After `InlineParser.Parse` returns, if `Pos` is still `-1` (the initial value set by `Init`), the parser sets it to the position of the trigger character.

**Manual setting is only needed when the default is wrong.** The most common case is when your parser advances past a fixed prefix before creating the node, and you want `Pos` to point to a position *after* that prefix — for example, the content start rather than the syntax character start:

```go no-run
func (s *myInlineParser) Parse(_ ast.Node, block text.Reader, pc parser.Context) ast.Node {
    line, segment := block.PeekLine()
    if !bytes.HasPrefix(line, []byte("@")) {
        return nil
    }
    block.Advance(1) // skip '@'
    _, afterAt := block.Position()
    node := NewMyMention()
    node.SetPos(afterAt.Start) // point to the mention name, not the '@'
    // ...
    return node
}
```

If you do not call `SetPos`, the parser will fall back to the trigger-character position, which is correct for most simple inline elements.

**ParagraphTransformer and ASTTransformer**: When you replace or restructure nodes during transformation, the new node does not automatically inherit `Pos` or `HasBlankPreviousLines` from the original. You must copy both explicitly:

```go no-run
func (t *myTransformer) Transform(para *ast.Paragraph, reader text.Reader, pc parser.Context) {
    newNode := NewMyBlock()

    // Copy the position from the paragraph being replaced.
    newNode.SetPos(para.Pos())

    // Preserve blank-line information so that tight/loose list rendering
    // and other spacing logic continues to work correctly.
    newNode.SetBlankPreviousLines(para.HasBlankPreviousLines())

    parent := para.Parent()
    parent.ReplaceChild(para, newNode)
}
```

Forgetting either of these is a common source of subtle rendering bugs.

### Choosing between `text.SingleLineValue`, `text.MultiLineValue`, and `text.Lines`

The `text` package provides three types for holding source content in AST nodes. Choose based on the CommonMark specification for the field, not on implementation convenience.

| Type | When to use | Examples |
|---|---|---|
| `text.Value` | An interface for a single-line value or a multi-line value | - |
| `text.SingleLineValue` | The spec guarantees the value fits on a single line | Link destination (`[text](url)`), fenced code block info string |
| `text.MultiLineValue` | The spec allows the value to span multiple lines | Link title, code span content, raw HTML |
| `text.Lines` | A special block element that holds raw, unparsed block content line-by-line | `CodeBlock.Value`, `HTMLBlock.Value` |

`text.SingleLineValue` and `text.MultiLineValue` both reference source positions via `text.Index` (a `[Start, Stop)` byte range) or hold a literal string, so they never copy the source unnecessarily. `text.Lines` is a slice of `text.Segment`, where each segment corresponds to one source line with optional padding.

Use the generic constructors to create values:

```go no-run
import "github.com/yuin/goldmark/v2/text"

// SingleLineValue — always single-line. Every constructor takes an explicit text.Decoder
// (e.g. text.IdentityDecoder for raw content like inline HTMLs, or a decoder from text.NewDecoder() or reader.Decoder()).
dest := text.NewSingleLineValueFromIndex(text.NewIndex(start, stop), reader.Decoder()) // source position
dest := text.NewSingleLineValueFromString("https://example.com", reader.Decoder())     // literal string

// MultiLineValue — may span lines
title := text.NewMultiLineValueFromIndex(text.NewIndex(start, stop), text.IdentityDecoder)      // single span
title := text.NewMultiLineValueFromIndices([]text.Index{idx1, idx2}, reader.Decoder())      // multiple spans

// Lines — raw block content
var lines text.Lines
lines.AppendSegment(segment) // add one source line at a time
```

For more complex construction (e.g. building up a value from several segments while deciding the decoder once), use `text.ValueBuilder`: `var builder text.ValueBuilder; builder.AddSegment(seg).Decoder(d).BuildSingleLine()` (or `.BuildMultiLine()` and `.Build`).

### Complete examples

- See <https://github.com/yuin/goldmark/tree/v2/extension>  for complete examples of custom extensions.
