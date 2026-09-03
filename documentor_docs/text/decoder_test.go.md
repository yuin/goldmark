# Technical Documentation: `text/decoder_test.go`

## Overview

The `text/decoder_test.go` file contains unit tests for the decoding functionality provided by the `github.com/yuin/goldmark/v2/text` package. Its primary purpose is to verify that the `Decoder` implementation correctly unescapes and converts specific encoded text patterns (such as backslash-escaped characters and HTML entities) into their expected raw byte representations.

---

## Package and Dependencies

* **Package Name:** `text_test`
* **Imports:**
  * `bytes`: Used to compare byte slices via `bytes.Equal`.
  * `testing`: Standard Go package for writing unit tests.
  * `github.com/yuin/goldmark/v2/text`: The package under test, providing `text.NewDecoder()`.

---

## Functions

### `TestDecoder(t *testing.T)`

`TestDecoder` is a standard Go unit test function that validates the behavior of the `text.NewDecoder()` instance and its `Decode` method.

#### Test Execution Flow

1. **Input Definition:**
   A test string `s` is defined:
   ```go
   s := "foo\\:\\) &amp;ab&#x3A; &#58;"
   ```
   This string contains:
   * Backslash-escaped characters (`\:` and `\)`)
   * Named HTML entities (`&amp;`)
   * Hexadecimal HTML entities (`&#x3A;`)
   * Decimal HTML entities (`&#58;`)

2. **Decoder Initialization:**
   A decoder instance is created:
   ```go
   d := text.NewDecoder()
   ```

3. **Decoding and Assertion:**
   The test executes `d.Decode([]byte(s))` and checks whether the result matches the expected byte slice:
   ```go
   []byte("foo:) &ab: :")
   ```

4. **Error Handling:**
   If `bytes.Equal` evaluates to `false`, the test logs an error using `t.Errorf`, displaying:
   * The original input string (`s`)
   * The expected output string (`foo:) &ab: :`)
   * The actual output returned by `d.Decode([]byte(s))`

---

## Summary of Test Expectations

| Input String (`s`) | Expected Decoded Output |
| :--- | :--- |
| `"foo\\:\\) &amp;ab&#x3A; &#58;"` | `"foo:) &ab: :"` |