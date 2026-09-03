# Documentation: `_tools/gen-oss-fuzz-corpus.go`

## Overview

The `_tools/gen-oss-fuzz-corpus.go` file provides a utility tool designed to generate a initial seed corpus file (`.zip`) for OSS-Fuzz. It reads test cases containing Markdown content from a JSON file (`_test/spec.json`) and packages each Markdown snippet into an individual file inside a target `.zip` archive.

---

## Data Structures

### `TestCase`
A struct used to map individual JSON objects from the input specification file (`_test/spec.json`).

```go
type TestCase struct {
	Example  int    `json:"example"`
	Markdown string `json:"markdown"`
}
```

* **`Example`** (`int`): Represents the identifier/number for the test case example (JSON key: `"example"`).
* **`Markdown`** (`string`): Contains the raw Markdown input string for the test case (JSON key: `"markdown"`).

---

## Function Documentation

### `ossFuzzCorpusSubCommand(args []string)`

The entry point for executing the corpus generation logic.

#### Parameters
* **`args`** (`[]string`): Command-line arguments passed to the subcommand. 
  * `args[0]` is expected to be the path of the output `.zip` file (e.g., `corpus.zip`).

#### Workflow Details

1. **Argument Validation**:
   * Inspects `args[0]` to check if it ends with the `.zip` extension using `strings.HasSuffix`.
   * If `args[0]` does not end in `.zip`, the program terminates with a fatal log message displaying the required usage format.

2. **Zip File Creation**:
   * Creates the output file on disk specified by `args[0]` using `os.Create`.
   * Instantiates a new zip archive writer (`zip.NewWriter`) wrapping the created file.
   * Checks for creation errors and logs a fatal error if file creation failed.

3. **Reading and Parsing Input Specification**:
   * Opens and reads the contents of the hardcoded file path `_test/spec.json` using `ioutil.ReadFile`.
   * Unmarshals the JSON byte slice into a slice of `TestCase` structs (`[]TestCase`).
   * If reading or unmarshaling fails, the execution panics.

4. **Populating the Zip Archive**:
   * Iterates through each `TestCase` in `testCases`:
     * Generates a filename for the archive entry using the format `example-<Example>` (e.g., `example-1`).
     * Creates a new file entry inside the zip archive using `zip_writer.Create`.
     * Writes the byte representation of the `c.Markdown` string into the created zip entry file.
     * Logs a fatal error if entry creation or writing fails.

5. **Closing Resources**:
   * Closes the zip writer using `zip_writer.Close()`.
   * Closes the underlying zip destination file handle via `zip_file.Close()`.

---

## Error Handling

The program handles operational errors using `log.Fatalln`, `log.Fatal`, `log.Fatalf`, or `panic`:

| Action / Trigger | Error Handling Mechanism | Result |
| :--- | :--- | :--- |
| Argument missing `.zip` suffix | `log.Fatalln` | Logs expected usage and exits execution. |
| Output `.zip` file creation failure | `log.Fatalln` | Logs failure message and exits execution. |
| Failure reading `_test/spec.json` | `log.Fatalln` followed by `panic` | Logs missing file error and panics. |
| JSON unmarshal failure | `panic` | Panics with the underlying `json.Unmarshal` error. |
| Failure creating entry in Zip | `log.Fatal` | Logs the error and exits execution. |
| Failure writing content to Zip entry | `log.Fatalf` | Logs target file write error and exits execution. |
| Zip writer close failure | `log.Fatal` | Logs close error and exits execution. |

---

## Expected Command-Line Usage

The subcommand requires one argument representing the output ZIP filename ending in `.zip`:

```bash
go run _tools/gen-oss-fuzz-corpus.go output_corpus.zip
```

### Input File Requirement
* **Path**: `_test/spec.json` relative to the working directory where the script is executed.
* **Format**: JSON array containing objects matching the `TestCase` structure:
  ```json
  [
    {
      "example": 1,
      "markdown": "# Header\nText"
    }
  ]
  ```

### Output Result
A ZIP file at the specified output path containing individual uncompressed/compressed files named `example-<id>` with their raw Markdown text as contents.