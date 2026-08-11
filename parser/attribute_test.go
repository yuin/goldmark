//go:build !goldmark_v1_attribute

package parser_test

import (
	"testing"

	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/text"
)

func TestParseAttributesCommaSeparated(t *testing.T) {
	type expectedAttribute struct {
		name  string
		value string
	}
	tests := []struct {
		name   string
		source string
		want   []expectedAttribute
	}{
		{
			name:   "numeric values",
			source: `{start=1,end=2}`,
			want:   []expectedAttribute{{"start", "1"}, {"end", "2"}},
		},
		{
			name:   "bracketed value",
			source: `{lines=[8,15,17],start=199}`,
			want:   []expectedAttribute{{"lines", "[8,15,17]"}, {"start", "199"}},
		},
		{
			name:   "quoted array element",
			source: `{hl\_lines=[8,"15,17"],linenostart=199}`,
			want:   []expectedAttribute{{`hl\_lines`, `[8,"15,17"]`}, {"linenostart", "199"}},
		},
		{
			name:   "issue 563",
			source: `{hl\_lines=[8,"15-17"],linenostart=199}`,
			want:   []expectedAttribute{{`hl\_lines`, `[8,"15-17"]`}, {"linenostart", "199"}},
		},
		{
			name:   "space separated",
			source: `{start=1 end=2}`,
			want:   []expectedAttribute{{"start", "1"}, {"end", "2"}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			source := []byte(test.source)
			attrs, ok := parser.ParseAttributes(text.NewReader(source))
			if !ok {
				t.Fatal("ParseAttributes failed")
			}
			if len(attrs) != len(test.want) {
				t.Fatalf("ParseAttributes returned %d attributes, want %d: %#v", len(attrs), len(test.want), attrs)
			}
			for i, attr := range attrs {
				if attr.Name != test.want[i].name {
					t.Errorf("attribute %d name = %q, want %q", i, attr.Name, test.want[i].name)
				}
				if got := attr.Value.Str(source); got != test.want[i].value {
					t.Errorf("attribute %d value = %q, want %q", i, got, test.want[i].value)
				}
			}
		})
	}
}

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
