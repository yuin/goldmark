# Technical Documentation: `util/html5entities.go`

## Overview

The `util/html5entities.go` file provides functionality for looking up HTML5 entities by their name. It utilizes lazy initialization backed by code generation to construct an in-memory lookup map of HTML5 entity names to their corresponding character representations.

---

## Code Generation Directive

```go
//go:generate go run ../_tools emb-structs -i ../_tools/html5entities.json -o ./html5entities.gen.go
```

The file includes a `go:generate` directive that executes a tool named `emb-structs` located in `../_tools`. This tool reads an input file (`../_tools/html5entities.json`) and generates a Go source file (`./html5entities.gen.go`). 

The generated file provides the backing data structures used by `buildHTML5Entities()`, including:
* `_html5entitiesLength`
* `_html5entitiesNameIndex`
* `_html5entitiesCharactersIndex`
* `_html5entitiesName`
* `_html5entitiesCharacters`

---

## Data Structures

### `HTML5Entity`

An exported struct representing an HTML5 entity entity.

```go
type HTML5Entity struct {
	Name       string
	Characters []byte
}
```

#### Fields
* `Name` (`string`): The textual identifier/name of the HTML5 entity.
* `Characters` (`[]byte`): The byte slice containing the character sequence corresponding to the entity.

---

## Package-Level Variables

* **`_html5entitiesOnce`** (`sync.Once`): Ensures that the HTML5 entities map (`_html5entitiesMap`) and backing storage are initialized lazily and exactly once across all goroutines.
* **`_html5entitiesMap`** (`map[string]*HTML5Entity`): A private, package-level lookup map that associates an entity name string with a pointer to its `HTML5Entity` instance.

---

## Internal Logic & Functions

### `buildHTML5Entities()`

```go
func buildHTML5Entities()
```

An unexported function responsible for initializing `_html5entitiesMap` and instantiating the slice of `HTML5Entity` objects.

#### How It Works:
1. **Concurrency Control**: Wraps the entire construction logic inside `_html5entitiesOnce.Do(...)` to guarantee thread-safe, single-execution initialization.
2. **Allocation**: Allocates a slice of `HTML5Entity` structs (`entities`) and the lookup map (`_html5entitiesMap`), both sized to `_html5entitiesLength`.
3. **Reconstruction Loop**: Iterates through each index `i` up to `_html5entitiesLength`:
   * Computes offset windows for the entity name and characters using cumulative indices from `_html5entitiesNameIndex` and `_html5entitiesCharactersIndex`.
   * Slices the packed string/byte data (`_html5entitiesName` and `_html5entitiesCharacters`).
   * Populates the fields (`Name` and `Characters`) of the `HTML5Entity` at `entities[i]`.
   * Inserts the pointer to the `HTML5Entity` into `_html5entitiesMap` using the entity `Name` as the key.

---

### `LookUpHTML5EntityByName(name string)`

```go
func LookUpHTML5EntityByName(name string) (*HTML5Entity, bool)
```

An exported function used to query the HTML5 entity registry by entity name.

#### Parameters
* `name` (`string`): The name of the HTML5 entity to look up.

#### Return Values
* `*HTML5Entity`: A pointer to the matching `HTML5Entity` if found; `nil` otherwise.
* `bool`: `true` if the entity exists in the map, `false` otherwise.

#### How It Works:
1. Calls `buildHTML5Entities()`. If it is the first call, the map will be constructed; subsequent calls bypass initialization via `sync.Once`.
2. Queries `_html5entitiesMap` with the provided `name`.
3. Returns the entity pointer and the boolean result indicating presence in the map.