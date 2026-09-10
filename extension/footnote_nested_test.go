package extension

import (
	"bytes"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer/html"
	"github.com/yuin/goldmark/v2/testutil"
	"github.com/yuin/goldmark/v2/util"
)

func BenchmarkFootnoteRender(b *testing.B) {
	for _, name := range []string{"flat", "unused"} {
		b.Run(name, func(b *testing.B) {
			var input strings.Builder
			input.WriteString("Document.\n\n")
			if name == "flat" {
				for i := range 50 {
					fmt.Fprintf(&input, "Root[^n%d] ", i)
				}
				input.WriteString("\n\n")
			}
			for i := range 50 {
				fmt.Fprintf(&input, "[^n%d]: Some footnote text for item %d.", i, i)
				if name == "unused" {
					fmt.Fprintf(&input, "[^n%d]", (i+1)%50)
				}
				input.WriteString("\n\n")
			}
			source := util.StringToReadOnlyBytes(input.String())
			p := parser.New(parser.WithExtensions(NewFootnoteParser()))
			r := html.New(html.WithExtensions(NewFootnoteHTMLRenderer()))
			var output bytes.Buffer
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				output.Reset()
				if err := r.Render(&output, source, p.Parse(source)); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func TestFootnoteNestedReferences(t *testing.T) {
	definitions := regexp.MustCompile(`<li id="([^"]+)"`)
	backlinks := regexp.MustCompile(`href="#([^"]+)" class="footnote-backref"`)
	ids := regexp.MustCompile(` id="([^"]+)"`)
	hrefs := regexp.MustCompile(`href="#([^"]+)"`)
	values := func(pattern *regexp.Regexp, output string) []string {
		var result []string
		for _, match := range pattern.FindAllStringSubmatch(output, -1) {
			result = append(result, match[1])
		}
		return result
	}

	for _, tc := range []struct {
		name        string
		source      string
		options     []FootnoteHTMLRendererOption
		definitions []string
		backlinks   []string
		contains    []string
		omits       []string
	}{
		{
			name:        "nested reference",
			source:      "Parent[^p]\n\n[^p]: parent-body[^c]\n[^c]: child-body\n",
			definitions: []string{"fn:1", "fn:2"},
			backlinks:   []string{"fnref:1", "fnref:2"},
			contains:    []string{"parent-body", "child-body"},
		},
		{
			name:        "chain",
			source:      "Root[^a]\n\n[^a]: a-body[^b]\n[^b]: b-body[^c]\n[^c]: c-body\n",
			definitions: []string{"fn:1", "fn:2", "fn:3"},
			backlinks:   []string{"fnref:1", "fnref:2", "fnref:3"},
			contains:    []string{"a-body", "b-body", "c-body"},
		},
		{
			name:        "shared child",
			source:      "First[^a] and second[^b]\n\n[^a]: a-body[^c]\n[^b]: b-body[^c]\n[^c]: c-body\n",
			definitions: []string{"fn:1", "fn:2", "fn:3"},
			backlinks:   []string{"fnref:1", "fnref:2", "fnref:3", "fnref1:3"},
			contains:    []string{"a-body", "b-body", "c-body"},
		},
		{
			name:        "backlinks keep their reference order",
			source:      "Root[^a]\n\n[^a]: a-body[^b]\n\nChild[^b]\n\n[^b]: b-body\n",
			definitions: []string{"fn:1", "fn:2"},
			backlinks:   []string{"fnref:1", "fnref:2", "fnref1:2"},
			contains:    []string{"a-body", "b-body"},
		},
		{
			name:        "self reference",
			source:      "Root[^a]\n\n[^a]: a-body[^a]\n",
			definitions: []string{"fn:1"},
			backlinks:   []string{"fnref:1", "fnref1:1"},
			contains:    []string{"a-body"},
		},
		{
			name:        "cycle",
			source:      "Root[^a]\n\n[^a]: a-body[^b]\n[^b]: b-body[^a]\n",
			definitions: []string{"fn:1", "fn:2"},
			backlinks:   []string{"fnref:1", "fnref1:1", "fnref:2"},
			contains:    []string{"a-body", "b-body"},
		},
		{
			name:        "unreferenced chain stays hidden",
			source:      "Root[^a]\n\n[^a]: a-body\n[^u]: unused-body[^v]\n[^v]: hidden-body\n",
			definitions: []string{"fn:1"},
			backlinks:   []string{"fnref:1"},
			contains:    []string{"a-body"},
			omits:       []string{"unused-body", "hidden-body"},
		},
		{
			name:     "unreferenced cycle stays hidden",
			source:   "Text.\n\n[^u]: unused-body[^v]\n[^v]: hidden-body[^u]\n",
			contains: []string{"<p>Text.</p>"},
			omits:    []string{"unused-body", "hidden-body", "doc-endnotes"},
		},
		{
			name:        "definition inside an unreferenced definition",
			source:      "Child[^c]\n\n[^p]: parent-body\n\n    [^c]: child-body\n",
			definitions: []string{"fn:1"},
			backlinks:   []string{"fnref:1"},
			contains:    []string{"child-body"},
			omits:       []string{"parent-body"},
		},
		{
			name:        "prefixed references and metadata",
			source:      "Parent[^p]\n\n[^p]: parent-body[^c]\n[^c]: child-body\n",
			options:     []FootnoteHTMLRendererOption{WithIDPrefix("doc-"), WithLinkTitle("refs=%%")},
			definitions: []string{"doc-fn:1", "doc-fn:2"},
			backlinks:   []string{"doc-fnref:1", "doc-fnref:2"},
			contains:    []string{"child-body", `title="refs=1"`},
			omits:       []string{`title="refs=0"`},
		},
		{
			name:        "reference to a nested definition",
			source:      "Parent[^p]\n\n[^p]: parent-body[^c]\n\n    [^c]: child-body\n",
			definitions: []string{"fn:1", "fn:2"},
			backlinks:   []string{"fnref:1", "fnref:2"},
			contains:    []string{"parent-body", "child-body"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if caught := recover(); caught != nil {
					t.Errorf("render panicked: %v", caught)
				}
			}()
			markdown := testutil.NewMarkdownToStringFunc(
				parser.New(parser.WithExtensions(NewFootnoteParser())),
				html.New(html.WithExtensions(NewFootnoteHTMLRenderer(tc.options...))),
			)
			output, err := markdown(tc.source)
			if err != nil {
				t.Fatal(err)
			}
			if got := values(definitions, output); !slices.Equal(got, tc.definitions) {
				t.Errorf("definitions = %v, want %v", got, tc.definitions)
			}
			if got := values(backlinks, output); !slices.Equal(got, tc.backlinks) {
				t.Errorf("backlinks = %v, want %v", got, tc.backlinks)
			}
			seen := map[string]bool{}
			for _, id := range values(ids, output) {
				if seen[id] {
					t.Errorf("duplicate id %q", id)
				}
				seen[id] = true
			}
			for _, target := range values(hrefs, output) {
				if !seen[target] {
					t.Errorf("missing target for href #%s", target)
				}
			}
			for _, want := range tc.contains {
				if !strings.Contains(output, want) {
					t.Errorf("output does not contain %q", want)
				}
			}
			for _, unwanted := range tc.omits {
				if strings.Contains(output, unwanted) {
					t.Errorf("output contains %q", unwanted)
				}
			}
		})
	}
}
