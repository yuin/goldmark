package testutil

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// This will fail to compile if the TestingT interface is changed in a way
// that doesn't conform to testing.T.
var _ TestingT = (*testing.T)(nil)

func TestParseTestCaseFile(t *testing.T) {
	tests := []struct {
		desc string
		give string // contents of the test file
		want []MarkdownTestCase
	}{
		{
			desc: "empty",
			give: "",
			want: []MarkdownTestCase{},
		},
		{
			desc: "simple",
			give: strings.Join([]string{
				"1",
				"//- - - - - - - - -//",
				"input",
				"//- - - - - - - - -//",
				"output",
				"//= = = = = = = = = = = = = = = = = = = = = = = =//",
			}, "\n"),
			want: []MarkdownTestCase{
				{
					No:       1,
					Markdown: "input",
					Expected: "output\n",
				},
			},
		},
		{
			desc: "description",
			give: strings.Join([]string{
				"2:check something",
				"//- - - - - - - - -//",
				"hello",
				"//- - - - - - - - -//",
				"<p>hello</p>",
				"//= = = = = = = = = = = = = = = = = = = = = = = =//",
			}, "\n"),
			want: []MarkdownTestCase{
				{
					No:          2,
					Description: "check something",
					Markdown:    "hello",
					Expected:    "<p>hello</p>\n",
				},
			},
		},
		{
			desc: "options",
			give: strings.Join([]string{
				"3",
				`OPTIONS: {"trim": true}`,
				"//- - - - - - - - -//",
				"world",
				"//- - - - - - - - -//",
				"<p>world</p>",
				"//= = = = = = = = = = = = = = = = = = = = = = = =//",
			}, "\n"),
			want: []MarkdownTestCase{
				{
					No:       3,
					Options:  MarkdownTestCaseOptions{Trim: true},
					Markdown: "world",
					Expected: "<p>world</p>\n",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			filename := filepath.Join(t.TempDir(), "give.txt")
			if err := os.WriteFile(filename, []byte(tt.give), 0o644); err != nil {
				t.Fatal(err)
			}

			got, err := ParseTestCaseFile(filename)
			if err != nil {
				t.Fatalf("could not parse: %v", err)
			}

			if !reflect.DeepEqual(tt.want, got) {
				t.Errorf("output did not match:")
				t.Errorf(" got = %#v", got)
				t.Errorf("want = %#v", tt.want)
			}
		})
	}
}
