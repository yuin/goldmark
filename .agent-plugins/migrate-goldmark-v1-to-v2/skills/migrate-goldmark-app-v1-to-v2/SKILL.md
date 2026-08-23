---
name: migrate-goldmark-app-v1-to-v2
context: fork
description: Migrate your application from goldmark v1 to v2.
allowed-tools: Bash Read
---

# migrate-goldmark-app-v1-to-v2
## Description
This skill helps you migrate a goldmark (https://github.com/yuin/goldmark) extension from version 1 to version 2. It
provides guidance on the changes needed to update your project to be compatible with the new version of goldmark.

## Knowledges

- [CommonMark key points](../../references/commonmark-key-points.md) : List of key points of CommonMark spec that you should be aware of when implementing a goldmark extension.
- [Breaking changes in v2](../../references/breaking-changes-in-v2.md) : List of breaking changes in goldmark v2 that you should be aware of when migrating your extension from v1 to v2.

## Migration steps
### Overview of the migration process

- Create a migration plan for the application.
- **MUST** ask human to confirm that the migration plan is acceptable before proceeding with the migration.
- **MUST** ask human to how to test the application after migration before proceeding with the migration.
  - e.g. : "How do you want to test the application after migration? Do you have any test cases or examples that you want to use for testing?"
- Execute the migration plan to update the application code to be compatible with goldmark v2.
- Update the test cases to ensure that the application works as expected with goldmark v2.
- Test the application with goldmark v2 to ensure that it works as expected. If there are any issues, fix them and re-test until the application works as expected.
- Update the documentation to reflect any changes made during the migration process.

### Create a migration plan
#### Task

- Make sure you have read and understood the [Breaking changes in v2](../../references/breaking-changes-in-v2.md) document.
- Make sure you have read and understood the [How to create an extension](./references/how-to-create-an-extension.md) document.
- You create a ./features/goldmark-migration-plan.md file that contains a migration plan for the extension. 

#### Extension

- If the application contains own extensions, you can use `/migrate-goldmark-extension-v1-to-v2` skill to migrate the extensions. 
- If the application contains third-party extensions, you need to check if the extensions provide v2 compatible version.  - If not, **STOP** the migration.

#### Key points to consider when migrating your application
##### goldmark.Markdown alternatives

- In v2, `goldmark.Markdown` is removed. Therefore, you need to replace it with one of the following two patterns:
  - Pattern 1: Use `parser.Parser` and `renderer.Renderer` to create your own `goldmark.Markdown` alternative.
    - Example:
      ```go
      // MarkdownToStringFunc is a function type that converts markdown to HTML.
      type MarkdownToStringFunc func(source string) (string, error)
      
      // NewMarkdownToStringFunc returns a MarkdownToStringFunc that uses the given parser and renderer.
      func NewMarkdownToStringFunc(p parser.Parser, r html.Renderer) MarkdownToStringFunc {
      	return func(source string) (string, error) {
      		var buf bytes.Buffer
      		b := util.StringToReadOnlyBytes(source)
      		doc := p.Parse(b)
      		if err := r.Render(&buf, b, doc); err != nil {
      			return "", err
      		}
      		return buf.String(), nil
      	}
      }
  - Pattern 2: Use `parser.Parser` and `renderer.Renderer` separately.
    - Example:
      ```go
      var buf bytes.Buffer
      p := parser.New(parser.WithAttribute(), parser.WithExtensions(extension.StrikethroughParser))
      r := html.New(html.WithXHTML(), html.WithUnsafe(), html.WithExtensions(extension.StrikethroughHTMLRenderer))
      doc := p.Parse(b)
      err := r.Render(&buf, b, doc)
      ```

##### AST

- In v1, AST values are mostly 'raw' values; entities references and `\` escapes are not resolved. In v2, AST values are resolved values.
  - You need to check if your application relies on the 'raw' values of AST nodes.
    - If your application relies on the resolved values of AST nodes, you should replace `Value.Bytes` and `Value.Str` methods with `Value.Value` method.
    - Otherwise, you should replace `Value.Bytes` and `Value.Str` methods with `Value.Value` method.
- In v2, to make the AST more semantic, some breaking changes have occurred.
  - Please refer to the AST-related section of [Breaking changes in v2](../../references/breaking-changes-in-v2.md).

##### Parsing

- To customize the ID generation, you need to use `parser.IDGenerator` instead of `parser.IDs`.

##### Rendering

- In v2, `renderer.Renderer` has `renderer.Context`; if the application mimics context for rendering, it should be updated to use `renderer.Context`.
