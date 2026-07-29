//go:build !goldmark_v1_attribute

package parser_test

import (
	"testing"

	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/text"
)

func TestUnquotedAttribute(t *testing.T) {
	source := []byte("{key=value&quot;}")
	attrs, ok := parser.ParseAttributes(text.NewReader(source))
	if !ok {
		t.Fatalf("failed to parse attributes")
	}

	// if value contains entity reference, it must be an instance of StringMultilineValue
	if len(attrs) != 1 {
		t.Fatalf("expected 1 attribute, got %d", len(attrs))
	}
	if attrs[0].Name != "key" {
		t.Fatalf("expected attribute name 'key', got '%s'", attrs[0].Name)
	}
	if attrs[0].Value.Str(source) != "value\"" {
		t.Fatalf("expected attribute value 'value\"', got '%s'", attrs[0].Value.Str(source))
	}
	if !attrs[0].Value.IsOwned() {
		t.Fatalf("expected attribute value to be owned, but it is not")
	}

	source = []byte("{key=value}")
	attrs, ok = parser.ParseAttributes(text.NewReader(source))
	if !ok {
		t.Fatalf("failed to parse attributes")
	}

	// if value does not contain entity reference, it must be an instance of IndexMultilineValue
	if len(attrs) != 1 {
		t.Fatalf("expected 1 attribute, got %d", len(attrs))
	}
	if attrs[0].Name != "key" {
		t.Fatalf("expected attribute name 'key', got '%s'", attrs[0].Name)
	}
	if attrs[0].Value.Str(source) != "value" {
		t.Fatalf("expected attribute value 'value', got '%s'", attrs[0].Value.Str(source))
	}
	if attrs[0].Value.IsOwned() {
		t.Fatalf("expected attribute value to be not owned, but it is")
	}
}

func TestQuotedAttribute(t *testing.T) {
	source := []byte("{key=\"value\"}")
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
	if attrs[0].Value.Str(source) != "value" {
		t.Fatalf("expected attribute value 'value', got '%s'", attrs[0].Value.Str(source))
	}
	if attrs[0].Value.IsOwned() {
		t.Fatalf("expected attribute value to be not owned, but it is")
	}
}

func TestIdAndClassAttributes(t *testing.T) {
	source := []byte("{#id .class1 .class2}")
	attrs, ok := parser.ParseAttributes(text.NewReader(source))
	if !ok {
		t.Fatalf("failed to parse attributes")
	}

	if len(attrs) != 2 {
		t.Fatalf("expected 2 attributes, got %d", len(attrs))
	}
	if attrs[0].Name != "id" {
		t.Fatalf("expected attribute name 'id', got '%s'", attrs[0].Name)
	}
	if attrs[0].Value.Str(source) != "id" {
		t.Fatalf("expected attribute value 'id', got '%s'", attrs[0].Value.Str(source))
	}
	if attrs[1].Name != "class" {
		t.Fatalf("expected attribute name 'class', got '%s'", attrs[1].Name)
	}
	if attrs[1].Value.Str(source) != "class1 class2" {
		t.Fatalf("expected attribute value 'class1 class2', got '%s'", attrs[1].Value.Str(source))
	}
}

func TestMultilineAttributes(t *testing.T) {
	source := []byte("{value=\"aaa\n   bbb\"}")
	attrs, ok := parser.ParseAttributes(text.NewReader(source))
	if !ok {
		t.Fatalf("failed to parse attributes")
	}

	if len(attrs) != 1 {
		t.Fatalf("expected 1 attribute, got %d", len(attrs))
	}

	if attrs[0].Name != "value" {
		t.Fatalf("expected attribute name 'value', got '%s'", attrs[0].Name)
	}
	if attrs[0].Value.Str(source) != "aaa\n   bbb" {
		t.Fatalf("expected attribute value 'aaa\\nbbb', got '%s'", attrs[0].Value.Str(source))
	}
	if attrs[0].Value.IsOwned() {
		t.Fatalf("expected attribute value to be not owned, but it is")
	}
}
