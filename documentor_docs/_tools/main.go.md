# Technical Documentation: `_tools/main.go`

## Overview

The `_tools/main.go` file serves as the main entry point for a multi-purpose Command-Line Interface (CLI) developer utility tool. Its primary responsibility is parsing command-line arguments and routing execution to subcommands based on the first positional argument provided.

---

## Architecture & Key Components

The file contains two functions:
1. `main()`: The program entry point that parses command-line arguments and dispatches control to specified subcommand handlers.
2. `usage(u func(), err error)`: A utility function designed to execute a usage function, print error messages (if present), and terminate the process with an exit status code of `1`.

### Dependencies & Subcommands
The `main` function references three external subcommand handler functions (expected to be defined elsewhere within the `main` package):
- `ossFuzzCorpusSubCommand(args []string)`
- `unicodeCaseFoldingMapSubCommand(args []string)`
- `embStructsSubCommand(args []string)`

---

## Code Breakdown and Logic

### 1. `main()` Function

```go
func main()
```

#### Process Flow:
1. **Argument Extraction**:
   - Initializes `cmd` to `"-h"` by default.
   - Initializes `args` as an empty slice of strings (`[]string`).
   - If `os.Args` has more than 1 element, `cmd` is set to `os.Args[1]`.
   - If `os.Args` has more than 2 elements, `args` receives the remaining command-line arguments (`os.Args[2:]`).

2. **Command Dispatching (`switch cmd`)**:
   - `"oss-fuzz-corpus"`: Invokes `ossFuzzCorpusSubCommand(args)`.
   - `"unicode-case-folding-map"`: Invokes `unicodeCaseFoldingMapSubCommand(args)`.
   - `"emb-structs"`: Invokes `embStructsSubCommand(args)`.
   - `"-h"` or `default`:
     - Prints usage information to standard error (`os.Stderr`).
     - Calls `os.Exit(1)` to end execution.

---

### 2. `usage()` Function

```go
func usage(u func(), err error)
```

#### Parameters:
- `u func()`: A callback function that outputs subcommand-specific usage/help text.
- `err error`: An optional error object. If not `nil`, its value is printed to standard error (`os.Stderr`).

#### Execution Logic:
1. Calls the provided callback function `u()`.
2. Checks if `err != nil`. If true, outputs the error message to `os.Stderr`.
3. Calls `os.Exit(1)` to terminate the application with an error exit code.

---

## Command-Line Interface Usage

### Command Structure
```bash
_tools <subcommand> [options]
```

### Supported Subcommands

| Subcommand | Handler Function |
| :--- | :--- |
| `oss-fuzz-corpus` | `ossFuzzCorpusSubCommand(args)` |
| `unicode-case-folding-map` | `unicodeCaseFoldingMapSubCommand(args)` |
| `emb-structs` | `embStructsSubCommand(args)` |
| `-h` (or unrecognized) | Displays main usage instructions and exits with code `1`. |

### Default Usage Output
If no subcommand is passed, `-h` is passed, or an unknown command is supplied, the program prints the following text to `os.Stderr` and exits with code `1`:

```text
Usage: _tools <subcommand> [options]
subcommands:
  oss-fuzz-corpus
  unicode-case-folding-map
  emb-structs
```