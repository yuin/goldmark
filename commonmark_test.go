package goldmark_test

import (
	"encoding/json"
	"os"
	"slices"
	"testing"

	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer/html"
	"github.com/yuin/goldmark/v2/testutil"
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
	bs, err := os.ReadFile("_test/spec.json")
	if err != nil {
		panic(err)
	}
	var testCases []commonmarkSpecTestCase
	if err := json.Unmarshal(bs, &testCases); err != nil {
		panic(err)
	}
	cases := []testutil.MarkdownTestCase{}
	nos := testutil.ParseCliCaseArg()
	for _, c := range testCases {
		shouldAdd := len(nos) == 0
		if !shouldAdd {
			shouldAdd = slices.Contains(nos, c.Example)
		}

		if shouldAdd {
			cases = append(cases, testutil.MarkdownTestCase{
				No:       c.Example,
				Markdown: c.Markdown,
				Expected: c.HTML,
			})
		}
	}
	markdown := testutil.NewMarkdownToStringFunc(
		parser.New(),
		html.New(html.WithXHTML(), html.WithUnsafe()),
	)
	testutil.DoTestCases(markdown, cases, t)
}
