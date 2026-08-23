## Key point of CommonMark spec
- In most cases, punctuation characters can be escaped with a backslash (`\`) to prevent them from being interpreted as Markdown syntax. For example, `\*` will render as `*` instead of starting an emphasis.
- Spec defines sets of character categories. Use `util` package to check if a character belongs to a specific category. For example, `util.IsPunct` can be used to check if a character is a punctuation character. DO NOT use standard library functions like `unicode.IsPunct` as they may not cover all cases defined in the CommonMark spec.
- Most of inline elements (like emphasis, links, etc.) can exist within multiple lines. And they can be nested. So inline elements probably have multiple divided segments:

  ```markdown
  > [lin
  > nk](https://example.com)
  ```

  In this case, the link element has two divided segments: `lin` and `nk`. When parsing, you should keep track of these segments and combine them when necessary. This kind of elements should have `text.MultiLineValue` instead of `text.SingleLineValue`.
- Paragraph rendering can be changed by parent elements. For example, a paragraph inside a tight list item should not have `<p>` tags, while a paragraph inside a block quote should have `<p>` tags.
- Tabs can be 1,2,3,4 spaces or raw tab character, depending on its position. When parsing block elements, you should aware of this and calculate the correct indentation level.
