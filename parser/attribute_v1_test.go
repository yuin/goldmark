//go:build goldmark_v1_attribute

package parser_test

import (
	"bytes"
	"testing"

	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/text"
)

func TestAttributeV1(t *testing.T) {
	source := []byte("{key=[1,2,3]}")
	attrs, ok := parser.ParseAttributes(text.NewReader(source))
	if !ok {
		t.Fatalf("failed to parse attributes")
	}

	if len(attrs) != 1 {
		t.Fatalf("expected 1 attribute, got %d", len(attrs))
	}
	if attrs[0].Name != "key" {
		t.Fatalf("expected attribute name 'key', got '%s'", attrs[0].Name)
	}
	v := attrs[0].Value.Any(source)
	a, ok := v.([]any)
	if !ok {
		t.Fatalf("expected attribute value to be []any, got %T", v)
	}
	if len(a) != 3 {
		t.Fatalf("expected attribute value to have length 3, got %d", len(a))
	}
	if a[0] != float64(1) || a[1] != float64(2) || a[2] != float64(3) {
		t.Fatalf("expected attribute value to be [1,2,3], got %v", a)
	}

	source = []byte("{key=1, key2=\"value with spaces\"}")
	attrs, ok = parser.ParseAttributes(text.NewReader(source))
	if !ok {
		t.Fatalf("failed to parse attributes")
	}
	if len(attrs) != 2 {
		t.Fatalf("expected 2 attributes, got %d", len(attrs))
	}
	if attrs[0].Name != "key" {
		t.Fatalf("expected first attribute name 'key', got '%s'", attrs[0].Name)
	}
	if attrs[1].Name != "key2" {
		t.Fatalf("expected second attribute name 'key2', got '%s'", attrs[1].Name)
	}
	v = attrs[0].Value.Any(source)
	f, ok := v.(float64)
	if !ok {
		t.Fatalf("expected first attribute value to be float64, got %T", v)
	}
	if f != float64(1) {
		t.Fatalf("expected first attribute value to be 1, got %v", v)
	}
	v = attrs[1].Value.Any(source)
	b, ok := v.([]byte)
	if !ok {
		t.Fatalf("expected second attribute value to be []byte, got %T", v)
	}
	if !bytes.Equal(b, []byte("value with spaces")) {
		t.Fatalf("expected second attribute value to be 'value with spaces', got %#v", v)
	}

	source = []byte("{key={\"nested\":[1,2,3], \"another\"=true}}")
	attrs, ok = parser.ParseAttributes(text.NewReader(source))
	if !ok {
		t.Fatalf("failed to parse attributes")
	}

	if len(attrs) != 1 {
		t.Fatalf("expected 1 attribute, got %d", len(attrs))
	}
	if attrs[0].Name != "key" {
		t.Fatalf("expected attribute name 'key', got '%s'", attrs[0].Name)
	}
	v = attrs[0].Value.Any(source)
	m, ok := v.(map[string]any)
	if !ok {
		t.Fatalf("expected attribute value to be map[string]any, got %T", v)
	}
	if len(m) != 2 {
		t.Fatalf("expected attribute value to have length 2, got %d", len(m))
	}
	nested, ok := m["nested"].([]any)
	if !ok {
		t.Fatalf("expected nested attribute value to be []any, got %T", m["nested"])
	}
	if len(nested) != 3 || nested[0] != float64(1) || nested[1] != float64(2) || nested[2] != float64(3) {
		t.Fatalf("expected nested attribute value to be [1,2,3], got %v", nested)
	}
	if m["another"] != true {
		t.Fatalf("expected another attribute value to be true, got %v", m["another"])
	}

}
