# Technical Documentation: `util/util_unsafe.go`

## Overview

The `util/util_unsafe.go` file provides high-performance, zero-allocation conversion utility functions between byte slices (`[]byte`) and strings (`string`). It utilizes Go's standard `unsafe` package to perform low-level pointer operations that bypass standard memory allocations and copying.

---

## Build Constraints

```go
//go:build !appengine && !js
```

This file is included in the build set only when target platforms meet both of the following conditions:
* **Not App Engine** (`!appengine`)
* **Not JavaScript/WebAssembly** (`!js`)

These constraints restrict unsafe pointer access on environments where low-level memory access might be restricted or unsupported.

---

## Package and Imports

* **Package:** `util`
* **Imports:**
  * `unsafe`: Used to access memory address data of slice headers (`unsafe.SliceData`, `unsafe.Slice`) and string headers (`unsafe.StringData`, `unsafe.String`).

---

## Functions Documentation

### `BytesToReadOnlyString`

#### Signature
```go
func BytesToReadOnlyString(b []byte) string
```

#### Purpose
Converts a byte slice (`[]byte`) into a `string` without allocating new memory for the underlying string data.

#### Parameters
* `b []byte`: The source byte slice to convert.

#### Return Value
* `string`: A string pointing directly to the underlying byte array of the input slice `b`.

#### Internal Mechanics
```go
return unsafe.String(unsafe.SliceData(b), len(b))
```
1. **`unsafe.SliceData(b)`**: Retrieves a pointer to the underlying array backing the slice `b`.
2. **`unsafe.String(ptr, len)`**: Constructs a new `string` header using the pointer `ptr` and length `len(b)`.

---

### `StringToReadOnlyBytes`

#### Signature
```go
func StringToReadOnlyBytes(s string) []byte
```

#### Purpose
Converts a `string` into a byte slice (`[]byte`) without allocating memory or copying data.

#### Parameters
* `s string`: The source string to convert.

#### Return Value
* `[]byte`: A byte slice pointing directly to the backing array of the string `s`.

#### Internal Mechanics
```go
return unsafe.Slice(unsafe.StringData(s), len(s))
```
1. **`unsafe.StringData(s)`**: Retrieves a pointer to the underlying byte array of string `s`.
2. **`unsafe.Slice(ptr, len)`**: Constructs a byte slice starting at `ptr` with both length and capacity set to `len(s)`.

---

## Key Considerations

1. **Zero-Allocation Conversions:** Standard conversions like `string(b)` or `[]byte(s)` allocate memory and copy the underlying bytes. These functions perform direct header construction, achieving $O(1)$ zero-copy conversions.
2. **Read-Only Expectations:** 
   * The string returned by `BytesToReadOnlyString` shares memory with `b`. If the original slice `b` is mutated, the string's content changes, which violates Go's string immutability assumption.
   * The byte slice returned by `StringToReadOnlyBytes` points to immutable string memory. Modifying the elements of the returned `[]byte` will result in undefined behavior or panic due to writing to read-only memory segments.