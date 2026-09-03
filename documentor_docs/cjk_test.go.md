# Technical Documentation: `cjk_test.go`

## Overview

The `cjk_test.go` file contains unit tests for the `goldmark` Markdown parsing library (version 2). The primary focus of this test suite is to verify parsing and rendering behaviors related to CJK (Chinese, Japanese, Korean / East Asian) text processing. 

Specifically, it tests two key capabilities:
1. **Escaped Spaces (`parser.WithEscapedSpace`)**: Handling escaped spaces (`\ `) to enable correct inline emphasis formatting without introducing unnecessary visible spaces around East Asian characters and punctuation.
2. **East Asian Line Break Strategies (`html.WithLineBreakStrategy`)**: Handling soft line breaks between East Asian wide characters, CJK punctuation, and Western (ASCII) characters using strategies such as `html.SimpleEastAsianLineBreakStrategy` and `html.CSSText3LineBreakStrategy`.

---

## Package and Dependencies

* **Package:** `goldmark_test`
* **Imports:**
  * `testing`: Standard Go testing package.
  * `github.com/yuin/goldmark/v2/extension`: Provides extensions such as `extension.NewLinkifyParser()`.
  * `github.com/yuin/goldmark/v2/parser`: Provides markdown parsing configuration options (e.g., `WithEscapedSpace`, `WithExtensions`).
  * `github.com/yuin/goldmark/v2/renderer/html`: Provides HTML renderer options and line break strategies (e.g., `WithXHTML`, `WithUnsafe`, `WithHardWraps`, `WithLineBreakStrategy`).
  * `github.com/yuin/goldmark/v2/testutil`: Helper utility (`testutil.DoTestCase`, `testutil.NewMarkdownToStringFunc`) for executing Markdown conversion test cases.

---

## Functions Detail

### 1. `TestEscapedSpace(t *testing.T)`

Tests how Goldmark parses emphasis syntax around East Asian punctuation with and without escaped space handling enabled, as well as its interaction with inline parsers like Linkify.

#### Test Cases Executed:

* **Test Case 1 (Default CommonMark behavior):**
  * **Configuration:** Standard parser, standard HTML renderer (`WithXHTML`, `WithUnsafe`).
  * **Behavior:** Emphasis syntax adjacent to CJK punctuation (`太郎は**「こんにちわ」**と言った`) without surrounding spaces is **not** rendered as standard emphasis (per CommonMark spec).
  * **Expected Output:** `<p>太郎は**「こんにちわ」**と言った\nんです</p>`

* **Test Case 2 (Unescaped Spaces):**
  * **Configuration:** Standard parser, standard HTML renderer.
  * **Behavior:** Spaces added around CJK emphasis allowed it to parse as `<strong>`, but preserved unnecessary literal spaces in output.
  * **Expected Output:** `<p>太郎は <strong>「こんにちわ」</strong> と言った\nんです</p>`

* **Test Case 3 (Escaped Spaces Enabled):**
  * **Configuration:** Parser configured with `parser.WithEscapedSpace()`, Renderer configured with `html.SimpleEastAsianLineBreakStrategy`.
  * **Behavior:** Escaped spaces (`\ `) trigger emphasis parsing without including literal spaces in the HTML output, while soft line breaks are removed between CJK characters.
  * **Expected Output:** `<p>太郎は<strong>「こんにちわ」</strong>と言ったんです</p>`

* **Test Case 4 (Escaped Spaces + Linkify Extension):**
  * **Configuration:** Parser configured with `parser.WithEscapedSpace()` and `extension.NewLinkifyParser()`, Renderer configured with `html.SimpleEastAsianLineBreakStrategy`.
  * **Behavior:** Verifies that escaped spaces (`\ `) do not mistakenly trigger or break the `Linkify` inline parser.
  * **Expected Output:** `<p>太郎は<strong>「こんにちわ」</strong>と言ったんです</p>`

---

