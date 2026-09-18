package text_test

import (
	"bytes"
	"testing"

	"github.com/yuin/goldmark/v2/text"
)

func TestMultiLineValueNonContiguousIndices(t *testing.T) {
	source := []byte("x &amp; y\n> z \\* w")
	indices := []text.Index{{Start: 0, Stop: 10}, {Start: 12, Stop: 18}}
	v := text.NewMultiLineValueFromIndices(indices, text.NewDecoder())
	want := "x & y\nz * w"

	if got := v.Value(source); got != want {
		t.Errorf("Value() = %q, want %q", got, want)
	}
	var buf bytes.Buffer
	if _, err := v.WriteTo(&buf, source); err != nil {
		t.Fatalf("WriteTo returned error: %v", err)
	}
	if buf.String() != want {
		t.Errorf("WriteTo() = %q, want %q", buf.String(), want)
	}
}
