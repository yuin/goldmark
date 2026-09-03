# Technical Documentation: `util/util_safe.go`

## Overview

The `util/util_safe.go` file provides safe, standard Go conversions between byte slices (`[]byte`) and strings (`string`). This file is compiled specifically for restricted target environments, such as Google App Engine and JavaScript (WebAssembly / GopherJS), where memory operations using Go's `unsafe` package are either prohibited or unsupported.

---

## Build Constraints

```go
//go:build appengine || js
```

This file uses the `//go:build` directive with the condition `appengine || js`. 

* **`appengine`**: Targets execution on Google App Engine.
* **`js`**: Targets execution in JavaScript environments (e.g., WebAssembly or GopherJS).

When compiling for these build targets, Go selects this file to perform byte-to-string and string-to-byte conversions using standard Go type conversions, avoiding low-level unsafe memory manipulation.

---

## Package

```go
package util
```

The file belongs to the `util` package, providing general helper utilities for the application.

---

## Functions

### `BytesToReadOnlyString`

Converts a byte slice (`[]byte`) into a Go `string`.

#### Function Signature

```go
func BytesToReadOnlyString(b []byte) string
```

#### Parameters

* `b []byte`: The byte slice to be converted.

#### Return Value

* `string`: The string representation of the provided byte slice.

#### Implementation & Behavior

```go
func BytesToReadOnlyString(b []byte) string {
    return string(b)
}
```

* Performs a standard Go type conversion `string(b)`.
* Memory allocation occurs as Go copies the byte slice contents into an immutable string.
* Safe for use in environments where `unsafe` operations are restricted.

---

### `StringToReadOnlyBytes`

Converts a Go `string` into a byte slice (`[]byte`).

#### Function Signature

```go
func StringToReadOnlyBytes(s string) []byte
```

#### Parameters

* `s string`: The input string to be converted.

#### Return Value

* `[]byte`: A byte slice containing the raw bytes of the string.

#### Implementation & Behavior

```go
func StringToReadOnlyBytes(s string) []byte {
    return []byte(s)
}
```

* Performs a standard Go type conversion `[]byte(s)`.
* Memory allocation occurs as Go allocates a new underlying byte array and copies the string bytes into it.
* Ensures safe runtime behavior without direct memory pointer manipulation.