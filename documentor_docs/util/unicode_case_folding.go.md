# Technical Documentation: `util/unicode_case_folding.go`

## Overview

The `util/unicode_case_folding.go` file is responsible for constructing an in-memory lookup table (`unicodeCaseFoldings`) that maps individual source Unicode characters (`rune`) to their corresponding case-folded character sequences (`[]rune`).

It relies on code generation directives (`//go:generate`) to build raw mapping data, which is compiled into a generated Go file (`unicode_case_folding.gen.go`). At package initialization (`init()`), it constructs the dynamic lookup map from these generated data structures.

---

## Code Generation Directives (`//go:generate`)

The file contains two `//go:generate` directives used to generate the necessary underlying data structures before runtime:

1. **JSON Generation**:
   ```go
   //go:generate go run ../_tools unicode-case-folding-map -o ../_tools/unicode-case-folding-map.json
   ```
   Runs a tool located at `../_tools` called `unicode-case-folding-map` to output a JSON file containing Unicode case-folding mappings to `../_tools/unicode-case-folding-map.json`.

2. **Go Code Generation**:
   ```go
   //go:generate go run ../_tools emb-structs -i ../_tools/unicode-case-folding-map.json -o ./unicode_case_folding.gen.go
   ```
   Runs the `emb-structs` tool using the previously generated JSON file as input (`-i`) to produce a generated Go source file at `./unicode_case_folding.gen.go` (`-o`). This generated file provides the static variables and constants referenced during the initialization step.

---

## Data Structures and Variables

### `unicodeCaseFoldings`
```go
var unicodeCaseFoldings map[rune][]rune
```
* **Type**: `map[rune][]rune`
* **Scope**: Package-private
* **Description**: Holds the populated mapping table where the key is a source Unicode code point (`rune`), and the value is a slice of target Unicode code points (`[]rune`) representing its case-folded equivalent.

---

## Initialization Logic (`init`)

The `init()` function runs automatically when the `util` package is initialized. It builds the `unicodeCaseFoldings` map using data structures defined in the generated source file (`unicode_case_folding.gen.go`).

```go
func init() {
	unicodeCaseFoldings = make(map[rune][]rune, _unicodeCaseFoldingLength)
	cTo := 0
	for i := range _unicodeCaseFoldingLength {
		tTo := cTo + int(_unicodeCaseFoldingToIndex[i])
		to := _unicodeCaseFoldingTo[cTo:tTo]
		unicodeCaseFoldings[_unicodeCaseFoldingFrom[i]] = to
		cTo = tTo
	}
}
```

### Detailed Step-by-Step Execution

1. **Map Allocation**:
   `unicodeCaseFoldings` is allocated with an initial capacity equal to `_unicodeCaseFoldingLength` to optimize memory allocation for the map entries.

2. **Offset Tracking (`cTo`)**:
   An integer variable `cTo` is initialized to `0`. It serves as the starting index cursor for slicing the contiguous target rune array `_unicodeCaseFoldingTo`.

3. **Iterative Map Population**:
   The code loops `i` over the range `0` to `_unicodeCaseFoldingLength - 1`:
   * **Calculate Target End Index (`tTo`)**:
     `tTo` is calculated by adding the length/count stored at `_unicodeCaseFoldingToIndex[i]` (casted to `int`) to the current offset `cTo`.
   * **Slice Target Runes (`to`)**:
     A sub-slice `to` is created from `_unicodeCaseFoldingTo` spanning from index `cTo` up to `tTo` (`_unicodeCaseFoldingTo[cTo:tTo]`).
   * **Assign to Map**:
     The mapped source rune stored at `_unicodeCaseFoldingFrom[i]` is used as the key in `unicodeCaseFoldings`, assigning `to` as its target value.
   * **Advance Cursor**:
     `cTo` is updated to `tTo` for the next iteration.

---

## External/Generated Dependencies

The file relies on the following generated identifiers (expected to be present in `unicode_case_folding.gen.go`):

* `_unicodeCaseFoldingLength`: The total count of source-to-target rune mappings.
* `_unicodeCaseFoldingFrom`: An array/slice containing source key runes.
* `_unicodeCaseFoldingToIndex`: An array/slice containing the number of target runes corresponding to each entry `i`.
* `_unicodeCaseFoldingTo`: A continuous flat array/slice containing all target runes.