# Technical Documentation: `text/value.go`

## Overview

The `text/value.go` file defines data structures, interfaces, and utilities for working with text ranges, inline values, and block line contents within the text parsing framework (part of Goldmark). 

It primarily distinguishes between two memory models:
1. **Source-backed Data**: Data referenced via index ranges (`Index`, `Segment`) pointing into an original source byte slice (`source []byte`), minimizing allocations and copies.
2. **Owned Data**: Standalone strings stored directly within the value structs (`s string`), used when content is generated, modified, or decoupled from the original source buffer.

---

## Data Structures & Interfaces

---

### 1. Index

`Index` represents a simple position range within a source byte slice. Unlike `Segment`, `Index` does not store parsing metadata like `Padding` or `ForceNewline`.

```go
type Index struct {
    Start int
    Stop  int
}
```

#### Functions & Methods

* `NewIndex(start, stop int) Index`: Constructs a new `Index` given start and stop positions.
* `NewIndexFromSegment(seg Segment) Index`: Derives an `Index` from a `Segment` by taking its `Start` and `Stop` values, ignoring `Padding` and `ForceNewline`.
* `(i Index) IsEmpty() bool`: Returns `true` if `Start >= Stop`.

---

### 2. Value (Interface)

`Value` is an interface representing an inline text value. A `Value` can decode and output its content, check ownership status, and expose source positions.

```go
type Value interface {
    Value(source []byte) string
    Bytes(source []byte) []byte
    Str(source []byte) string
    IsOwned() bool
    IsEmpty() bool
    Index() Index
    Indices() []Index
    WriteTo(w io.Writer, source []byte) (int, error)
}
```

---

### 3. SingleLineValue

`SingleLineValue` represents an inline value restricted to a single line. It is backed either by an owned string or a single `Index` range into the source byte slice.

```go
type SingleLineValue struct {
    decoder Decoder
    s       string
    index   Index
}
```

#### Type Constraint

* `SingleLineValueInput`: Generic constraint allowing `string | []byte | Index`.

#### Constructors

* `NewSingleLineValue[T SingleLineValueInput](v T, decoder Decoder) SingleLineValue`: Converts input `v` into a `SingleLineValue` bound to the given `Decoder`.
* `NewSingleLineValueFromIndex(i Index, decoder Decoder) SingleLineValue`: Returns a position-backed `SingleLineValue`.
* `NewSingleLineValueFromSegment(seg Segment, decoder Decoder) SingleLineValue`: Returns a position-backed `SingleLineValue` derived from a `Segment`.
* `NewSingleLineValueFromString(s string, decoder Decoder) SingleLineValue`: Returns an owned string-backed `SingleLineValue`.

#### Methods

* `(v SingleLineValue) Value(source []byte) string`: Returns the decoded string representation using the bound decoder.
* `(v SingleLineValue) WriteTo(w io.Writer, source []byte) (int, error)`: Decodes and writes the value to `w`.
* `(v SingleLineValue) Bytes(source []byte) []byte`: Returns the raw slice of bytes from owned string or source range.
* `(v SingleLineValue) Str(source []byte) string`: Returns string representation of raw bytes without decoding.
* `(v SingleLineValue) IsOwned() bool`: Returns `true` if `index.Start == 0` and `index.Stop == 0`.
* `(v SingleLineValue) IsEmpty() bool`: Returns `true` if owned string is `""` or index range `Start >= Stop`.
* `(v SingleLineValue) Index() Index`: Returns the source position `Index`.
* `(v SingleLineValue) Indices() []Index`: Returns a single-element slice containing `v.index`.
* `(v SingleLineValue) WithStop(stop int) SingleLineValue`: Returns a copy with updated `Stop` position. **Panics** if the value is owned.

---

### 4. MultiLineValue

`MultiLineValue` represents an inline value that can span multiple lines. It is backed by either an owned string, a single source `Index` (optimizing the single-range case using an array `[1]Index`), or a slice of `Index` ranges (`[]Index`).

```go
type MultiLineValue struct {
    decoder Decoder
    s       string
    index   [1]Index
    indices []Index
}
```

#### Type Constraint

* `MultiLineValueInput`: Generic constraint allowing `string | []byte | Index | []Index`.

#### Constructors

* `NewMultiLineValue[T MultiLineValueInput](v T, decoder Decoder) MultiLineValue`: Converts `v` into a `MultiLineValue`.
* `NewMultiLineValueFromIndex(i Index, decoder Decoder) MultiLineValue`: Constructs a single-index `MultiLineValue`.
* `NewMultiLineValueFromIndices(indices []Index, decoder Decoder) MultiLineValue`: Constructs a multi-index `MultiLineValue`. If `len(indices) == 1`, optimizes storage using `[1]Index`.
* `NewMultiLineValueFromString(s string, decoder Decoder) MultiLineValue`: Constructs an owned string `MultiLineValue`.

#### Methods

* `(v MultiLineValue) Value(source []byte) string`: Decodes and concatenates string representation across ranges.
* `(v MultiLineValue) WriteTo(w io.Writer, source []byte) (int, error)`: Writes decoded content to `w`.
* `(v MultiLineValue) Bytes(source []byte) []byte`: Returns combined raw bytes. If multiple indices exist, chunks are concatenated into a newly allocated buffer.
* `(v MultiLineValue) Str(source []byte) string`: Converts raw bytes representation to string.
* `(v MultiLineValue) IsOwned() bool`: Returns `true` if no indices are set (`indices == nil` and `index[0]` is zero-valued).
* `(v MultiLineValue) IsEmpty() bool`: Returns `true` if owned string is empty or all stored index ranges are empty (`Start >= Stop`).
* `(v MultiLineValue) Index() Index`: Returns the primary or first `Index`.
* `(v MultiLineValue) Indices() []Index`: Returns all `Index` objects as a slice.

