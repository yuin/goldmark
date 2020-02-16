package goldmark_test

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	. "github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/testutil"
)

type commonmarkSpecTestCase struct {
	Markdown  string `json:"markdown"`
	HTML      string `json:"html"`
	Example   int    `json:"example"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
	Section   string `json:"section"`
}

func TestSpec(t *testing.T) {
	bs, err := ioutil.ReadFile("_test/spec.json")
	if err != nil {
		panic(err)
	}
	var testCases []commonmarkSpecTestCase
	if err := json.Unmarshal(bs, &testCases); err != nil {
		panic(err)
	}
	cases := []testutil.MarkdownTestCase{}
	for _, c := range testCases {
		cases = append(cases, testutil.MarkdownTestCase{
			No:       c.Example,
			Markdown: c.Markdown,
			Expected: c.HTML,
		})
	}
	markdown := New(WithRendererOptions(
		html.WithXHTML(),
		html.WithUnsafe(),
	))
	testutil.DoTestCases(markdown, cases, t)
}

func TestSpec_EdgeCase_LinkWithEmptyText(t *testing.T) {
	// TODO: maybe this test cases will be part of the official spec in the future.
	//       check: https://github.com/commonmark/commonmark-spec/issues/636

	cases := []testutil.MarkdownTestCase{
		testutil.MarkdownTestCase{
			No:       -1,
			Markdown: "[](./target.md)",
			Expected: "<p><a href=\"./target.md\"></a></p>",
		},
		testutil.MarkdownTestCase{
			No:       -1,
			Markdown: "[]()",
			Expected: "<p><a href=\"\"></a></p>",
		},
	}
	markdown := New(WithRendererOptions(
		html.WithXHTML(),
		html.WithUnsafe(),
	))
	testutil.DoTestCases(markdown, cases, t)
}
