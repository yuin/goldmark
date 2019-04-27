package goldmark

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/yuin/goldmark/renderer/html"
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
	markdown := New(WithRendererOptions(
		html.WithXHTML(),
		html.WithUnsafe(),
	))
	for _, testCase := range testCases {
		var out bytes.Buffer
		if err := markdown.Convert([]byte(testCase.Markdown), &out); err != nil {
			panic(err)
		}
		if !bytes.Equal(bytes.TrimSpace(out.Bytes()), bytes.TrimSpace([]byte(testCase.HTML))) {
			format := `============= case %d ================
Markdown:
-----------
%s

Expected:
----------
%s

Actual
---------
%s
`
			t.Errorf(format, testCase.Example, testCase.Markdown, testCase.HTML, out.Bytes())
		}
	}
}
