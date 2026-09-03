# Technical Documentation: `goldmark_benchmark.go`

## Overview

The `goldmark_benchmark.go` file is a Go command-line tool designed to benchmark the performance of the **Goldmark** Markdown parser and renderer (`github.com/yuin/goldmark/v2`). It measures the execution time required to parse a specified Markdown file and render it to HTML across a set number of iterations, then outputs the average execution time. It also optionally generates a CPU profile.

---

## Command-Line Arguments

The executable accepts up to three positional command-line arguments:

| Position | Argument | Type | Default Value | Description |
| :--- | :--- | :--- | :--- | :--- |
| `os.Args[1]` | Iterations (`n`) | Integer | `50` | The number of times to run the parse and render loop. |
| `os.Args[2]` | Source File (`file`) | String | `"_data.md"` | The path to the Markdown source file to benchmark. |
| `os.Args[3]` | CPU Profile Output | String | *None* | File path to write a CPU profile using `runtime/pprof`. If omitted, profiling is skipped. |

---

## Key Components

### 1. Argument Parsing & Profiling Setup
* Checks `os.Args` length to extract iteration count, input file path, and optional pprof output path.
* If a third positional argument is supplied:
  * Creates the specified profile destination file using `os.Create`.
  * Starts CPU profiling via `pprof.StartCPUProfile(f)`.
  * Registers deferred cleanup functions to stop profiling (`pprof.StopCPUProfile()`) and close the file (`f.Close()`).

### 2. Source File Ingestion
* Reads the Markdown source file into memory as a byte slice (`source`) using `ioutil.ReadFile(file)`.
* Triggers a `panic` if the file cannot be read.

### 3. Goldmark Configuration
* **Parser**: Instantiated via `parser.New()`.
* **Renderer**: Instantiated via `html.New()` with the following options:
  * `html.WithXHTML()`: Configures renderer to generate XHTML-compliant HTML tags.
  * `html.WithUnsafe()`: Enables rendering of raw/unsafe HTML within the Markdown source.

### 4. Benchmark Execution Loop
* Executes a loop `n` times to perform the benchmark:
  1. Records start time using `time.Now()`.
  2. Resets the target output byte buffer (`out.Reset()`).
  3. Parses the source into an AST document (`p.Parse(source)`).
  4. Renders the AST to HTML (`r.Render(&out, source, doc)`).
  5. Accumulates elapsed execution duration into `sum`.

### 5. Benchmark Results Output
Prints the following benchmark metrics to standard output (`fmt.Printf`):
* Header divider (`------- goldmark -------`).
* File path used.
* Iteration count (`n`).
* Average execution time per iteration in seconds (calculated as `(sum / n)` converted to float seconds formatted to 10 decimal places).

---

## Execution Flow

```
+-------------------------------------------------------+
| Parse command-line arguments (n, file, profile path)  |
+-------------------------------------------------------+
                           |
                           v
+-------------------------------------------------------+
| Optional: Start CPU Profile if 3rd argument provided |
+-------------------------------------------------------+
                           |
                           v
+-------------------------------------------------------+
| Read target Markdown file into byte slice             |
+-------------------------------------------------------+
                           |
                           v
+-------------------------------------------------------+
| Initialize Goldmark parser and HTML renderer          |
+-------------------------------------------------------+
                           |
                           v
+-------------------------------------------------------+
| Loop n times:                                         |
|  1. Reset buffer                                      |
|  2. Measure start time                                |
|  3. Parse Markdown -> AST                             |
|  4. Render AST -> HTML                                |
|  5. Accumulate duration                               |
+-------------------------------------------------------+
                           |
                           v
+-------------------------------------------------------+
| Print average time per iteration to stdout            |
+-------------------------------------------------------+
```

---

## Example Usage

### Default Run
Runs 50 iterations against `_data.md`:
```bash
go run goldmark_benchmark.go
```

### Custom Iterations and File
Runs 100 iterations against `document.md`:
```bash
go run goldmark_benchmark.go 100 document.md
```

### Enabling CPU Profiling
Runs 100 iterations against `document.md` and generates a CPU profile named `cpu.pprof`:
```bash
go run goldmark_benchmark.go 100 document.md cpu.pprof
```