# Technical Documentation: `extension/linkify_test.go`

## Overview

The `extension/linkify_test.go` file contains unit tests for the `Linkify` extension in the Goldmark Markdown processing library. It verifies the functionality of auto-linking plain text URLs, WWW domain names, and email addresses into HTML `<a>` tags. It also tests configurable extension options such as restricted allowed protocols and custom regular expressions for matching URLs, WWW hosts, and email addresses.

---

## Code Structure & Dependencies

### Package
* **`package extension`**: Belongs to the extension module of the Goldmark library.

### External Dependencies
* **`regexp`**: Go standard library package used to compile custom regular expressions for testing URL, WWW, and email pattern overrides.
* **`testing`**: Standard Go package for unit testing.
* **`github.com/yuin/goldmark/v2/parser`**: Core Goldmark package providing parser customization functions (`parser.New`, `parser.WithExtensions`).
* **`github.com/yuin/goldmark/v2/renderer/html`**: Core Goldmark renderer options (`html.New`, `html.WithUnsafe`, `html.WithXHTML`).
* **`github.com/yuin/goldmark/v2/testutil`**: Utility package for testing Goldmark parsers and renderers.

---

## Detailed Test Function Reference

### 1. `TestLinkify(t *testing.T)`

#### Purpose
Tests the default behavior of `NewLinkifyParser()` against test cases stored in an external test file.

#### Mechanics
1. **Markdown Converter Setup**:
   * Initializes a conversion function using `testutil.NewMarkdownToStringFunc`.
   * **Parser**: Configured with `NewLinkifyParser()` default extension.
   * **Renderer**: HTML renderer initialized with `html.WithUnsafe()` to allow rendering raw HTML/links.
2. **Execution**:
   * Calls `testutil.DoTestCaseFile(...)` targeting the test suite defined in `_test/linkify.txt`.
   * Accepts command-line argument flags via `testutil.ParseCliCaseArg()...`.

---

### 2. `TestLinkifyWithAllowedProtocols(t *testing.T)`

#### Purpose
Verifies that `NewLinkifyParser()` can be restricted to specific URL schemes using `WithAllowedProtocols` and a custom URL regular expression via `WithURLRegexp`.

#### Configuration Options Tested
* `WithAllowedProtocols([]string{"ssh:"})`: Limits auto-linking to URLs that begin with the `ssh:` protocol scheme.
* `WithURLRegexp(regexp.MustCompile(`\w+://[^\s]+`))`: Supplies a regular expression matching scheme-based URLs.

#### Test Execution & Verification
* **Renderer Options**: `html.WithXHTML()` and `html.WithUnsafe()`.
* **Input Markdown**: `hoge ssh://user@hoge.com. http://example.com/`
* **Expected Output**: `<p>hoge <a href="ssh://user@hoge.com">ssh://user@hoge.com</a>. http://example.com/</p>`
* **Behavior Observed**:
  * `ssh://user@hoge.com` is matched and converted into an HTML anchor tag.
  * `http://example.com/` is ignored (left as plain text) because `http:` is not listed in `WithAllowedProtocols`.

---

### 3. `TestLinkifyWithWWWRegexp(t *testing.T)`

#### Purpose
Verifies that custom regular expressions for matching domain names beginning with `www.` can be configured using `WithWWWRegexp`.

#### Configuration Options Tested
* `WithWWWRegexp(regexp.MustCompile(`www\.example\.com`))`: Restricts WWW auto-linking exclusively to the explicit string `www.example.com`.

#### Test Execution & Verification
* **Renderer Options**: `html.WithXHTML()` and `html.WithUnsafe()`.
* **Input Markdown**: `www.google.com www.example.com`
* **Expected Output**: `<p>www.google.com <a href="http://www.example.com">www.example.com</a></p>`
* **Behavior Observed**:
  * `www.google.com` is ignored and remains unlinked because it does not match the custom regular expression filter.
  * `www.example.com` matches the regular expression and is automatically converted into `<a href="http://www.example.com">www.example.com</a>`.

---

### 4. `TestLinkifyWithEmailRegexp(t *testing.T)`

#### Purpose
Verifies that custom regular expressions for matching email addresses can be set using `WithEmailRegexp`.

#### Configuration Options Tested
* `WithEmailRegexp(regexp.MustCompile(`user@example\.com`))`: Restricts email auto-linking exclusively to `user@example.com`.

#### Test Execution & Verification
* **Renderer Options**: `html.WithXHTML()` and `html.WithUnsafe()`.
* **Input Markdown**: `hoge@example.com user@example.com`
* **Expected Output**: `<p>hoge@example.com <a href="mailto:user@example.com">user@example.com</a></p>`
* **Behavior Observed**:
  * `hoge@example.com` remains plain text because it fails to match the strict regex.
  * `user@example.com` matches the regular expression and is converted into a mailto link (`<a href="mailto:user@example.com">user@example.com</a>`).