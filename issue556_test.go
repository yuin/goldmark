package goldmark_test

import (
	"bytes"
	"testing"

	"github.com/yuin/goldmark"
)

// TestFencedCodeBlockBlankLineNoPanic is a regression test for
// https://github.com/yuin/goldmark/issues/556: the fenced code block parser
// panicked with "index out of range [-1]" when opened on a blank / all-whitespace
// line, where BlockIndent returns -1. Minimal fuzzer input: "*\n\t* \t~".
func TestFencedCodeBlockBlankLineNoPanic(t *testing.T) {
	var buf bytes.Buffer
	if err := goldmark.Convert([]byte("*\n\t* \t~"), &buf); err != nil {
		t.Fatal(err)
	}
}
