---
name: migrate-goldmark-extension-v1-to-v2
context: fork
description: Migrate goldmark extension from goldmark v1 to v2.
allowed-tools: Bash Read
---

# migrate-goldmark-extension-v1-to-v2
## Description
This skill helps you migrate a goldmark (https://github.com/yuin/goldmark) extension from version 1 to version 2. It
provides guidance on the changes needed to update your project to be compatible with the new version of goldmark.

## Knowledges

- [CommonMark key points](../../references/commonmark-key-points.md) : List of key points of CommonMark spec that you should be aware of when implementing a goldmark extension.
- [Breaking changes in v2](../../references/breaking-changes-in-v2.md) : List of breaking changes in goldmark v2 that you should be aware of when migrating your extension from v1 to v2.
- [How to create an extension](./references/how-to-create-an-extension.md) : Guide on how to create a goldmark extension in v2, including the new extension pattern and how to implement parser and renderer extensions.

## Migration steps
### Overview of the migration process

- Create a migration plan for the extension.
- **MUST** ask human to confirm that the migration plan is acceptable before proceeding with the migration.
- **MUST** ask human to how to test the extension after migration before proceeding with the migration.
  - e.g. : "How do you want to test the extension after migration? Do you have any test cases or examples that you want to use for testing?"
- Execute the migration plan to update the extension code to be compatible with goldmark v2.
- Update the test cases to ensure that the extension works as expected with goldmark v2.
- Test the extension with goldmark v2 to ensure that it works as expected. If there are any issues, fix them and re-test until the extension works as expected.
- Update the documentation to reflect any changes made during the migration process.

### Create a migration plan
#### Task

- Make sure you have read and understood the [Breaking changes in v2](../../references/breaking-changes-in-v2.md) document.
- Make sure you have read and understood the [How to create an extension](./references/how-to-create-an-extension.md) document.
- You create a ./features/goldmark-migration-plan.md file that contains a migration plan for the extension. 

#### Key points to consider when migrating your extension
##### Extension options

- If the extension uses "unified" options for both parser and renderer, they should be split into separate options for each.
  - e.g. :
    - v1
      ```go
      type Option interface {
          myOption()
      }

      type ParserOption interface {
          Option
          applyParserOption(*parserConfig)
      }

      type RendererOption interface {
          Option
          applyRendererOption(*rendererConfig)
      }

      func New(opts ...Option) goldmark.Extender { // takes unified options
          // ...
      }
      ```
    - v2
      ```go
      type ParserOption interface {
          applyParserOption(*parserConfig)
      }

      type HTMLRendererOption interface { // explicitly named for **HTML**
          applyRendererOption(*htmlRendererConfig) // you can access the shared renderer config like `XHTML` or `Unsafe` in the renderer config
      }

      func NewParser(opts ...ParserOption) parser.Extension { // takes parser options
          // ...
      }

      var Parser = NewParser() // Default instance of parser extension

      func NewHTMLRenderer(opts ...HTMLRendererOption) html.Extension { // takes renderer options
          // ...
      }

      var HTMLRenderer = NewHTMLRenderer() // Default instance of renderer extension
      ```

##### AST nodes

- use `text.Value`(single line), `text.MultiLineValue`(multi-line) instead of `[]byte` for values that can be parsed from source text in inline AST nodes.
  - In your parser, you must choose `text.Decoder` implementation to decode the source value
    - `text.IdentityDecoder` : for raw contents like inline HTMLs, inline code, etc.
    - `reader.Decoder` : other contents like text, links, etc. This decoder decodes entity references, `\` escapes, etc.
  - In most cases, you will choose `text.Decoder`. **DO NOT** use `text.IdentityDecoder` unless you have a clear intention to do so.
- use `text.Lines` instead of `[]text.Segment` for values in block AST nodes that have **raw contents** like HTML blocks, code blocks, etc. 
- Properties in AST Dump should be `text.Value` as possible.
  - e.g.
    - OK:
      ```
      // Dump implements Node.Dump.
      func (n *Text) Dump(_ []byte) *NodeDump {
      	m := map[string]any{
      		"Value": n.Value, // text.Value
      	}
      	fs := textFlagsString(n.flags)
      	if len(fs) != 0 {
      		m["Flags"] = fs
      	}
      	return NewNodeDump(n, m)
      }
      ```
    - Not OK:
      ```
      // Dump implements Node.Dump.
      func (n *Text) Dump(source []byte) *NodeDump {
      	m := map[string]any{
      		"Value": n.Value.Str(source), // string
      	}
      	fs := textFlagsString(n.flags)
      	if len(fs) != 0 {
      		m["Flags"] = fs
      	}
      	return NewNodeDump(n, m)
      }
      ```
- In v2, attribute values are `text.Value` which has almost the same specification as HTML attributes.
  - Therefore, if the project were using non-string attributes in v1, human must decide on one of the following policies:
     - Use the `goldmark_v1_attribute` build tag to continue using v1 attributes as they are.
     - Convert attribute values to strings to comply with the v2 specification.
  - **MUST** ask human to decide on one of the above policies before proceeding with the migration.

##### Parsing

- In v2, all nodes have a start position. goldmark/v2 automatically sets the start position to the node. However, if you want to customize the start position, you need to call `SetPos` appropriately.

##### HTML Rendering

- `text.Value` and `text.Lines` can be rendered using the `WriteTo` method whenever possible. Also, the output destination of `WriteTo` should use `html.ContextHTMLWriter(rc)` or `html.ContextTextWriter(rc)`.
   - e.g. :
     ```go
     tw := html.ContextTextWriter(rc)
     _, _ = n.Value.WriteTo(tw, source)
     ```
   - `WriteTo` is fast because it does not allocate new memory. On the other hand, if you write `Value` directly like `tw.Write(n.Value.Value(source))`, it may copy the contents of `Value`, which can degrade performance.

##### Recommended naming convention(for public stuff)

- Use `myext.NewParser()` and `myext.NewHTMLRenderer()` for the extension constructors.
  - e.g. : `meta` extension
    - `meta.NewParser()`, `meta.NewHTMLRenderer()`
- Use `myext.Parser` and `myext.HTMLRenderer` as the default extension values.
  - e.g. : `var Parser = NewParser()`, `var HTMLRenderer = NewHTMLRenderer()`
- Use `myext.ParserOption` and `myext.HTMLRendererOption` for functional options.
   - e.g. : `type ParseOption func(*parserConfig)`, `type HTMLRendererOption func(*htmlRendererConfig)`

### Execute migration plan
- Make sure you are on a branch that is not `main` or `master`. User must create a new branch like 'v2' to work on the migration before using this skill.
  - If you are on `main` or `master`, **STOP** this skill and ask human to create a new branch like 'v2' to work on the migration.
- Make sure `go.mod` file is updated to use `github.com/yuin/goldmark/v2` instead of `github.com/yuin/goldmark`. User must add `goldmark/v2` before using this skill.
  - If `go.mod` file is not updated, **STOP** this skill and ask human to update `go.mod` file to use `github.com/yuin/goldmark/v2` instead of `github.com/yuin/goldmark`.
- Update the module path in your `go.mod` file with new major version. For example, change `github.com/you/yourextension` to `github.com/you/yourextension/v2`.
  - **MUST** ask human to make sure that the module path is updated in `go.mod` file before proceeding with the migration.
    - If human confirms that the module path is updated, proceed with the migration, otherwise, **STOP** this skill and ask human to update the module path in `go.mod` file with new major version.

