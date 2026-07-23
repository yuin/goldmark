package goldmark_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yuin/goldmark"
)

func TestConvertStripsUTF8BOM(t *testing.T) {
	src := append([]byte{0xEF, 0xBB, 0xBF}, []byte("# Hi")...)
	var buf bytes.Buffer
	if err := goldmark.Convert(src, &buf); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "<h1") || !strings.Contains(out, "Hi") {
		t.Fatalf("unexpected: %q", out)
	}
	// BOM should not appear as content
	if strings.Contains(out, "\ufeff") {
		t.Fatalf("BOM leaked into output: %q", out)
	}
}
