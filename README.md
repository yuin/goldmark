# goldmark

[![https://pkg.go.dev/github.com/yuin/goldmark/v2](https://pkg.go.dev/badge/github.com/yuin/goldmark/v2.svg)](https://pkg.go.dev/github.com/yuin/goldmark/v2)
[![https://github.com/yuin/goldmark/actions?query=branch%3Av2+workflow%3Atest](https://github.com/yuin/goldmark/actions/workflows/test.yaml/badge.svg?branch=v2&event=push)](https://github.com/yuin/goldmark/actions?query=branch%3Av2+workflow%3Atest)
[![https://coveralls.io/github/yuin/goldmark?branch=v2](https://coveralls.io/repos/github/yuin/goldmark/badge.svg?branch=v2)](https://coveralls.io/github/yuin/goldmark?branch=v2)

> A Markdown parser written in Go. Easy to extend, standards-compliant, well-structured.

goldmark is compliant with CommonMark 0.31.2.

- [goldmark playground](https://yuin.github.io/goldmark/playground/v2/) : Try goldmark online. This playground is built with WASM(5-10MB).

There is also a Rust version of goldmark: [rushdown](https://github.com/yuin/rushdown)

## v2 status
v2 is still in the early stages of release. If you are using an extension that does not support v2, please use v1.

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

In particular, as Markdown documents have come to be used as Lingua franca for AI, there is an increasing need to analyze Markdown documents semantically. For the same reason, there are also increasing use cases for generating Markdown documents rather than parsing them. The use of CLI in AI agents is increasing, also a growing need to convert to formats other than HTML.

Breaking changes to an extensible library like goldmark have a huge impact, as third-party extensions will no longer work. Therefore, I have avoided making breaking changes for a long time.

It has been more than 7 years since goldmark was created, and technical debt has been accumulating. In the meantime, the Go language specification has changed significantly, including the introduction of generics. I believe that the changes in use cases, represented by AI, are a good opportunity to fundamentally review the design of goldmark, and I have decided to make breaking changes.

## v2 and v1 differences

- v2 focuses on building a more semantic AST.
- v2 uses generics.
- v2 clearly separates the parser and renderer. This makes it easier to implement rendering to formats other than HTML.
- v2 allows you to programmatically build an AST. And you can render the constructed AST to another format.
- v2 has all nodes hold the start position. In the future, third-party extensions that support v2 are also expected to hold the start position.
- The core parsing algorithm is the same as v1. Third-party extensions must support v2, but the most complex parsing part can be used almost as it is.

## Maintenance policy
This project will maintain bug fixes, including security fixes, up to one major version prior to the latest major version.

## Migrating from v1 to v2
You can use LLMs to migrate your code from v1 to v2. 

Claude Code / Copilot CLI

```bash
/plugin marketplace add yuin/goldmark@v2
/plugin install migrate-goldmark-v1-to-v2@yuin-goldmark-v2
```

Migrating your goldmark extension projects:

```
/migrate-goldmark-v1-to-v2:migrate-goldmark-extension-v1-to-v2
```

Migrating your applications using goldmark:

```
/migrate-goldmark-v1-to-v2:migrate-goldmark-app-v1-to-v2
```

These skills will create a migration plan for your project and execute the migration plan to update your code to be compatible with goldmark v2.

Of course, even you can migrate manually if they understand these contents :)
See `.agent-plugins` directory for the implementation of these skills.

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
        n.SetAttribute("class", text.NewMultiLineValue("greeting", text.IdentityDecoder))
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

p := parser.New(parser.WithAttribute(), parser.WithExtensions(extension.StrikethroughParser))
r := html.New(html.WithXHTML(), html.WithUnsafe(), html.WithExtensions(extension.StrikethroughHTMLRenderer))

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

### Parse options

| Functional option | Type | Description |
| ----------------- | ---- | ----------- |
| `parser.WithContext` | `parser.Context` | Context for parsing. |
| `parser.WithPrettyPrint` | `[]ast.PrettyPrintOption` | Prints the parsed AST tree to stdout (or a custom `io.Writer` via `ast.PrettyPrintOption`) for debugging. |


### HTML Renderer options

| Functional option | Type | Description |
| ----------------- | ---- | ----------- |
| `html.WithLineBreakStrategy` | `html.LineBreakStrategy` | Soft line breaks are rendered as a newline. Some asian users will see it as an unnecessary space. With this option, you can change the behavior. |
| `html.WithHardWraps` | `-` | Render newlines as `<br>`.|
| `html.WithIsInTightBlockFunc` | `html.IsInTightBlockFunc` | Function that determines whether a node is in a tight block. |
| `html.WithNodeRenderer` | `ast.NodeKind`, `html.NodeRenderer` | Add a node renderer for a specific node kind. |
| `html.WithNodeRenderers` | `map[ast.NodeKind]html.NodeRenderer` | Add node renderers for specific node kinds. |
| `html.WithNodeRendererDecorator` | `ast.NodeKind`, `html.NodeRendererDecorator` | Add a decorator for a node renderer. |
| `html.WithNodeRendererDecorators` | `map[ast.NodeKind]html.NodeRendererDecorator` | Add decorators for node renderers. |
| `html.WithXHTML` | `-` | Render as XHTML. |
| `html.WithUnsafe` | `-` | By default, goldmark does not render raw HTML or potentially dangerous links. With this option, goldmark renders such content as written. |
| `html.WithExtensions` | `[]html.Extension` | Enables parser extensions. |

#### Defined line break strategies

| Style | Description |
| ----- | ----------- |
| `SimpleEastAsianLineBreakStrategy` | Soft line breaks are ignored if both sides of the break are east asian wide character. This behavior is the same as [`east_asian_line_breaks`](https://pandoc.org/MANUAL.html#extension-east_asian_line_breaks) in Pandoc. |
| `CSSText3LineBreakStrategy` | This option implements CSS text level3 [Segment Break Transformation Rules](https://drafts.csswg.org/css-text-3/#line-break-transform) with [some enhancements](https://github.com/w3c/csswg-drafts/issues/5086). |

**Example of `SimpleEastAsianLineBreakStrategy`**

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

**Example of `CSSText3LineBreakStrategy`**

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

### Render options

| Functional option | Type | Description |
| ----------------- | ---- | ----------- |
| `renderer.WithContext` | `renderer.Context` | Context for rendering. Passed to `Renderer[W].Render` as a `RenderOption`. |

### Built-in extensions

Each extension is a pair of a parser extension and an HTML renderer extension. 

- `Table(Parser|HTMLRenderer)`
    - [GitHub Flavored Markdown: Tables](https://github.github.com/gfm/#tables-extension-)
- `Strikethrough(Parser|HTMLRenderer)`
    - [GitHub Flavored Markdown: Strikethrough](https://github.github.com/gfm/#strikethrough-extension-)
- `LinkifyParser`
    - [GitHub Flavored Markdown: Autolinks](https://github.github.com/gfm/#autolinks-extension-)
    - This extension only affects parsing; autolinks render through the same `ast.AutoLink` renderer as CommonMark autolinks, so it has no HTML renderer half.
- `TaskList(Parser|HTMLRenderer)`
    - [GitHub Flavored Markdown: Task list items](https://github.github.com/gfm/#task-list-items-extension-)
- `GFM(Parser|HTMLRenderer)`
    - This extension enables Table, Strikethrough, Linkify and TaskList.
    - This extension does not filter tags defined in [6.11: Disallowed Raw HTML (extension)](https://github.github.com/gfm/#disallowed-raw-html-extension-).
      If you need to filter HTML tags, see [Security](#security).
    - If you need to parse github emojis, you can use [goldmark-emoji](https://github.com/yuin/goldmark-emoji) extension.
    - [goldmark-github-slugger](https://github.com/yuin/goldmark-github-slugger) generates the same heading ids as GitHub.
    - [goldmark-alert](https://github.com/yuin/goldmark-alert) adds support for GitHub-style alert blocks.
- `DefinitionList(Parser|HTMLRenderer)`
    - [PHP Markdown Extra: Definition lists](https://michelf.ca/projects/php-markdown/extra/#def-list)
- `Footnote(Parser|HTMLRenderer)`
    - [PHP Markdown Extra: Footnotes](https://michelf.ca/projects/php-markdown/extra/#footnotes)
- `TypographerParser`
    - This extension substitutes punctuations with typographic entities like [smartypants](https://daringfireball.net/projects/smartypants/).
    - This extension only affects parsing (it emits already-substituted text), so it has no HTML renderer half.

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

Like other CommonMark attribute values (e.g., FencedCodeBlock language, link title), attribute values can contain entity references and symbol escapes with `\`.


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

You can override the default substitutions via `extension.WithTypographicSubstitutions`.

```go
import (
    "github.com/yuin/goldmark/v2/extension"
    "github.com/yuin/goldmark/v2/parser"
)
_ = parser.New(
        parser.WithExtensions(extension.NewTypographerParser(
            extension.WithTypographicSubstitutions(extension.TypographicSubstitutions{
                extension.LeftSingleQuote:  "&sbquo;",
                extension.RightSingleQuote: "", // "" disables a substitution
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

This extension has some options. All of them are `extension.FootnoteHTMLRendererOption`s, i.e. they configure `extension.NewFootnoteHTMLRenderer(opts...)`, not the parser:

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
    "github.com/yuin/goldmark/v2/renderer/html"
)

for _, path := range files {
    source := readAll(path)
    prefix := getPrefix(path)

    p := parser.New(parser.WithExtensions(extension.NewFootnoteParser()))
    r := html.New(
        html.WithExtensions(
            extension.NewFootnoteHTMLRenderer(
                extension.WithIDPrefix(prefix),
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
    "github.com/yuin/goldmark/v2/renderer/html"
    "github.com/yuin/goldmark/v2/util"
)

p := parser.New(parser.WithExtensions(extension.NewFootnoteParser()))
r := html.New(
    html.WithExtensions(
        extension.NewFootnoteHTMLRenderer(
            extension.WithIDPrefixFunction(func(n gast.Node) []byte {
                v, ok := n.OwnerDocument().Metadata()["footnote-prefix"]
                if ok {
                    return util.StringToReadOnlyBytes(v.(string))
                }
                return nil
            }),
        ),
    ),
)

for _, path := range files {
    source := readAll(path)
    doc := p.Parse(source)
    doc.AddMeta("footnote-prefix", getPrefix(path))
    // convert doc to HTML with r
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

Go1.27.0

```
BenchmarkMarkdown/GoMarkdown(not_CM)-16                      169           7165929 ns/op         2704039 B/op      27019 allocs/op
BenchmarkMarkdown/Lute-16                                     69          16476617 ns/op        13832888 B/op      32490 allocs/op
BenchmarkMarkdown/golang-commonmark-16                       172           6991769 ns/op         2703246 B/op      20129 allocs/op
BenchmarkMarkdown/goldmark/v2-16                             188           6156894 ns/op         2629375 B/op      12791 allocs/op
BenchmarkMarkdown/goldmark/v1-16                             176           6525718 ns/op         2539293 B/op      14471 allocs/op
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
- [goldmark-alert](https://github.com/yuin/goldmark-alert): An alert extension for the goldmark Markdown parser.
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
- [goldmark-cjk-friendly](https://github.com/tats-u/goldmark-cjk-friendly): Port of npm package [`remark-cjk-friendly` / `markdown-it-cjk-friendly`](https://github.com/tats-u/markdown-cjk-friendly) to goldmark. Similar to the `parser.WithEscapedSpace` [parser option](#parser-options), but you do not need to explicitly add `\ ` around `*` and `**`. You can combine this with `parser.WithEscapedSpace`.
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

import (
    gast "github.com/yuin/goldmark/v2/ast"
    "github.com/yuin/goldmark/v2/text"
)

// MyNode represents a custom inline element.
type MyNode struct {
    gast.BaseInline
    // Add fields for data that belongs to the node semantics.
    // Do NOT store parser-internal state here.
    MyField text.SingleLineValue
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
| `text.Value` | An interface for a single-line value or a multi-line value | `CodeSpan.Value` |
| `text.SingleLineValue` | The spec guarantees the value fits on a single line | Link destination (`[text](url)`), fenced code block info string |
| `text.MultiLineValue` | The spec allows the value to span multiple lines | Link title, code span content, raw HTML |
| (FYR) `text.Lines` | A special block element that holds raw, unparsed block content line-by-line | `CodeBlock.Value`, `HTMLBlock.Value` |

It is recommended to use `SingleLineValue` or `MultiLineValue` instead of the `text.Value` interface when defining AST nodes whenever possible. The reasons are:

- `text.Value` will require new memory allocation.
- Default values of `text.Value` are nil, but in many cases an empty string is more appropriate. Using an empty `SingleLineValue` or `MultiLineValue` avoids nil checks.

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

If you need to normalize a value, create your own `text.Value` implementation. For example, CommonMark requires code spans to trim surrounding whitespace and convert newlines to spaces; the `parser/code_span.go` uses a custom `text.Value` implementation that performs this normalization. In cases where 'normalization' is required like this, you should use the `text.Value` interface when defining your AST.


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

## Donation

BTC: 1NEDSyUmo4SMTDP83JJQSWi1MvQUGGNMZB

## License

MIT

## Author

Yusuke Inuzuka