---

### 5. ValueBuilder

`ValueBuilder` is a helper struct for constructing `Value` objects (`SingleLineValue` or `MultiLineValue`) dynamically.

```go
type ValueBuilder struct {
    decoder Decoder
    s       string
    index   Index
    indices []Index
}
```

#### Methods

* `(b *ValueBuilder) AddIndex(i Index) *ValueBuilder`: Adds an `Index`. Automatically transitions from single-index storage (`b.index`) to slice-backed storage (`b.indices`) upon adding a second index.
* `(b *ValueBuilder) AddSegment(seg Segment) *ValueBuilder`: Converts `seg` to an `Index` and adds it.
* `(b *ValueBuilder) Decoder(d Decoder) *ValueBuilder`: Sets the bound `Decoder`.
* `(b *ValueBuilder) OwnedString(s string) *ValueBuilder`: Sets an owned string.
* `(b *ValueBuilder) OwnedBytes(bts []byte) *ValueBuilder`: Sets an owned string converted from byte slice.
* `(b *ValueBuilder) BuildSingleLine() SingleLineValue`: Constructs a `SingleLineValue` using the accumulated string or first index.
* `(b *ValueBuilder) BuildMultiLine() MultiLineValue`: Constructs a `MultiLineValue`.
* `(b *ValueBuilder) Build() Value`: Dynamically constructs either a `SingleLineValue` or `MultiLineValue`:
  * If an owned string is set, returns `SingleLineValue` if `\n` is absent, or `MultiLineValue` if present.
  * If source ranges are set, returns `SingleLineValue` if only one index exists, or `MultiLineValue` if multiple indices exist.

---

### 6. Lines

`Lines` holds raw content for block nodes (e.g., Code Blocks, HTML Blocks). Content is un-decoded ("raw") and stored either as an owned string or a slice of `Segment`s.

```go
type Lines struct {
    s    string
    segs []Segment
}
```

#### Type Constraint

* `LinesInput`: Generic constraint allowing `string | []byte | []Segment | Lines`.

#### Constructors

* `NewLines[T LinesInput](v T) Lines`: Converts generic input `v` into a `Lines` instance.
* `NewLinesFromSegments(segs []Segment) Lines`: Returns a segment-backed `Lines` object.
* `NewLinesFromString(s string) Lines`: Returns an owned string `Lines` object.

#### Methods

* `(l Lines) Segments() []Segment`: Returns internal slice of segments.
* `(l *Lines) AppendSegment(seg Segment)`: Appends a `Segment` to internal storage.
* `(l Lines) Bytes(source []byte) []byte`: Returns raw concatenated byte slice.
  * *Optimization*: If all segments are contiguous in source without `Padding` or `ForceNewline`, it slices `source` directly to avoid extra memory allocations.
* `(l Lines) Str(source []byte) string`: Returns `Bytes(source)` converted to string.
* `(l Lines) WriteTo(w io.Writer, source []byte) (int, error)`: Writes segment contents directly to `w`. Appends `\n` if a segment has `ForceNewline == true` and does not already end with `\n`.
* `(l Lines) IsOwned() bool`: Returns `true` if `segs == nil`.

---

### 7. Segment

`Segment` holds detailed positioning and metadata regarding a portion of the source content.

```go
type Segment struct {
    Start        int
    Stop         int
    Padding      int
    ForceNewline bool
}
```

#### Constructors

* `NewSegment(start, stop int) Segment`: Creates a segment with `Padding = 0`.
* `NewSegmentPadding(start, stop, padding int) Segment`: Creates a segment with specified padding.

#### Methods

* `(t Segment) Bytes(source []byte) []byte`: Extracts segment content from `source`. Prepends spaces for `Padding` if non-zero. Appends `\n` if `ForceNewline` is set and content doesn't end with `\n`.
* `(t Segment) Str(source []byte) string`: Returns `Bytes(source)` as a string.
* `(t Segment) Len() int`: Returns `Stop - Start + Padding`.
* `(t Segment) Between(other Segment) Segment`: Returns a new segment spanning between `t` and `other`. **Panics** if `t.Stop != other.Stop`.
* `(t Segment) IsEmpty() bool`: Returns `true` if `Start >= Stop` and `Padding == 0`.
* `(t Segment) TrimRightSpace(source []byte) Segment`: Returns a new `Segment` with trailing whitespace removed.
* `(t Segment) TrimLeftSpace(source []byte) Segment`: Returns a new `Segment` with leading whitespace removed.
* `(t Segment) TrimLeftSpaceWidth(width int, source []byte) Segment`: Consumes up to `width` spaces/tabs from left side (accounting for `Padding`). Tabs count as 4 spaces.
* `(t Segment) WithStart(v int) Segment`: Returns a copy with updated `Start`. Resets `Padding` and `ForceNewline`.
* `(t Segment) WithStop(v int) Segment`: Returns a copy with updated `Stop`. Resets `ForceNewline` to `false`.

---

## Utility Functions Used

This file makes extensive use of `github.com/yuin/goldmark/v2/util` zero-allocation conversion helpers:
* `util.BytesToReadOnlyString([]byte) string`
* `util.StringToReadOnlyBytes(string) []byte`
* `util.TrimRightSpaceLength([]byte) int`
* `util.TrimLeftSpaceLength([]byte) int`