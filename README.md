goldmark
==========================================

[![http://godoc.org/github.com/yuin/goldmark](https://godoc.org/github.com/yuin/goldmark?status.svg)](http://godoc.org/github.com/yuin/goldmark)
[![https://travis-ci.org/yuin/goldmark](https://travis-ci.org/yuin/goldmark.svg)](https://travis-ci.org/yuin/goldmark)
[![https://coveralls.io/r/yuin/goldmark](https://coveralls.io/repos/yuin/goldmark/badge.svg)](https://coveralls.io/r/yuin/goldmark)
[![https://goreportcard.com/report/github.com/yuin/goldmark](https://goreportcard.com/badge/github.com/yuin/goldmark)](https://goreportcard.com/report/github.com/yuin/goldmark)

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
    - AST based, and preserves source position of nodes.
- Written in pure Go.

[golang-commonmark](https://gitlab.com/golang-commonmark/markdown) may be a good choice, but it seems copy of the [markdown-it](https://github.com/markdown-it) .

[blackfriday.v2](https://github.com/russross/blackfriday/tree/v2) is a fast and widely used implementation, but it is not CommonMark compliant and can not be extended from outside of the package since it's AST is not interfaces but structs. 

Furthermore, its behavior differs with other implementations in some cases especially of lists.  ([Deep nested lists don't output correctly #329](https://github.com/russross/blackfriday/issues/329), [List block cannot have a second line #244](https://github.com/russross/blackfriday/issues/244), etc).

This behavior sometimes causes problems. If you migrate your markdown text to blackfriday based wikis from Github, many lists will immediately be broken.

As mentioned above, CommonMark is too complicated and hard to implement, So Markdown parsers based on CommonMark barely exist.

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
- **Robust** : goldmark is tested with [go-fuzz](https://github.com/dvyukov/go-fuzz), a fuzz testing tool.
- **Builtin extensions.** : goldmark ships with common extensions like tables, strikethrough,
  task lists, and definition lists.
- **Depends only on standard libraries.**

Installation
----------------------
```bash
$ go get github.com/yuin/goldmark
```


Usage
----------------------
Import packages:

```
import (
	"bytes"
	"github.com/yuin/goldmark"
)
```


Convert Markdown documents with the CommonMark compliant mode:

```go
var buf bytes.Buffer
if err := goldmark.Convert(source, &buf); err != nil {
  panic(err)
}
```

With options
------------------------------

```go
var buf bytes.Buffer
if err := goldmark.Convert(source, &buf, parser.WithWorkers(16)); err != nil {
  panic(err)
}
```

| Functional option | Type | Description |
| ----------------- | ---- | ----------- |
| `parser.WithContext` | A parser.Context | Context for the parsing phase. |
|  parser.WithWorkers | int | Number of goroutines that execute concurrent inline element parsing. |

`parser.WithWorkers` may make performance better a little if markdown text
is relatively large. Otherwise, `parser.Workers` may cause performance degradation due to
goroutine overheads.

Custom parser and renderer
--------------------------
```go
import (
	"bytes"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

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

### HTML Renderer options

| Functional option | Type | Description |
| ----------------- | ---- | ----------- |
| `html.WithWriter` | `html.Writer` | `html.Writer` for writing contents to an `io.Writer`. |
| `html.WithHardWraps` | `-` | Render new lines as `<br>`.|
| `html.WithXHTML` | `-` | Render as XHTML. |
| `html.WithUnsafe` | `-` | By default, goldmark does not render raw HTMLs and potentially dangerous links. With this option, goldmark renders these contents as it is. |

### Built-in extensions

- `extension.Table`
  - [Github Flavored Markdown: Tables](https://github.github.com/gfm/#tables-extension-)
- `extension.Strikethrough`
  - [Github Flavored Markdown: Strikethrough](https://github.github.com/gfm/#strikethrough-extension-)
- `extension.Linkify`
  - [Github Flavored Markdown: Autolinks](https://github.github.com/gfm/#autolinks-extension-)
- `extension.TaskList`
  - [Github Flavored Markdown: Task list items](https://github.github.com/gfm/#task-list-items-extension-)
- `extension.GFM`
  - This extension enables Table, Strikethrough, Linkify and TaskList.
  - This extension does not filter tags defined in [6.11Disallowed Raw HTML (extension)](https://github.github.com/gfm/#disallowed-raw-html-extension-).
    If you need to filter HTML tags, see [Security](#security)
- `extension.DefinitionList`
  - [PHP Markdown Extra: Definition lists](https://michelf.ca/projects/php-markdown/extra/#def-list)
- `extension.Footnote`
  - [PHP Markdown Extra: Footnotes](https://michelf.ca/projects/php-markdown/extra/#footnotes)
- `extension.Typographer`
  - This extension substitutes punctuations with typographic entities like [smartypants](https://daringfireball.net/projects/smartypants/).

### Attributes
`parser.WithAttribute` option allows you to define attributes on some elements.

Currently only headings support attributes.

**Attributes are being discussed in the 
[CommonMark forum](https://talk.commonmark.org/t/consistent-attribute-syntax/272). 
This syntax possibly changes in the future.**


#### Headings

```
## heading ## {#id .className attrName=attrValue class="class1 class2"}

## heading {#id .className attrName=attrValue class="class1 class2"}
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

This benchmark parses a relatively large markdown text. In such text, concurrent parsing
makes performance better a little.

```
BenchmarkMarkdown/Blackfriday-v2-4                   300           5316935 ns/op         3321072 B/op      20050 allocs/op
BenchmarkMarkdown/GoldMark(workers=16)-4             300           5506219 ns/op         2702358 B/op      14494 allocs/op
BenchmarkMarkdown/GoldMark-4                         200           5903779 ns/op         2594304 B/op      13861 allocs/op
BenchmarkMarkdown/CommonMark-4                       200           7147659 ns/op         2752977 B/op      18827 allocs/op
BenchmarkMarkdown/Lute-4                             200           5930621 ns/op         2839712 B/op      21165 allocs/op
BenchmarkMarkdown/GoMarkdown-4                        10         120953070 ns/op         2192278 B/op      22174 allocs/op
```

### against cmark(A CommonMark reference implementation written in c)

```
----------- cmark -----------
file: _data.md
iteration: 50
average: 0.0047014618 sec
------- goldmark -------
file: _data.md
iteration: 50
average: 0.0052624750 sec
------- goldmark(workers=16) -------
file: _data.md
iteration: 50
average: 0.0044918780 sec
```

As you can see, goldmark performs pretty much equally to the cmark.

Extensions
--------------------

- [goldmark-meta](https://github.com/yuin/goldmark-meta) : A YAML metadata 
  extension for the goldmark markdown parser.
- [goldmark-highlighting](https://github.com/yuin/goldmark-highlighting) : A Syntax highlighting extension 
  for the goldmark markdown parser. 
- [goldmark-mathjax](https://github.com/litao91/goldmark-mathjax) : Mathjax support for goldmark markdown parser

Donation
--------------------
BTC: 1NEDSyUmo4SMTDP83JJQSWi1MvQUGGNMZB

License
--------------------
MIT

Author
--------------------
Yusuke Inuzuka
