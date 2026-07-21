package parser

import (
	"testing"

	"github.com/yuin/goldmark/v2/text"
)

func TestParseAttributesNumericAndArray(t *testing.T) {
	// Regression for #563: Hugo-style highlighter attrs with numeric values / arrays.
	src := []byte(`{hl_lines=[8,"15-17"],linenostart=199}`)
	r := text.NewReader(src)
	attrs, ok := ParseAttributes(r)
	if !ok {
		t.Fatal("ParseAttributes returned false")
	}
	got := map[string]string{}
	for _, a := range attrs {
		got[a.Name] = a.Value.Str(nil)
	}
	if got["linenostart"] != "199" {
		t.Fatalf("linenostart=%q want 199", got["linenostart"])
	}
	if got["hl_lines"] != `8 15-17` {
		t.Fatalf("hl_lines=%q want %q", got["hl_lines"], `8 15-17`)
	}
}

func TestParseAttributesQuotedAndWord(t *testing.T) {
	src := []byte(`{id=main class=foo bar="baz qux"}`)
	r := text.NewReader(src)
	attrs, ok := ParseAttributes(r)
	if !ok {
		t.Fatal("ParseAttributes returned false")
	}
	got := map[string]string{}
	for _, a := range attrs {
		got[a.Name] = a.Value.Str(nil)
	}
	if got["id"] != "main" {
		t.Fatalf("id=%q", got["id"])
	}
	if got["class"] != "foo" {
		t.Fatalf("class=%q", got["class"])
	}
	if got["bar"] != "baz qux" {
		t.Fatalf("bar=%q", got["bar"])
	}
}
