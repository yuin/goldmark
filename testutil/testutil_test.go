package testutil

import (
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"
)

// This will fail to compile if the TestingT interface is changed in a way
// that doesn't conform to testing.T.
var _ TestingT = (*testing.T)(nil)

func TestParseTestCases(t *testing.T) {
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
			got, err := ParseTestCases(strings.NewReader(tt.give))
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

func TestParseTestCases_Errors(t *testing.T) {
	tests := []struct {
		desc   string
		give   string // contents of the test file
		errMsg string
	}{
		{
			desc: "bad number/no description",
			give: strings.Join([]string{
				"1 not a number",
				"//- - - - - - - - -//",
				"world",
				"//- - - - - - - - -//",
				"<p>world</p>",
				"//= = = = = = = = = = = = = = = = = = = = = = = =//",
			}, "\n"),
			errMsg: "line 1: invalid case No",
		},
		{
			desc: "bad number/description",
			give: strings.Join([]string{
				"1 not a number:description",
				"//- - - - - - - - -//",
				"world",
				"//- - - - - - - - -//",
				"<p>world</p>",
				"//= = = = = = = = = = = = = = = = = = = = = = = =//",
			}, "\n"),
			errMsg: "line 1: invalid case No",
		},
		{
			desc: "eof after number",
			give: strings.Join([]string{
				"1",
			}, "\n"),
			errMsg: "line 1: invalid case: expected content after",
		},
		{
			desc: "bad options",
			give: strings.Join([]string{
				"3",
				`OPTIONS: {not valid JSON}`,
				"//- - - - - - - - -//",
				"world",
				"//- - - - - - - - -//",
				"<p>world</p>",
				"//= = = = = = = = = = = = = = = = = = = = = = = =//",
			}, "\n"),
			errMsg: "line 2: invalid options:",
		},
		{
			desc: "bad separator",
			give: strings.Join([]string{
				"3",
				"// not the right separator //",
			}, "\n"),
			errMsg: `line 2: invalid separator "// not the right separator //"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			cases, err := ParseTestCases(strings.NewReader(tt.give))
			if err == nil {
				t.Fatalf("expected error, got:\n%#v", cases)
			}

			if got := err.Error(); !strings.Contains(got, tt.errMsg) {
				t.Errorf("unexpected error message:")
				t.Errorf("             got = %v", got)
				t.Errorf("does not contain = %v", tt.errMsg)
			}
		})
	}
}

func TestParseTestCaseFile_Error(t *testing.T) {
	cases, err := ParseTestCaseFile("does_not_exist.txt")
	if err == nil {
		t.Fatalf("expected error, got:\n%#v", cases)
	}

	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("  unexpected error = %v", err)
		t.Errorf("expected unwrap to = %v", os.ErrNotExist)
	}
}

func TestTestCaseParseError(t *testing.T) {
	wrapped := errors.New("great sadness")
	err := &testCaseParseError{Line: 42, Err: wrapped}

	t.Run("Error", func(t *testing.T) {
		want := "line 42: great sadness"
		got := err.Error()
		if want != got {
			t.Errorf("Error() = %q, want %q", got, want)
		}
	})

	t.Run("Unwrap", func(t *testing.T) {
		if !errors.Is(err, wrapped) {
			t.Errorf("error %#v should unwrap to %#v", err, wrapped)
		}
	})
}
