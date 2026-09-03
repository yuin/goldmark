# Technical Documentation: `parser/auto_link.go`

## Overview

The `parser/auto_link.go` file provides an inline parser implementation for Goldmark (`github.com/yuin/goldmark`). Its primary purpose is to identify and parse **autolinks**—URLs or email addresses delimited by angle brackets (`<` and `>`) as defined in the Markdown specification—and construct corresponding AST (`ast.AutoLink`) nodes.

---

## Key Components

### 1. Types and Factory Functions

#### `autoLinkParser`
* **Type**: `struct` (empty)
* **Description**: Internal struct implementing the `InlineParser` interface to process angle-bracket enclosed autolinks.
* **Singleton Instance**: `defaultAutoLinkParser` (`*autoLinkParser`)

#### `NewAutoLinkParser()`
* **Signature**: `func NewAutoLinkParser() InlineParser`
* **Returns**: An `InlineParser` (the `defaultAutoLinkParser` instance).
* **Purpose**: Serves as the public constructor to register the autolink parser within the Markdown parsing pipeline.

---

### 2. Lookup Tables and Regular Expressions

To maximize performance, `autoLinkParser` uses byte lookup tables (`[256]uint8`) for scheme and character set validation, avoiding regex overhead where possible.

#### `urlTable` (`[256]uint8`)
A 256-byte lookup array used by `findURLIndex` to validate characters in URIs. It uses bitwise flags:
* `& 7 == 7`: Valid initial character for a URI scheme (ASCII letters `A-Z`, `a-z`).
* `& 4 == 4`: Valid subsequent character for a URI scheme (ASCII letters, digits `0-9`, `.`, `+`, `-`).
* `& 1 == 1`: Valid character within the URI body (excludes control characters `\x00-\x20`, `<`, `>`, etc.).

#### `emailTable` (`[256]uint8`)
A 256-byte lookup array used by `findEmailIndex` to validate characters in the local part (before `@`) of an email address.
* `& 1 == 1`: Valid email local-part character.

#### `emailDomainRegexp`
* **Type**: `*regexp.Regexp`
* **Pattern**: `^[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*`
* **Purpose**: Validates the domain portion (after `@`) of an email address. Matches domain labels compliant with standard RFC hostnames (1–63 characters per label separated by dots).

---

### 3. Parsing Methods

#### `Trigger()`
* **Signature**: `func (s *autoLinkParser) Trigger() []byte`
* **Returns**: `[]byte{'<'}`
* **Purpose**: Informs the inline parsing engine that this parser should be invoked when the `<` character is encountered in the text reader.

#### `Parse()`
* **Signature**: `func (s *autoLinkParser) Parse(_ ast.Node, block text.Reader, _ Context) ast.Node`
* **Parameters**:
  * `_ ast.Node`: Parent node (unused).
  * `block text.Reader`: The text reader pointing to the current position in the source document.
  * `_ Context`: Parsing context (unused).
* **Returns**: An `ast.Node` representing the parsed autolink, or `nil` if parsing fails.

---

### 4. Internal Helper Functions

#### `findURLIndex(b []byte) int`
Validates whether the slice `b` begins with a valid absolute URL. It implements the standard Markdown autolink URL pattern: `[A-Za-z][A-Za-z0-9.+-]{1,31}:[^<>\x00-\x20]*`.

**Algorithm**:
1. Checks if `b` is non-empty and starts with a valid scheme character (`urlTable[b[0]] & 7 == 7`).
2. Scans subsequent scheme characters (`urlTable[c] & 4 == 4`).
3. Enforces scheme length constraints: scheme length must be between 2 and 32 characters (1 to 31 extra characters before `:`).
4. Verifies the colon `:` delimiter immediately following the scheme.
5. Scans remaining valid URI body characters (`urlTable[c] & 1 == 1`).
6. Returns the index after the last valid URI character, or `-1` if invalid.

#### `findEmailIndex(b []byte) int`
Validates whether the slice `b` begins with a valid email address.

**Algorithm**:
1. Scans contiguous local-part characters using `emailTable` (`emailTable[c] & 1 == 1`).
2. Checks that at least one local character was matched and is immediately followed by `@`.
3. Validates the remaining slice after `@` using `emailDomainRegexp`.
4. Returns the index after the end of the email domain match, or `-1` if invalid.

---

## Complete Parsing Flow (`Parse` Logic)

When `autoLinkParser.Parse` is invoked at a trigger position `<`:

```
   Peek line from text reader (starts with '<')
                     │
                     ▼
       Try matching email on line[1:]
       via findEmailIndex(line[1:])
                     │
        ┌────────────┴────────────┐
   Match found              No match found
        │                         │
        ▼                         ▼
   Set isEmail = true     Try matching URL on line[1:]
   Set stop index         via findURLIndex(line[1:])
        │                         │
        └────────────┬────────────┘
                     │
            Is stop index < 0?
             ┌───────┴───────┐
            Yes              No
             │               │
             ▼               ▼
        Return nil    Adjust stop index (stop++)
                             │
                             ▼
                  Does line[stop] == '>'?
                     ┌───────┴───────┐
                    No              Yes
                     │               │
                     ▼               ▼
                Return nil    Extract text segment values:
                              - textVal  (includes '<' and '>')
                              - labelVal (inner content)
                                     │
                                     ▼
                              Advance text reader past '>'
                                     │
                                     ▼
                              Construct destination (dest):
                              - Email: "mailto:" + email string
                              - URL: inner segment index
                                     │
                                     ▼
                              Create ast.NewAutoLink(...)
                              and return node
```

1. **Peek Line**: Reads the remaining bytes on the current line and gets the current `text.Segment`.
2. **Email Match First**: Attempts `findEmailIndex` on `line[1:]` (skipping the initial `<`).
3. **URL Match Fallback**: If email scanning fails (`stop < 0`), attempts `findURLIndex` on `line[1:]`.
4. **Boundary Verification**:
   * If neither matched, returns `nil`.
   * Increments `stop` by 1 to offset the initial `<` slice shift.
   * Ensures `stop < len(line)` and `line[stop] == '>'`.
5. **Segment Value Construction**:
   * `textVal`: Full span including `<` and `>` using `text.NewSingleLineValueFromIndex`.
   * `labelVal`: Content inside `<` and `>` without the delimiters.
6. **Advance Reader**: Calls `block.Advance(stop + 1)` to consume the autolink from input.
7. **Destination Assignment**:
   * If email: Creates a `SingleLineValue` string prefixed with `"mailto:"`.
   * If URL: Uses the segment range corresponding to the inner content.
8. **Node Creation**: Constructs an `*ast.AutoLink` node with `dest`, `labelVal`, and the option `ast.WithAutoLinkText(textVal)`, then returns it.