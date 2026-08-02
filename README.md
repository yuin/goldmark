# goldmark

[![https://pkg.go.dev/github.com/yuin/goldmark](https://pkg.go.dev/badge/github.com/yuin/goldmark.svg)](https://pkg.go.dev/github.com/yuin/goldmark)
[![https://github.com/yuin/goldmark/actions?query=workflow:test](https://github.com/yuin/goldmark/actions/workflows/test.yaml/badge.svg?branch=master&event=push)](https://github.com/yuin/goldmark/actions?query=workflow:test)
[![https://coveralls.io/github/yuin/goldmark](https://coveralls.io/repos/github/yuin/goldmark/badge.svg?branch=master)](https://coveralls.io/github/yuin/goldmark)
[![https://goreportcard.com/report/github.com/yuin/goldmark](https://goreportcard.com/badge/github.com/yuin/goldmark)](https://goreportcard.com/report/github.com/yuin/goldmark)

> A Markdown parser written in Go. Easy to extend, standards-compliant, well-structured.

goldmark is compliant with CommonMark 0.31.2.

- [goldmark playground](https://yuin.github.io/goldmark/playground/) : Try goldmark online. This playground is built with WASM(5-10MB).

There is also a Rust version of goldmark: [rushdown](https://github.com/yuin/rushdown)

## v2 status

v2 is currently in beta. The API is almost stable, but some parts may change. Please report any issues you find.

## v2 Motivation
goldmark was originally created with a focus on my personal goals. 

- Extensible. You can easily add your own syntax to Markdown.
- Performance. Prioritize performance over semantically perfect AST.
- Focus on the purpose of converting to HTML.

Unexpectedly, goldmark has been used by many people.

goldmark has become a major Markdown parser in Go ecosystem.

In such a situation, there have been many requests regarding use cases that were not emphasized at the time of initial creation.

- Semantic analysis of Markdown documents using AST
- Use cases that use more detailed position information, such as LSP servers

In particular, as Markdown documents have come to be used as Lingua franca for AI, there is an increasing need to analyze Markdown documents semantically.For the same reason, there are also increasing use cases for generating Markdown documents rather than parsing them. The use of CLI in AI agents is increasing, also a growing need to convert to formats other than HTML.

Breaking changes to an extensible library like goldmark have a huge impact, as third-party extensions will no longer work. Therefore, I have avoided making breaking changes for a long time.

It has been more than 7 years since goldmark was created, and technical debt has been accumulating. In the meantime, the Go language specification has changed significantly, including the introduction of generics. I believe that the changes in use cases, represented by AI, are a good opportunity to fundamentally review the design of goldmark, and I have decided to make breaking changes.

## v2 and v1 differences

- v2 focuses on building a more semantic AST. For that reason, it is slightly less performant than v1.
- v2 uses generics.
- v2 clearly separates the parser and renderer. This makes it easier to implement rendering to formats other than HTML.
- v2 allows you to programmatically build an AST. And you can render the constructed AST to another format.
- v2 has all nodes hold the start position. In the future, third-party extensions that support v2 are also expected to hold the start position.
- The core parsing algorithm is the same as v1. Third-party extensions must support v2, but the most complex parsing part can be used almost as it is.

## Maintenance policy
This project will maintain bug fixes, including security fixes, up to one major version prior to the latest major version.

## Features

- **Standards-compliant** : goldmark is fully compliant with the latest [CommonMark](https://commonmark.org/) specification.
- **Extensible** : Do you want to add a `@username` mention syntax to Markdown?
  You can easily do so in goldmark. You can add your AST nodes,
  parsers for block-level elements, parsers for inline-level elements,
  transformers for paragraphs, transformers for the whole AST structure, and
  renderers.
- **Performance** : goldmark is one of the fastest CommonMark-compliant Markdown parsers in Go.
- **Robust** : goldmark is tested with `go test --fuzz`.
- **Built-in extensions** : goldmark ships with common extensions like tables, strikethrough,
  task lists, and definition lists.
- **Semantically clean AST** :  goldmark builds a clean AST structure that is easy to analyze and transform.
- **Depends only on standard libraries.**

## Installation

```bash
$ go get github.com/yuin/goldmark/v2
```


## Usage

Convert Markdown documents with the CommonMark-compliant mode:

```go
import (
    "bytes"
    "github.com/yuin/goldmark/v2/parser"
    "github.com/yuin/goldmark/v2/renderer/html"
)

source := []byte("こんにちは、 **世界** 。")

var buf bytes.Buffer
p := parser.New()
r := html.New()

doc := p.Parse(source)
if err := r.Render(&buf, source, doc); err != nil {
    panic(err)
}
if "<p>こんにちは、 <strong>世界</strong> 。</p>\n" != buf.String() {
    panic("unexpected output:" + buf.String())
}
```

Build an AST and render it to HTML:

```go
import (
    "bytes"
    "github.com/yuin/goldmark/v2/ast"
    "github.com/yuin/goldmark/v2/text"
    "github.com/yuin/goldmark/v2/renderer/html"
)

doc := ast.N(ast.NewDocument(),
    ast.N(ast.NewParagraph(),
        "こんにちは、",
        ast.N(ast.NewEmphasis(),
            "世界",
        ),
        "。",
    ),
    ast.N(func() ast.Node {
        n := ast.NewParagraph()
        n.SetAttribute("class", text.NewMultilineValue("greeting"))
        return n
    }(), "Hello, world."),
)

var buf bytes.Buffer
r := html.New()
if err := r.Render(&buf, nil, doc); err != nil {
    panic(err)
}
if "<p>こんにちは、<em>世界</em>。</p>\n<p class=\"greeting\">Hello, world.</p>\n" != buf.String() {
    panic("unexpected output:" + buf.String())
}
```


## Custom parser and renderer

```go
import (
    "bytes"
    "github.com/yuin/goldmark/v2/extension"
    "github.com/yuin/goldmark/v2/parser"
    "github.com/yuin/goldmark/v2/renderer/html"
)

source := []byte("こんにちは、 ~~世界~~ 。")

p := parser.New(parser.WithAttribute(), parser.WithExtensions(extension.NewStrikethroughParser()))
r := html.New(html.WithXHTML(), html.WithUnsafe(), html.WithExtensions(extension.NewStrikethroughHTMLRenderer()))

var buf bytes.Buffer
doc := p.Parse(source)
if err := r.Render(&buf, source, doc); err != nil {
    panic(err)
}
if "<p>こんにちは、 <del>世界</del> 。</p>\n" != buf.String() {
    panic("unexpected output:" + buf.String())
}

```

### Parser options

| Functional option | Type | Description |
| ----------------- | ---- | ----------- |
| `parser.WithBlockParsers` | `[]util.PrioritizedValue[parser.BlockParser]` | Parsers for parsing block level elements. |
| `parser.WithInlineParsers` | `[]util.PrioritizedValue[parser.InlineParser]` | Parsers for parsing inline level elements. |
| `parser.WithParagraphTransformers` | `[]util.PrioritizedValue[parser.ParagraphTransformer]` | Transformers for transforming paragraph nodes. |
| `parser.WithASTTransformers` | `[]util.PrioritizedValue[parser.ASTTransformer]` | Transformers for transforming an AST. |
| `parser.WithAutoHeadingID` | `-` | Enables auto heading ids. |
| `parser.WithAttribute` | `-` | Enables custom attributes. Currently only headings supports attributes. |
| `parser.WithIDGenerator` | `parser.IDGenerator` |  Generator for heading ids. |
| `parser.WithDefaultParsers` | `bool` | Enables default parsers. Default is true. |
| `parser.WithEscapedSpace` | `-` | Enables escaped space. This is useful for CJK users. |
| `parser.WithExtensions` | `[]parser.Extension` | Enables parser extensions. |

### HTML Renderer options

| Functional option | Type | Description |
| ----------------- | ---- | ----------- |
| `html.WithEscapedSpace` | `-` | Enables escaped space. This is useful for CJK users. |
| `html.WithEastAsianLineBreaks` | `html.EastAsianLineBreaks` | Soft line breaks are rendered as a newline. Some asian users will see it as an unnecessary space. With this option, soft line breaks between east asian wide characters will be ignored. |
| `html.WithHardWraps` | `-` | Render newlines as `<br>`.|
| `html.WithIsInTightBlock` | `html.IsInTightBlockFunc` | Function that determines whether a node is in a tight block. |
| `html.WithNodeRenderer` | `ast.NodeKind`, `html.NodeRenderer` | Add a node renderer for a specific node kind. |
| `html.WithNodeRenderers` | `map[ast.NodeKind]html.NodeRenderer` | Add node renderers for specific node kinds. |
| `html.WithNodeRendererDecorator` | `ast.NodeKind`, `html.NodeRendererDecorator` | Add a decorator for a node renderer. |
| `html.WithNodeRendererDecorators` | `map[ast.NodeKind]html.NodeRendererDecorator` | Add decorators for node renderers. |
| `html.WithXHTML` | `-` | Render as XHTML. |
| `html.WithUnsafe` | `-` | By default, goldmark does not render raw HTML or potentially dangerous links. With this option, goldmark renders such content as written. |
| `html.WithExtensions` | `[]html.Extension` | Enables parser extensions. |

#### East asian line breaks

| Style | Description |
| ----- | ----------- |
| `EastAsianLineBreaksSimple` | Soft line breaks are ignored if both sides of the break are east asian wide character. This behavior is the same as [`east_asian_line_breaks`](https://pandoc.org/MANUAL.html#extension-east_asian_line_breaks) in Pandoc. |
| `EastAsianLineBreaksCSS3Draft` | This option implements CSS text level3 [Segment Break Transformation Rules](https://drafts.csswg.org/css-text-3/#line-break-transform) with [some enhancements](https://github.com/w3c/csswg-drafts/issues/5086). |

**Example of `EastAsianLineBreaksSimple`**

Input Markdown:

```md
私はプログラマーです。
東京の会社に勤めています。
GoでWebアプリケーションを開発しています。
```

Output:

```html
<p>私はプログラマーです。東京の会社に勤めています。\nGoでWebアプリケーションを開発しています。</p>
```

**Example of `EastAsianLineBreaksCSS3Draft`**

Input Markdown:

```md
私はプログラマーです。
東京の会社に勤めています。
GoでWebアプリケーションを開発しています。
```

Output:

```html
<p>私はプログラマーです。東京の会社に勤めています。GoでWebアプリケーションを開発しています。</p>
```

### Built-in extensions

- `extension.Table`
    - [GitHub Flavored Markdown: Tables](https://github.github.com/gfm/#tables-extension-)
- `extension.Strikethrough`
    - [GitHub Flavored Markdown: Strikethrough](https://github.github.com/gfm/#strikethrough-extension-)
- `extension.Linkify`
    - [GitHub Flavored Markdown: Autolinks](https://github.github.com/gfm/#autolinks-extension-)
- `extension.TaskList`
    - [GitHub Flavored Markdown: Task list items](https://github.github.com/gfm/#task-list-items-extension-)
- `extension.GFM`
    - This extension enables Table, Strikethrough, Linkify and TaskList.
    - This extension does not filter tags defined in [6.11: Disallowed Raw HTML (extension)](https://github.github.com/gfm/#disallowed-raw-html-extension-).
    If you need to filter HTML tags, see [Security](#security).
    - If you need to parse github emojis, you can use [goldmark-emoji](https://github.com/yuin/goldmark-emoji) extension.
- `extension.DefinitionList`
    - [PHP Markdown Extra: Definition lists](https://michelf.ca/projects/php-markdown/extra/#def-list)
- `extension.Footnote`
    - [PHP Markdown Extra: Footnotes](https://michelf.ca/projects/php-markdown/extra/#footnotes)
- `extension.Typographer`
    - This extension substitutes punctuations with typographic entities like [smartypants](https://daringfireball.net/projects/smartypants/).

### Attributes
The `parser.WithAttribute` option allows you to define attributes on some elements.

Currently only headings support attributes.

**Attributes are being discussed in the
[CommonMark forum](https://talk.commonmark.org/t/consistent-attribute-syntax/272).
This syntax may possibly change in the future.**


#### Headings

```
## heading ## {#id .className attrName=attrValue class="class1 class2"}

## heading {#id .className attrName=attrValue class="class1 class2"}
```

```
heading {#id .className attrName=attrValue}
============
```

Attributes specification is almost the same as HTML attributes.

- `"` or `'` quoted strings can contain any character except the quote character itself. HTML entity references are also allowed.
- Unquoted values cannot contain any whitespace characters or `}`.

In addition to the HTML attribute specification, there is a special syntax for IDs and class names.

- `#`-prefixed strings are interpreted as ID attributes.
- `.`-prefixed strings are interpreted as class names.


### Table extension
The Table extension implements [Table(extension)](https://github.github.com/gfm/#tables-extension-), as
defined in [GitHub Flavored Markdown Spec](https://github.github.com/gfm/).

Specs are defined for XHTML, so specs use some deprecated attributes for HTML5.

You can override alignment rendering method via options.

| Functional option | Type | Description |
| ----------------- | ---- | ----------- |
| `extension.WithTableCellAlignMethod` | `extension.TableCellAlignMethod` | Option indicates how are table cells aligned. |

### Typographer extension

The Typographer extension translates plain ASCII punctuation characters into typographic-punctuation HTML entities.

Default substitutions are:

| Punctuation | Default entity |
| ------------ | ---------- |
| `'`           | `&lsquo;`, `&rsquo;` |
| `"`           | `&ldquo;`, `&rdquo;` |
| `--`       | `&ndash;` |
| `---`      | `&mdash;` |
| `...`      | `&hellip;` |
| `<<`       | `&laquo;` |
| `>>`       | `&raquo;` |

You can override the default substitutions via `extensions.WithTypographicSubstitutions`.

```go
import (
    "github.com/yuin/goldmark/v2/extension"
    "github.com/yuin/goldmark/v2/parser"
)
_ = parser.New(
        parser.WithExtensions(extension.NewTypographerParser(
            extension.WithTypographicSubstitutions(extension.TypographicSubstitutions{
                extension.LeftSingleQuote:  []byte("&sbquo;"),
                extension.RightSingleQuote: nil, // nil disables a substitution
            }),
        )),
)
```

### Linkify extension

The Linkify extension implements [Autolinks(extension)](https://github.github.com/gfm/#autolinks-extension-), as
defined in [GitHub Flavored Markdown Spec](https://github.github.com/gfm/).

Since the spec does not define details about URLs, there are numerous ambiguous cases.

You can override autolinking patterns via options.

| Functional option | Type | Description |
| ----------------- | ---- | ----------- |
| `extension.WithAllowedProtocols` | `[][]byte \| []string` | List of allowed protocols such as `[]string{ "http:" }` |
| `extension.WithURLRegexp` | `*regexp.Regexp` | Regexp that defines URLs, including protocols |
| `extension.WithWWWRegexp` | `*regexp.Regexp` | Regexp that defines URL starting with `www.`. This pattern corresponds to [the extended www autolink](https://github.github.com/gfm/#extended-www-autolink) |
| `extension.WithEmailRegexp` | `*regexp.Regexp` | Regexp that defines email addresses` |

Example, using [xurls](https://github.com/mvdan/xurls):

```go
import (
    "mvdan.cc/xurls/v2"
    "github.com/yuin/goldmark/v2/extension"
    "github.com/yuin/goldmark/v2/parser"
)

_ = parser.New(
      parser.WithExtensions(
          extension.NewLinkifyParser(
              extension.WithAllowedProtocols([]string{
                  "http:",
                  "https:",
              }),
              extension.WithURLRegexp(
                  xurls.Strict(),
              ),
          ),
      ),
)
```

### Footnotes extension

The Footnote extension implements [PHP Markdown Extra: Footnotes](https://michelf.ca/projects/php-markdown/extra/#footnotes).

This extension has some options:

| Functional option | Type | Description |
| ----------------- | ---- | ----------- |
| `extension.WithIDPrefix` | `[]byte \| string` |  a prefix for the id attributes.|
| `extension.WithIDPrefixFunction` | `func(gast.Node) []byte` |  a function that determines the id attribute for given Node.|
| `extension.WithLinkTitle` | `[]byte \| string` |  an optional title attribute for footnote links.|
| `extension.WithBacklinkTitle` | `[]byte \| string` |  an optional title attribute for footnote backlinks. |
| `extension.WithLinkClass` | `[]byte \| string` |  a class for footnote links. This defaults to `footnote-ref`. |
| `extension.WithBacklinkClass` | `[]byte \| string` |  a class for footnote backlinks. This defaults to `footnote-backref`. |
| `extension.WithBacklinkHTML` | `[]byte \| string` |  a class for footnote backlinks. This defaults to `&#x21a9;&#xfe0e;`. |

Some options can have special substitutions. Occurrences of “^^” in the string will be replaced by the corresponding footnote number in the HTML output. Occurrences of “%%” will be replaced by a number for the reference (footnotes can have multiple references).

`extension.WithIDPrefix` and `extension.WithIDPrefixFunction` are useful if you have multiple Markdown documents displayed inside one HTML document to avoid footnote ids to clash each other.

`extension.WithIDPrefix` sets fixed id prefix, so you may write codes like the following:

```go no-run
import (
    "github.com/yuin/goldmark/v2/extension"
    "github.com/yuin/goldmark/v2/parser"
)

for _, path := range files {
    source := readAll(path)
    prefix := getPrefix(path)

    p := parser.New(
        parser.WithExtensions(
            extension.NewFootnote(
                extension.WithIDPrefix(path),
            ),
        ),
    )
    // convert source to HTML
}
```

`extension.WithIDPrefixFunction` determines an id prefix by calling given function, so you may write codes like the following:

```go no-run
import (
    "github.com/yuin/goldmark/v2/extension"
    "github.com/yuin/goldmark/v2/parser"
)

p := parser.New(
        extension.NewFootnote(
                extension.WithIDPrefixFunction(func(n gast.Node) []byte {
                    v, ok := n.OwnerDocument().Meta()["footnote-prefix"]
                    if ok {
                        return util.StringToReadOnlyBytes(v.(string))
                    }
                    return nil
                }),
        ),
    )


for _, path := range files {
    source := readAll(path)
    doc := p.Parse(source)
    doc.Meta()["footnote-prefix"] = getPrefix(path)
    // convert doc to HTML
}
```

You can use [goldmark-meta](https://github.com/yuin/goldmark-meta) to define a id prefix in the markdown document:


```markdown
---
title: document title
slug: article1
footnote-prefix: article1
---

# My article

```


## Security

By default, goldmark does not render raw HTML or potentially-dangerous URLs.
If you need to gain more control over untrusted contents, it is recommended that you
use an HTML sanitizer such as [bluemonday](https://github.com/microcosm-cc/bluemonday).

## Benchmark

You can run this benchmark in the `_benchmark` directory.

### against other golang libraries

```
goos: linux
goarch: amd64
pkg: banchmark
cpu: AMD Ryzen 7 8840HS w/ Radeon 780M Graphics
BenchmarkMarkdown/goldmark/v2-16                     176           6786133 ns/op         2587272 B/op      14460 allocs/op
BenchmarkMarkdown/goldmark/v1-16                     170           6815477 ns/op         2538288 B/op      14471 allocs/op
BenchmarkMarkdown/CommonMark-16                      146           8520952 ns/op         2699081 B/op      20129 allocs/op
BenchmarkMarkdown/Lute-16                             12          87993431 ns/op        12112462 B/op      33447 allocs/op
BenchmarkMarkdown/GoMarkdown-16                       10         115382667 ns/op         2272684 B/op      23832 allocs/op
```

## Extensions
### List of extensions

Note that not all extensions support v2.

- [goldmark-meta](https://github.com/yuin/goldmark-meta): A YAML metadata
  extension for the goldmark Markdown parser.
- [goldmark-highlighting](https://github.com/yuin/goldmark-highlighting): A syntax-highlighting extension
  for the goldmark markdown parser.
- [goldmark-emoji](https://github.com/yuin/goldmark-emoji): An emoji
  extension for the goldmark Markdown parser.
- [goldmark-mathjax](https://github.com/litao91/goldmark-mathjax): Mathjax support for the goldmark markdown parser
- [goldmark-pdf](https://github.com/stephenafamo/goldmark-pdf): A PDF renderer that can be passed to `goldmark.WithRenderer()`.
- [goldmark-hashtag](https://github.com/abhinav/goldmark-hashtag): Adds support for `#hashtag`-based tagging to goldmark.
- [goldmark-wikilink](https://github.com/abhinav/goldmark-wikilink): Adds support for `[[wiki]]`-style links to goldmark.
- [goldmark-anchor](https://github.com/abhinav/goldmark-anchor): Adds anchors (permalinks) next to all headers in a document.
- [goldmark-figure](https://github.com/mangoumbrella/goldmark-figure): Adds support for rendering paragraphs starting with an image to `<figure>` elements.
- [goldmark-frontmatter](https://github.com/abhinav/goldmark-frontmatter): Adds support for YAML, TOML, and custom front matter to documents.
- [goldmark-toc](https://github.com/abhinav/goldmark-toc): Adds support for generating tables-of-contents for goldmark documents.
- [goldmark-mermaid](https://github.com/abhinav/goldmark-mermaid): Adds support for rendering [Mermaid](https://mermaid-js.github.io/mermaid/) diagrams in goldmark documents.
- [goldmark-pikchr](https://github.com/jchenry/goldmark-pikchr): Adds support for rendering [Pikchr](https://pikchr.org/home/doc/trunk/homepage.md) diagrams in goldmark documents.
- [goldmark-embed](https://github.com/13rac1/goldmark-embed): Adds support for rendering embeds from YouTube links.
- [goldmark-latex](https://github.com/soypat/goldmark-latex): A $\LaTeX$ renderer that can be passed to `goldmark.WithRenderer()`.
- [goldmark-fences](https://github.com/stefanfritsch/goldmark-fences): Support for pandoc-style [fenced divs](https://pandoc.org/MANUAL.html#divs-and-spans) in goldmark.
- [goldmark-d2](https://github.com/FurqanSoftware/goldmark-d2): Adds support for [D2](https://d2lang.com/) diagrams.
- [goldmark-katex](https://github.com/FurqanSoftware/goldmark-katex): Adds support for [KaTeX](https://katex.org/) math and equations.
- [goldmark-img64](https://github.com/tenkoh/goldmark-img64): Adds support for embedding images into the document as DataURL (base64 encoded).
- [goldmark-enclave](https://github.com/quailyquaily/goldmark-enclave): Adds support for embedding youtube/bilibili video, X's [oembed X](https://publish.x.com/), [tradingview chart](https://www.tradingview.com/widget/)'s chart, [quaily widget](https://quaily.com), [spotify embeds](https://developer.spotify.com/documentation/embeds), [dify embed](https://dify.ai/) and html audio into the document.
- [goldmark-wiki-table](https://github.com/movsb/goldmark-wiki-table): Adds support for embedding Wiki Tables.
- [goldmark-tgmd](https://github.com/Mad-Pixels/goldmark-tgmd): A Telegram markdown renderer that can be passed to `goldmark.WithRenderer()`.
- [goldmark-treeblood](https://github.com/Wyatt915/goldmark-treeblood): Renders $\LaTeX$ expressions as MathML (pure Go, no external dependencies).
- [goldmark-subtext](https://github.com/zeozeozeo/goldmark-subtext): Support for Discord-style markdown subtexts
- [goldmark-customtag](https://github.com/tendstofortytwo/goldmark-customtag): Allows you to define custom block tags.
- [goldmark-cjk-friendly](https://github.com/tats-u/goldmark-cjk-friendly): Port of npm package [`remark-cjk-friendly` / `markdown-it-cjk-friendly`](https://github.com/tats-u/markdown-cjk-friendly) to goldmark. Similar to the [CJK extension](#cjk-extension) (`WithEscapedSpace`), but you do not need to explicitly add `\ ` around `*` and `**`. You can combine this with the [CJK extension](#cjk-extension).
- [goldmark-chart](https://github.com/TheGreatRambler/goldmark-chart): Generate static ChartJS charts using the simple [Markvis](https://markvis.js.org/#/) format.

<!--

### Loading extensions at runtime
[goldmark-dynamic](https://github.com/yuin/goldmark-dynamic) allows you to write a goldmark extension in Lua and load it at runtime without re-compilation.

Please refer to  [goldmark-dynamic](https://github.com/yuin/goldmark-dynamic) for details.

-->


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

func (n *MyNode) Dump(_ []byte) *NodeDump {
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
    html.RenderAttributes(w, source, n, MyAttributeFilter)
    _ = w.WriteByte('>')
} else {
    _, _ = w.WriteString("<del>")
}
```

`MyAttributeFilter` is a `util.BytesFilter` that controls which attribute names are allowed. Start from `html.GlobalAttributeFilter` and extend it as needed:

```go no-run
var MyAttributeFilter = html.GlobalAttributeFilter.ExtendString(`align,width`)
```

#### Writing text values safely: `html.Writer`

When writing content that comes from the parsed source (node fields, attribute values, etc.) you **must** go through the `html.Writer` obtained from the render context, not write directly to the `BufWriter`. The `html.Writer` applies all escaping required by CommonMark and HTML:

| Method | What it applies |
|---|---|
| `WriteText(w, src)` | HTML-escapes `&`, `<`, `>`, `"`, replaces NUL (`\x00`) with the replacement character (`\uFFFD`), and resolves backslash-escaped characters (`\*`, `\.`, etc.) and entity references (`&amp;`, `&#x41;`, etc.) |
| `RawWriteText(w, src)` | Same as `WriteText` but does **not** resolve backslash escapes or entity references — use for content that is already in its final logical form (e.g. a link destination stored as a `text.Value`) |
| `WriteHTML(w, src)` | Replaces NUL only — use when writing a value that is already valid, escaped HTML |


```go no-run
// Render display text that may contain backslash escapes or entity references.
writer.WriteText(w, n.Value.Bytes(source))

// Render a URL that should not have backslash interpretation applied.
writer.RawWriteText(w, n.Destination.Bytes(source))
```

Writing a constant string (a fixed HTML tag or literal punctuation that contains no characters needing escaping) directly to the `BufWriter` is fine. Writing a **variable** value — anything derived from node fields or the source byte slice — must always go through `html.Writer`.

```go no-run
bw := w.(util.BufWriter)
_, _ = bw.WriteString("<em>")        // OK: constant tag, no escaping needed
writer.WriteText(bw, value)          // required: variable content from source
_, _ = bw.WriteString("</em>")       // OK: constant tag
```

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

type myHTMLRendererExtension struct{
    writer html.Writer
}

func NewMyHTMLRenderer() html.Extension { return &myHTMLRendererExtension{} }

func (e *myHTMLRendererExtension) RendererOptions(cfg *html.Config) []html.Option {
    e.writer = cfg.Writer()
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
  - If your extension do not have options, you can use `myext.Parser` and `myext.HTMLRenderer` as the extension values.
- Use `myext.ParserOption` and `myext.HTMLRendererOption` for functional options.

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

### Choosing between `text.Value`, `text.MultilineValue`, and `text.Lines`

The `text` package provides three types for holding source content in AST nodes. Choose based on the CommonMark specification for the field, not on implementation convenience.

| Type | When to use | Examples |
|---|---|---|
| `text.Value` | The spec guarantees the value fits on a single line | Link destination (`[text](url)`), fenced code block info string |
| `text.MultilineValue` | The spec allows the value to span multiple lines | Link title, code span content, raw HTML |
| `text.Lines` | A special block element that holds raw, unparsed block content line-by-line | `CodeBlock.Value`, `HTMLBlock.Value` |

`text.Value` and `text.MultilineValue` both reference source positions via `text.Index` (a `[Start, Stop)` byte range) or hold a literal string, so they never copy the source unnecessarily. `text.Lines` is a slice of `text.Segment`, where each segment corresponds to one source line with optional padding.

Use the generic constructors to create values:

```go no-run
import "github.com/yuin/goldmark/v2/text"

// Value — always single-line
dest := text.NewIndexValue(text.NewIndex(start, stop))       // source position
dest := text.NewStringValue("https://example.com")           // literal string

// MultilineValue — may span lines
title := text.NewIndexMultilineValue(text.NewIndex(start, stop))       // single span
title := text.NewIndicesMultilineValue([]text.Index{seg1, seg2})       // multiple spans

// Lines — raw block content
var lines text.Lines
lines.AppendSegment(segment) // add one source line at a time
```

### Complete examples

- See extension directory for complete examples of custom extensions.

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

### `renderer/html` package

`html.NewRenderer(opts ...Option) renderer.NodeRenderer` has been replaced by `html.New(opts ...Option) Renderer`.

A `renderer.NodeRendererDecorator[W any]` is new in v2, allowing you to run code before and after the render pass.

`html.RenderAttributes` now takes a source as second argument.

### `ast.Node` interface

- `Type() NodeType` and the `NodeType` type (with constants `TypeBlock`, `TypeInline`, `TypeDocument`) have been removed. Use type assertions to `ast.BlockNode` or `ast.InlineNode` instead.
- `Text(source []byte) []byte` (was already deprecated in v1) has been removed.
- `HasBlankPreviousLines()`, `SetBlankPreviousLines()`, `Lines()`, `SetLines()`, and `IsRaw()` have been removed from `Node` and moved to the `BlockNode` interface.
- Tree mutation methods (`AppendChild`, `RemoveChild`, `RemoveChildren`, `InsertBefore`, `InsertAfter`, `ReplaceChild`) no longer take a `self Node` as their first argument.
- `Dump` now returns `*NodeDump`. `Dump(source []byte) *NodeDump` is the new signature.
- `ast.Attribute` uses `string` names instead of `[]byte`.
  - `SetAttributeString` and `AttributeString` has been removed.
  - `SetAttribute` and `Attribute` now take `string` names instead of `[]byte`.
- Attributes other than string type are no longer supported.
  - `goldmark_v1_attribute` build tag allows using v1-compatible attributes.
    - You can get parsed value as `any` using `Attribute#Any(source)`
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

- `Segment text.Segment` → `Value text.Value`
- `NewTextSegment(v Segment)` → `NewSegmentText(v Segment)`
- `NewRawTextSegment(v Segment)` removed; use `NewSegmentText(v)` then `n.SetRaw(true)`
- New: `NewStringText(s string) *Text`

**`ast.String` (inline node) — removed**

Use `ast.NewStringText(s string) *Text` instead.

**`ast.Emphasis`**

- `Level int` field removed. `*Emphasis` always represents single emphasis (`*`/`_`), `*Strong` always represents strong emphasis (`**`/`__`). They are now separate types.
- `NewEmphasis(level int)` → `NewEmphasis()`

**`ast.CodeSpan`**

- No longer holds `Text` child nodes. Now has `Value text.MultilineValue`.
- `NewCodeSpan()` → `NewCodeSpan(value text.MultilineValue)`

**`ast.RawHTML`**

- Now has `Value text.MultilineValue` (no children).
- `NewRawHTML()` → `NewRawHTML(value text.MultilineValue)`

**`ast.Heading`**

- Added `HeadingKind HeadingKind` (`HeadingKindATX` or `HeadingKindSetext`).
- `NewHeading(level int)` → `NewHeading(level int, kind HeadingKind)`

**`ast.CodeBlock` (fenced)**

- `FencedCodeBlock` is merged; `NewFencedCodeBlock(info *Text)` is gone.
- `Info` is now `text.Value` (not `*Text`); `Value text.Lines` holds the code body.
- `NewCodeBlock(kind CodeBlockKind)` → `NewCodeBlock(kind CodeBlockKind, value text.Lines, opts ...CodeBlockOption)`
- Optional info string: `ast.WithCodeBlockInfo(info)`

**`ast.Link` and `ast.Image`**

- `Destination []byte` → `Destination text.Value`
- `Title []byte` → `Title text.MultilineValue`
- `NewLink()` → `NewLink(destination text.Value, opts ...LinkOption)`
- `NewImage(link *Link)` → `NewImage(destination text.Value, opts ...LinkOption)`
- Optional title: `ast.WithLinkTitle(title)`
- Optional reference: `ast.WithLinkReference(kind, value)`

**`ast.AutoLink`**

- `AutoLinkType AutoLinkType`, `Protocol []byte`, and `value *Text` removed.
- `URL(source []byte) []byte` and `Label(source []byte) []byte` methods removed.
- Now has three `text.Value` fields: `Destination` (full href, email includes `mailto:`), `Label` (display text), `Text` (raw source text).
- `NewAutoLink(typ AutoLinkType, value *Text)` → `NewAutoLink(destination, label text.Value, opts ...AutoLinkOption)`
- Optional source text: `ast.WithAutoLinkText(text)`

**`ast.ReferenceLink`**

- `Type ReferenceLinkType` → `ReferenceLinkKind ReferenceLinkKind`
- Constants renamed: `ReferenceLinkFull/Collapsed/Shortcut` → `ReferenceLinkKindFull/Collapsed/Shortcut`
- `Value []byte` → `Value text.MultilineValue`

**`ast.HTMLBlock`**

- `HTMLBlockType` type renamed to `HTMLBlockKind`.
- Constants renamed: `HTMLBlockType1`..`HTMLBlockType7` → `HTMLBlockKind1`..`HTMLBlockKind7`.

**`ast.ListItem`**

- `Offset int` (public field) → unexported; access via `Offset() int` / `SetOffset(int)`.
- `NewListItem(offset int)` → `NewListItem()`

**`ast.LinkReferenceDefinition`**

- `Label/Destination/Title []byte` → `Label text.MultilineValue`, `Destination text.Value`, `Title text.MultilineValue`

**`extension/ast.DefinitionList`**

- `Offset int` and `TemporaryParagraph *Paragraph` (public fields) → unexported; access via `Offset()` / `SetOffset()` / `TemporaryParagraph()` / `SetTemporaryParagraph()`.
- `NewDefinitionList(offset int, para *Paragraph)` → `NewDefinitionList()`

**`extension/ast.TableCell`**

- `NewTableCell()` → `NewTableCell(alignment Alignment)` (alignment is now a required argument)

**`extension/ast.TableHeader`**

- `NewTableHeader(row *TableRow)` → `NewTableHeader()` (child nodes must be moved manually)

### `text` package

The `text.Segments` type (`*Segments` holding `[]Segment`) is no longer part of the public `Node` API.

New types for representing text values:
- `text.Value` — a single contiguous source span or a literal string
- `text.Index` — a raw `(Start, Stop)` index pair
- `text.MultilineValue` — a value that may span multiple source lines
- `text.Lines` — a list of source `Segment`s for block-level content (e.g. code blocks)

`text.Reader.FindClosure()` and `text.FindClosureOptions` have been removed (they were moved to parser-internal use only).

### `parser` package

`parser.Parser.Parse(reader, opts ...ParseOption)` has been simplified. The `opts ...ParseOption` parameter is removed. `reader` is now `source []byte`.

Attributes other than string type are no longer supported.

- `goldmark_v1_attribute` build tag allows using v1-compatible attributes.
  - You can get parsed value as `any` using `Attribute#Any(source)`
    ```go no-run
    attr, ok := node.Attribute("data-count")
    v := attr.Any(source) // returns float64 if the attribute value is a number
    ```

### Task list extension

The `extension/ast.TaskCheckBox` inline node no longer exists. Task state is stored as a `text.Value` attribute on the `ListItem` node. Use `extension.IsTask(node)` and `extension.TaskStatusOf(node)` to inspect task items.

### New in v2

- `ast.N(node Node, children ...any) Node` — builder helper that appends child nodes (or strings) to a node, useful for programmatically constructing an AST.
- `ast.BlockNode` / `ast.InlineNode` interfaces for type-safe node categorization.
- `parser.Parser.ParseStringSource()` and `renderer.Renderer.RenderStringSource()` convenience methods.
- `renderer.NodeRendererDecorator[W any]` for decorating a node renderer.

## Donation

BTC: 1NEDSyUmo4SMTDP83JJQSWi1MvQUGGNMZB

## License

MIT

## Author

Yusuke Inuzuka