### 2. `TestEastAsianLineBreaks(t *testing.T)`

Tests soft line break behaviors across lines containing CJK wide characters, CJK punctuation, and Western/Latin characters under different renderer configurations.

#### Test Cases Executed:

* **Test Case 1 (Default Behavior):**
  * **Configuration:** Default parser and renderer.
  * **Behavior:** Soft line breaks in CJK text default to standard `\n` characters in HTML output (which can render as unwanted space in browser displays).

* **Test Cases 2–5 (`html.SimpleEastAsianLineBreakStrategy` Rules):**
  * **Configuration:** `parser.WithEscapedSpace()`, `html.SimpleEastAsianLineBreakStrategy`.
  * **Behaviors:**
    * **Case 2:** Soft line breaks between East Asian wide characters are **removed** (`言った\nんです` $\rightarrow$ `言ったんです`).
    * **Case 3:** Soft line breaks between Western characters are **preserved** (`a\nb` $\rightarrow$ `a\nb`).
    * **Case 4:** Soft line breaks between a Western character and an East Asian character are **preserved** (`a\nん` $\rightarrow$ `a\nん`).
    * **Case 5:** Soft line breaks between an East Asian character and a Western character are **preserved** (`た\nb` $\rightarrow$ `た\nb`).

* **Test Case 6 (Precedence of `WithHardWraps`):**
  * **Configuration:** `parser.WithEscapedSpace()`, `html.WithHardWraps()`, `html.SimpleEastAsianLineBreakStrategy`.
  * **Behavior:** `WithHardWraps()` takes precedence over `SimpleEastAsianLineBreakStrategy`. Soft line breaks render as `<br />\n`.

* **Test Cases 7–9 (`SimpleEastAsianLineBreakStrategy` Advanced Scenarios):**
  * **Configuration:** `extension.NewLinkifyParser()`, `html.SimpleEastAsianLineBreakStrategy`.
  * **Behaviors:**
    * **Case 7:** CRLF line breaks (`\r\n`) between CJK characters are ignored/removed.
    * **Case 8:** Soft line breaks following CJK punctuation (e.g., `と、\r\n言った`) are ignored/removed.
    * **Case 9:** Verifies line break behavior in multiline text blocks containing both CJK-to-CJK transitions and CJK-to-Western transitions.

* **Test Cases 10–12 (`html.CSSText3LineBreakStrategy` Rules):**
  * **Configuration:** Standard parser, `html.CSSText3LineBreakStrategy`.
  * **Behaviors:**
    * **Case 10:** Soft line breaks between a Western character and an East Asian wide character are **removed** (`a\nん` $\rightarrow$ `aん`).
    * **Case 11:** Soft line breaks between an East Asian wide character and a Western character are **removed** (`た\nb` $\rightarrow$ `たb`).
    * **Case 12:** Soft line breaks between CJK text and Western text across multi-sentence paragraphs are completely removed without introducing explicit spaces or newlines.

---

## Line Break Strategy Matrix

Based on the test specifications within `cjk_test.go`, the behavior of line breaks across character boundaries can be summarized as:

| Boundary Type | Default Strategy | `SimpleEastAsianLineBreakStrategy` | `CSSText3LineBreakStrategy` |
| :--- | :--- | :--- | :--- |
| **CJK $\leftrightarrow$ CJK** | Retained (`\n`) | Removed | Removed |
| **CJK $\leftrightarrow$ Punctuation** | Retained (`\n`) | Removed | Removed |
| **Western $\leftrightarrow$ Western** | Retained (`\n`) | Retained (`\n`) | Retained (`\n`) |
| **Western $\rightarrow$ CJK** | Retained (`\n`) | Retained (`\n`) | Removed |
| **CJK $\rightarrow$ Western** | Retained (`\n`) | Retained (`\n`) | Removed |
| **Any (with `WithHardWraps`)**| `<br />\n` | `<br />\n` | `<br />\n` |