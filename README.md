goldmark
==========================================

[![http://godoc.org/github.com/yuin/goldmark](https://godoc.org/github.com/yuin/goldmark?status.svg)](http://godoc.org/github.com/yuin/goldmark)
[![https://travis-ci.org/yuin/goldmark](https://travis-ci.org/yuin/goldmark.svg)](https://travis-ci.org/yuin/goldmark)

> A markdown parser written in Go. Easy to extend, standard compliant, well structured.

goldmark is compliant to CommonMark 0.29.

Motivation
----------------------
I need a markdown parser for Go that meets following conditions:

- Easy to extend.
    - Markdown is poor in document expressions compared with other light markup languages like restructuredText.
    - We have extended a markdown syntax. i.e. : PHPMarkdownExtra, Github Flavored Markdown.
- Standard compliant.
    - Markdown has many dialects.
    - Github Flavored Markdown is widely used and it is based on CommonMark aside from whether CommonMark is good specification or not.
        - CommonMark is too complicated and hard to implement.
- Well structured.
    - AST based, and preserves source potision of nodes.
- Written in pure Go.

[golang-commonmark](https://gitlab.com/golang-commonmark/markdown) may be a good choice, but it seems copy of the [markdown-it](https://github.com/markdown-it) .

[blackfriday.v2](https://github.com/russross/blackfriday/tree/v2) is a fast and widely used implementation, but it is not CommonMark compliant and can not be extended from outside of the package since it's AST is not interfaces but structs.

As mentioned above, CommonMark is too complicated and hard to implement, So Markdown parsers base on CommonMark barely exist.

Usage
----------------------

Convert Markdown documents with the CommonMark compliant mode:

```go
var buf bytes.Buffer
if err := goldmark.Convert(source, &buf); err != nil {
  panic(err)
}
```

Customize a parser and a renderer:

```go
md := goldmark.NewMarkdown(
          goldmark.WithExtensions(extension.GFM),
          goldmark.WithParserOptions(
              parser.WithHeadingID(),
          ),
          goldmark.WithRendererOptions(
              html.WithHardWraps(),
              html.WithXHTML(),
          ),
      )
var buf bytes.Buffer
if err := md.Convert(source, &buf); err != nil {
    panic(err)
}
```

Parser and Renderer options
------------------------------

### Parser options

| Functional option | Type | Description |
| ----------------- | ---- | ----------- |
| `parser.WithBlockParsers` | List of `util.PrioritizedSlice` whose elements are `parser.BlockParser` | Parsers for parsing block level elements. | 
| `parser.WithInlineParsers` | List of `util.PrioritizedSlice` whose elements are `parser.InlineParser` | Parsers for parsing inline level elements. | 
| `parser.WithParagraphTransformers` | List of `util.PrioritizedSlice` whose elements are `parser.ParagraphTransformer` | Transformers for transforming paragraph nodes. | 
| `parser.WithHeadingID` | `-` | Enables custom heading ids( `{#custom-id}` ) and auto heading ids. |
| `parser.WithFilterTags` | `...string` | HTML tag names forbidden in HTML blocks and Raw HTMLs. |

### HTML Renderer options

| Functional option | Type | Description |
| ----------------- | ---- | ----------- |
| `html.WithWriter` | `html.Writer` | `html.Writer` for writing contents to an `io.Writer`. |
| `html.WithHardWraps` | `-` | Render new lines as `<br>`.|
| `html.WithXHTML` | `-` | Render as XHTML. |
| `html.WithUnsafe` | `-` | By default, goldmark does not render raw HTMLs and potentially dangerous links. With this option, goldmark renders these contents as it is. |

### Built-in extensions

- `extension.Table`
- `extension.Strikethrough`
- `extension.Linkify`
- `extension.TaskList`
- `extension.GFM`
  - This extension enables Table, Strikethrough, Linkify and TaskList.
    In addition, this extension sets some tags to `parser.FilterTags` .

Create extensions
--------------------
**TODO**

See `extension` directory for examples of extensions.

Summary:

1. Define AST Node as a struct in which `ast.BaseBlock` or `ast.BaseInline` is embedded.
2. Write a parser that implements `parser.BlockParser` or `parser.InlineParser`.
3. Write a renderer that implements `renderer.NodeRenderer`.
4. Define your goldmark extension that implements `goldmark.Extender`.

Security
--------------------
By default, goldmark does not render raw HTMLs and potentially dangerous urls.
If you need to gain more control about untrusted contents, it is recommended to
use HTML sanitizer such as [bluemonday](https://github.com/microcosm-cc/bluemonday).

Benchmark
--------------------
You can run this benchmark in the `_benchmark` directory.

blackfriday v2 is fastest, but it is not CommonMark compiliant and not an extensible.

Though goldmark builds clean extensible AST structure and get full compliance with 
Commonmark, it is resonably fast and less memory consumption.

```
BenchmarkGoldMark-4                  200           7291402 ns/op         2259603 B/op   16867 allocs/op
BenchmarkGolangCommonMark-4          200           7709939 ns/op         3053760 B/op   18682 allocs/op
BenchmarkBlackFriday-4               300           5776369 ns/op         3356386 B/op   17480 allocs/op
```

Donation
--------------------
BTC: 1NEDSyUmo4SMTDP83JJQSWi1MvQUGGNMZB

License
--------------------
MIT

Author
--------------------
Yusuke Inuzuka
