# Technical Documentation: `cmark_benchmark.c`

## Overview

The `cmark_benchmark.c` source file is a C benchmark utility designed to measure the execution performance of converting Markdown content to HTML using the `cmark` library (`cmark_markdown_to_html`). It reads a Markdown file into memory, runs the conversion in a loop for a specified number of iterations, measures the elapsed execution time per iteration, and prints the average runtime.

---

## Preprocessor Directives & Dependencies

### Included Headers
* `<stdio.h>`: Provides standard input/output functions (file handling via `fopen`, `fseek`, `ftell`, `fread`, `fclose`, and formatted printing via `printf`, `fprintf`).
* `<stdlib.h>`: Provides memory allocation (`malloc`, `free`), process control (`exit`), and string conversion (`atoi`).
* Conditional Platform Headers:
  * **Windows (`#ifdef WIN32`)**: Includes `<windows.h>` for high-resolution timing functions.
  * **Unix/POSIX (`#else`)**: Includes `<sys/time.h>` and `<sys/resource.h>` for microsecond-precision timing functions.
* `"cmark.h"`: External library header providing the `cmark_markdown_to_html` function and the `CMARK_OPT_UNSAFE` option flag.

---

## Command-Line Arguments

The executable accepts up to two optional positional command-line arguments:

```bash
./cmark_benchmark [iterations] [markdown_file]
```

1. `argv[1]` (**Iterations**): The number of conversion iterations to execute. 
   * **Type**: Integer (`int`).
   * **Default**: `50` (if `argc <= 1`).
2. `argv[2]` (**Markdown File Path**): The path to the Markdown file used for benchmarking.
   * **Type**: String (`char *`).
   * **Default**: `"_data.md"` (if `argc <= 2`).

---

## Function Documentation

### `double get_time()`

Calculates and returns the current time in seconds as a high-precision `double`.

* **Platform Implementations**:
  * **Windows (`WIN32` defined)**:
    * Uses `QueryPerformanceCounter` to retrieve the current tick count.
    * Uses `QueryPerformanceFrequency` to retrieve the counter frequency.
    * Returns the elapsed seconds as `(double)t.QuadPart / (double)f.QuadPart`.
  * **POSIX/Unix (`WIN32` not defined)**:
    * Uses `gettimeofday(&t, &tzp)` to populate a `timeval` structure.
    * Computes and returns the total seconds as `t.tv_sec + t.tv_usec * 1e-6`.

---

### `int main(int argc, char **argv)`

The main execution entry point for the benchmark program.

#### Local Variables
| Variable | Type | Description |
| :--- | :--- | :--- |
| `markdown_file` | `char *` | Target file path for the input Markdown document. |
| `fp` | `FILE *` | File handle for reading the input file. |
| `size` | `size_t` | Byte size of the input Markdown file. |
| `buf` | `char *` | Heap-allocated buffer containing the file content. |
| `html` | `char *` | Pointer to the output HTML string returned by `cmark_markdown_to_html`. |
| `start` | `double` | Timestamp recorded immediately prior to a conversion iteration. |
| `sum` | `double` | Accumulator for the total execution time across all iterations. |
| `i` | `int` | Loop counter for iterations. |
| `n` | `int` | Total number of iterations to perform. |

---

## Step-by-Step Execution Flow

1. **Parse Input Arguments**:
   * Resolves the number of iterations `n` from `argv[1]` (defaults to `50`).
   * Resolves the file path `markdown_file` from `argv[2]` (defaults to `"_data.md"`).

2. **File Reading & Memory Allocation**:
   * Opens `markdown_file` in read mode (`"r"`). Exits with code `1` if opening fails.
   * Uses `fseek(fp, 0, SEEK_END)` and `ftell(fp)` to calculate the total size of the file in bytes. Exits on seek/tell failures.
   * Rewinds the file pointer back to the start using `fseek(fp, 0, SEEK_SET)`.
   * Dynamically allocates memory (`buf`) equal to `sizeof(char) * size`. Exits if allocation fails.
   * Reads the entire content of the file into `buf` using `fread`. Exits if fewer bytes than expected are read.
   * Closes the file handle (`fclose(fp)`).

3. **Benchmarking Loop**:
   * Iterates `n` times (`0` to `n - 1`).
   * Captures the starting timestamp via `get_time()`.
   * Invokes `cmark_markdown_to_html(buf, size, CMARK_OPT_UNSAFE)`.
   * Immediately frees the dynamically allocated `html` string returned by `cmark`.
   * Measures the time difference (`get_time() - start`) and adds it to `sum`.

4. **Results Output**:
   * Prints the benchmark header and metrics to standard output (`stdout`):
     * Benchmark identifier (`----------- cmark -----------`).
     * Target file name (`file: <filename>`).
     * Iteration count (`iteration: <n>`).
     * Computed average duration in seconds (`average: <sum / n>`).

5. **Cleanup**:
   * Frees the input file buffer `buf`.
   * Returns `0` to signal successful termination.

---

## Output Format

When executed, the program outputs formatted text to `stdout`. 

### Example Output

```text
----------- cmark -----------
file: _data.md
iteration: 50
average: 0.0001234567 sec
```