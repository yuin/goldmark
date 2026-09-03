# Technical Documentation: `extension/linkify.go`

## Overview

The `extension/linkify.go` package extension for `goldmark/v2` provides an inline parser that automatically detects URLs, `www.` domain names, and email addresses in plain text and converts them into Markdown auto-link AST nodes (`ast.AutoLink`).

---

## Key Components

### 1. Configuration & Functional Options

The parser uses a functional option pattern driven by the `linkifyConfig` struct:

```go
type linkifyConfig struct {
	AllowedProtocols [][]byte
	URLRegexp        *regexp.Regexp
	WWWRegexp        *regexp.Regexp
	EmailRegexp      *regexp.Regexp
}
```

#### Option Functions
- **`WithAllowedProtocols[T []byte | string](value []T) LinkifyParserOption`**: Specifies explicit protocols allowed for matching (e.g., `http:`, `https:`). Protocols must end with a colon (`:`). Uses Go generics to accept both byte slices and strings.
- **`WithURLRegexp(value *regexp.Regexp) LinkifyParserOption`**: Sets a custom regular expression pattern for matching standard URLs containing a protocol scheme.
- **`WithWWWRegexp(value *regexp.Regexp) LinkifyParserOption`**: Sets a custom regular expression pattern for URLs starting with `www.`.
- **`WithEmailRegexp(value *regexp.Regexp) LinkifyParserOption`**: Sets a custom regular expression pattern for email address detection.

---

### 2. Extension Initialization

- **`NewLinkifyParser(opts ...LinkifyParserOption) parser.Extension`**: Returns a new `parser.Extension` configured with the supplied options.
- **`LinkifyParser`**: Pre-configured global instance of `parser.Extension` using default settings.

#### Extension Registration

```go
func (e *linkifyParserExtension) ParserOptions(_ *parser.Config) []parser.Option
```
Registers `linkifyParser` as an inline parser using `util.Prioritized` with a priority weight of **999**.

---

### 3. Inline Parser Logic (`linkifyParser`)

The `linkifyParser` implements `parser.InlineParser`.

#### Triggers
```go
func (s *linkifyParser) Trigger() []byte
```
Triggers inline parsing when any of the following characters are encountered:
- `' '` (space / white space / line start)
- `'*'`
- `'_'`
- `'~'`
- `'('`

---

#### Detailed Parsing Process (`Parse` Method)

When `Parse(parent ast.Node, block text.Reader, pc parser.Context)` is invoked:

1. **Context Check**: Returns `nil` immediately if currently inside a link label (`pc.IsInLinkLabel()`).
2. **Delimiter Handling**: Peeks at the line slice. If the line begins with one of the trigger characters (`' '`, `'*'`, `'_'`, `'~'`, `'('`), it consumes 1 byte as a prefix delimiter.
3. **URL Detection**:
   - Checks against allowed protocols (`AllowedProtocols`) or defaults (`protoHTTP`, `protoHTTPS`, `protoFTP`).
   - If a scheme prefix matches, evaluates `URLRegexp`.
4. **`www.` Detection**:
   - If no protocol URL matches, checks if the text starts with `www.`.
   - If matched, runs `WWWRegexp` and flags `isWWW = true`.
5. **Email Detection**:
   - If no URL or `www.` prefix matches, and the first character is not a punctuation mark (`util.IsPunct`), it attempts email detection (`isEmail = true`).
   - Evaluates either `EmailRegexp` or the internal `findEmailIndex()` helper function.
   - Validates that the detected email contains `@` and `.`, does not end with a period, and is not followed by trailing invalid domain characters like `-` or `_`.
6. **Trailing Character Trimming**:
   - Trims trailing punctuation (`?`, `!`, `.`, `,`, `:`, `*`, `_`, `~`).
   - Balances unclosed trailing parentheses `)` against internal opening parentheses `(`.
   - Detects and handles HTML entities ending with `;` (e.g., `&...;`).
7. **AST Node Generation**:
   - If a prefix delimiter was consumed, appends a plain text node (`ast.NewText`) to `parent`.
   - Advances the `block` reader by the consumed byte count.
   - Prepends appropriate URI schemes:
     - Emails: Prepends `mailto:`
     - `www.` URLs: Prepends `http://`
     - Regular URLs: Preserves original matched text scheme
   - Returns an `ast.NewAutoLink` containing the formatted destination URL and original raw text.

---

### 4. Helper Variables and Email Parser

#### Default Regular Expressions
- **`urlRegexp`**: Matches `http://`, `https://`, or `ftp://` schemes followed by host, port, and path/query/fragment components.
- **`wwwURLRegxp`**: Matches `www.` hostnames followed by domain and path/query/fragment components.
- **`emailDomainRegexp`**: Validates the domain part of an email address following the `@` symbol.

#### Email Scanning Utility (`findEmailIndex`)
```go
func findEmailIndex(b []byte) int
```
Scans raw byte slices for valid email addresses without relying entirely on complex regular expressions for local parts:
1. Validates local-part characters using the 256-element bitmask lookup table (`emailTable`).
2. Verifies the presence of the `@` symbol.
3. Validates the domain segment using `emailDomainRegexp`.
4. Returns the end byte index of the valid email, or `-1` if invalid.