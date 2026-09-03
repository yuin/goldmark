# Technical Documentation: `ast/ast_test.go`

## Overview

The `ast/ast_test.go` file contains unit tests for the Abstract Syntax Tree (AST) package within the `github.com/yuin/goldmark` library. Specifically, this file tests the `Walk` function, which provides depth-first traversal of AST node hierarchies with support for controlling traversal flow using status flags (`WalkStatus`).

---

## Package and Dependencies

### Package
* **`ast_test`**: Defined as an external test package (`ast_test`), enabling black-box testing of the `ast` package.

### Dependencies
* **`reflect`**: Standard Go library package used for deep equality comparisons (`reflect.DeepEqual`) on slices of `NodeKind`.
* **`testing`**: Standard Go unit testing framework.
* **`. "github.com/yuin/goldmark/v2/ast"`**: Dot-imported `ast` package, allowing direct access to exported types, constructors, and constants (e.g., `Node`, `NodeKind`, `Walk`, `WalkStatus`, `KindDocument`, etc.).
* **`textm "github.com/yuin/goldmark/v2/text"`**: Imported with alias `textm`, providing text-handling structures such as `textm.SingleLineValue{}`.

---

## Test Implementation Details

### `TestWalk(t *testing.T)`

`TestWalk` is a table-driven test that validates the tree traversal logic of the `Walk` function under different traversal control actions.

#### Data Structures

##### Test Case Struct
The unit test uses an anonymous struct slice `tests` with the following fields:

| Field | Type | Description |
| :--- | :--- | :--- |
| `name` | `string` | The descriptive name of the test case. |
| `node` | `Node` | The root node of the constructed test AST hierarchy. |
| `want` | `[]NodeKind` | The expected slice of `NodeKind` values visited during the node-entering phase of traversal. |
| `action` | `map[NodeKind]WalkStatus` | A mapping of `NodeKind` to a specific `WalkStatus` to control traversal behavior when encountering that node type. |

#### Helper Construction Function (`N`)
The test tree structure is constructed using the `N()` helper function (provided by the `ast` package):
* `N(NewDocument(), N(NewHeading(1, HeadingKindATX), ""), NewLink(textm.SingleLineValue{}))`

This constructs an AST with the following structure:
1. **Document Node** (`KindDocument`)
   * **Heading Node** (`KindHeading`)
     * **Text Node** (`KindText`)
   * **Link Node** (`KindLink`)

---

## Test Cases

The file defines three test scenarios to verify AST traversal behaviors:

### 1. "visits all in depth first order"
* **Goal**: Verify that default traversal visits every node in the tree in depth-first order when no special action is returned.
* **Action Map**: Empty (`map[NodeKind]WalkStatus{}`)
* **Expected Sequence**: `[KindDocument, KindHeading, KindText, KindLink]`

### 2. "stops after heading"
* **Goal**: Verify that returning `WalkStop` halts the entire tree traversal immediately.
* **Action Map**: `{KindHeading: WalkStop}`
* **Expected Sequence**: `[KindDocument, KindHeading]` (traversal terminates when visiting the heading).

### 3. "skip children"
* **Goal**: Verify that returning `WalkSkipChildren` bypasses the child nodes of the target node while continuing to sibling nodes.
* **Action Map**: `{KindHeading: WalkSkipChildren}`
* **Expected Sequence**: `[KindDocument, KindHeading, KindLink]` (the child `KindText` node of `KindHeading` is skipped).

---

## Execution Logic

For each test case in `tests`:

1. **Visitor Function (`collectKinds`)**:
   * A closure `collectKinds(n Node, entering bool)` is passed to `Walk`.
   * **Node Entrance**: If `entering` is `true`, `n.Kind()` is appended to the `kinds` slice.
   * **Action Trigger**: Checks if `tt.action` defines a explicit `WalkStatus` for `n.Kind()`. If present, that status is returned.
   * **Default Behavior**: If no specific action is configured, returns `WalkContinue, nil`.

2. **Assertions**:
   * **Error Check**: Asserts that `Walk(tt.node, collectKinds)` completes without returning an error.
   * **Sequence Verification**: Uses `reflect.DeepEqual(kinds, tt.want)` to assert that the visited node types matched the expected sequence precisely in order.

---

## Walk Statuses Tested

| Status | Behavior Tested |
| :--- | :--- |
| `WalkContinue` | Standard depth-first continuation through all nodes. |
| `WalkStop` | Halts further traversal of the tree immediately. |
| `WalkSkipChildren` | Bypasses child nodes of the current node and proceeds to the next sibling. |