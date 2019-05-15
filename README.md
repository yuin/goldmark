goldmark
==========================================

[![http://godoc.org/github.com/yuin/goldmark](https://godoc.org/github.com/yuin/goldmark?status.svg)](http://godoc.org/github.com/yuin/goldmark)
[![https://travis-ci.org/yuin/goldmark](https://travis-ci.org/yuin/goldmark.svg)](https://travis-ci.org/yuin/goldmark)
[![https://coveralls.io/r/yuin/goldmark](https://coveralls.io/repos/yuin/goldmark/badge.svg)](https://coveralls.io/r/yuin/goldmark)

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

Furthermore, its behavior differs with other implementations in some cases especially of lists.  ([Deep nested lists don't output correctly #329](https://github.com/russross/blackfriday/issues/329), [List block cannot have a second line #244](https://github.com/russross/blackfriday/issues/244), etc).

This behavior sometimes causes problems. If you migrate your markdown text to blackfriday based wikis from Github, many lists will immediately be broken.

As mentioned above, CommonMark is too complicated and hard to implement, So Markdown parsers base on CommonMark barely exist.

Features
----------------------

- **Standard compliant.** : goldmark get full compliance with latest CommonMark spec.
- **Extensible.** : Do you want to add a `@username` mention syntax to the markdown?
  You can easily do it in goldmark. You can add your AST nodes, 
  parsers for block level elements, parsers for inline level elements, 
  transformers for paragraphs, transformers for whole AST structure, and
  renderers.
- **Preformance.** : goldmark performs pretty much equally to the cmark
  (CommonMark reference implementation written in c).
- **Builtin extensions.** : goldmark ships with common extensions like tables, strikethrough,
  task lists, and definition lists.
- **Depends only on standard libraries.**

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
md := goldmark.New(
          goldmark.WithExtensions(extension.GFM),
          goldmark.WithParserOptions(
              parser.WithAutoHeadingID(),
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
| `parser.WithBlockParsers` | A `util.PrioritizedSlice` whose elements are `parser.BlockParser` | Parsers for parsing block level elements. | 
| `parser.WithInlineParsers` | A `util.PrioritizedSlice` whose elements are `parser.InlineParser` | Parsers for parsing inline level elements. | 
| `parser.WithParagraphTransformers` | A `util.PrioritizedSlice` whose elements are `parser.ParagraphTransformer` | Transformers for transforming paragraph nodes. | 
| `parser.WithAutoHeadingID` | `-` | Enables auto heading ids. |
| `parser.WithAttribute` | `-` | Enables custom attributes. Currently only headings supports attributes. |
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
  - [Gitmark Flavored Markdown: Tables](https://github.github.com/gfm/#tables-extension-)
- `extension.Strikethrough`
  - [Gitmark Flavored Markdown: Strikethrough](https://github.github.com/gfm/#strikethrough-extension-)
- `extension.Linkify`
  - [Gitmark Flavored Markdown: Autolinks](https://github.github.com/gfm/#autolinks-extension-)
- `extension.TaskList`
  - [Gitmark Flavored Markdown: Task list items](https://github.github.com/gfm/#task-list-items-extension-)
- `extension.GFM`
  - This extension enables Table, Strikethrough, Linkify and TaskList.
    In addition, this extension sets some tags to `parser.FilterTags` .
- `extension.DefinitionList`
  - [PHP Markdown Extra: Definition lists](https://michelf.ca/projects/php-markdown/extra/#def-list)
- `extension.Footnote`
  - [PHP Markdown Extra: Footnotes](https://michelf.ca/projects/php-markdown/extra/#footnotes)
- `extension.Typographer`
  - This extension substitutes punctuations with typographic entities like [smartypants](https://daringfireball.net/projects/smartypants/).

### Attributes
`parser.WithAttribute` option allows you to define attributes on some elements.

Currently only headings support attributes.

#### Headings

```
## heading ## {#id .className attrName=attrValue class="class1 class2"}
```

```
heading {#id .className attrName=attrValue}
============
```

### Typographer extension

Typographer extension translates plain ASCII punctuation characters into typographic punctuation HTML entities. 

Default substitutions are:

| Punctuation | Default entitiy |
| ------------ | ---------- |
| `'`           | `&lsquo;`, `&rsquo;` |
| `"`           | `&ldquo;`, `&rdquo;` |
| `--`       | `&ndash;` |
| `---`      | `&mdash;` |
| `...`      | `&hellip;` |
| `<<`       | `&laquo;` |
| `>>`       | `&raquo;` |

You can overwrite the substitutions by `extensions.WithTypographicSubstitutions`.

```go
markdown := goldmark.New(
	goldmark.WithExtensions(
		extension.NewTypographer(
			extension.WithTypographicSubstitutions(extension.TypographicSubstitutions{
				extension.LeftSingleQuote:  []byte("&sbquo;"),
				extension.RightSingleQuote: nil, // nil disables a substitution
			}),
		),
	),
)
```



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
If you need to gain more control over untrusted contents, it is recommended to
use HTML sanitizer such as [bluemonday](https://github.com/microcosm-cc/bluemonday).

Benchmark
--------------------
You can run this benchmark in the `_benchmark` directory.

### against other golang libraries

blackfriday v2 seems fastest, but it is not CommonMark compiliant so performance of the
blackfriday v2 can not simply be compared with other Commonmark compliant libraries.

Though goldmark builds clean extensible AST structure and get full compliance with 
Commonmark, it is resonably fast and less memory consumption.

```
BenchmarkGoldMark-4                  200           6388385 ns/op         2085552 B/op      13856 allocs/op
BenchmarkGolangCommonMark-4          200           7056577 ns/op         2974119 B/op      18828 allocs/op
BenchmarkBlackFriday-4               300           5635122 ns/op         3341668 B/op      20057 allocs/op
```

### against cmark(A CommonMark reference implementation written in c)

```
----------- cmark -----------
file: _data.md
iteration: 50
average: 0.0050112160 sec
go run ./goldmark_benchmark.go
------- goldmark -------
file: _data.md
iteration: 50
average: 0.0064833820 sec
```

As you can see, goldmark performs pretty much equally to the cmark.


Donation
--------------------
BTC: 1NEDSyUmo4SMTDP83JJQSWi1MvQUGGNMZB

License
--------------------
MIT

Author
--------------------
Yusuke Inuzuka
