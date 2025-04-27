package goldmark_test

import (
	"bytes"
	"testing"

	. "github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/testutil"
	"github.com/yuin/goldmark/text"
)

func TestASTBlockNodeText(t *testing.T) {
	var cases = []struct {
		Name   string
		Source string
		T1     string
		T2     string
		C      bool
	}{
		{
			Name: "AtxHeading",
			Source: `# l1

a

# l2`,
			T1: `l1`,
			T2: `l2`,
		},
		{
			Name: "SetextHeading",
			Source: `l1
l2
===============

a

l3
l4
==============`,
			T1: `l1
l2`,
			T2: `l3
l4`,
		},
		{
			Name: "CodeBlock",
			Source: `    l1
    l2

a

    l3
	l4`,
			T1: `l1
l2
`,
			T2: `l3
l4
`,
		},
		{
			Name: "FencedCodeBlock",
			Source: "```" + `
l1
l2
` + "```" + `

a

` + "```" + `
l3
l4`,
			T1: `l1
l2
`,
			T2: `l3
l4
`,
		},
		{
			Name: "Blockquote",
			Source: `> l1
> l2

a

> l3
> l4`,
			T1: `l1
l2`,
			T2: `l3
l4`,
		},
		{
			Name: "List",
			Source: `- l1
  l2

a

- l3
  l4`,
			T1: `l1
l2`,
			T2: `l3
l4`,
			C: true,
		},
		{
			Name: "HTMLBlock",
			Source: `<div>
l1
l2
</div>

a

<div>
l3
l4`,
			T1: `<div>
l1
l2
</div>
`,
			T2: `<div>
l3
l4`,
		},
	}

	for _, cs := range cases {
		t.Run(cs.Name, func(t *testing.T) {
			s := []byte(cs.Source)
			md := New()
			n := md.Parser().Parse(text.NewReader(s))
			c1 := n.FirstChild()
			c2 := c1.NextSibling().NextSibling()
			if cs.C {
				c1 = c1.FirstChild()
				c2 = c2.FirstChild()
			}
			if !bytes.Equal(c1.Text(s), []byte(cs.T1)) { // nolint: staticcheck

				t.Errorf("%s unmatch: %s", cs.Name, testutil.DiffPretty(c1.Text(s), []byte(cs.T1))) // nolint: staticcheck

			}
			if !bytes.Equal(c2.Text(s), []byte(cs.T2)) { // nolint: staticcheck

				t.Errorf("%s(EOF) unmatch: %s", cs.Name, testutil.DiffPretty(c2.Text(s), []byte(cs.T2))) // nolint: staticcheck

			}
		})
	}

}

func TestASTInlineNodeText(t *testing.T) {
	var cases = []struct {
		Name   string
		Source string
		T1     string
	}{
		{
			Name:   "CodeSpan",
			Source: "`c1`",
			T1:     `c1`,
		},
		{
			Name:   "Emphasis",
			Source: `*c1 **c2***`,
			T1:     `c1 c2`,
		},
		{
			Name:   "Link",
			Source: `[label](url)`,
			T1:     `label`,
		},
		{
			Name:   "AutoLink",
			Source: `<http://url>`,
			T1:     `http://url`,
		},
		{
			Name:   "RawHTML",
			Source: `<span>c1</span>`,
			T1:     `<span>`,
		},
	}

	for _, cs := range cases {
		t.Run(cs.Name, func(t *testing.T) {
			s := []byte(cs.Source)
			md := New()
			n := md.Parser().Parse(text.NewReader(s))
			c1 := n.FirstChild().FirstChild()
			if !bytes.Equal(c1.Text(s), []byte(cs.T1)) { // nolint: staticcheck
				t.Errorf("%s unmatch:\n%s", cs.Name, testutil.DiffPretty(c1.Text(s), []byte(cs.T1))) // nolint: staticcheck
			}
		})
	}

}

func TestHasBlankPreviousLines(t *testing.T) {
	var cases = []struct {
		Name     string
		Source   string
		Node     func(n ast.Node) ast.Node
		Expected bool
	}{
		{
			Name: "nesting paragraphs in blockquotes",
			Source: `
> a
> 
> b
`,
			Node: func(n ast.Node) ast.Node {
				return n.FirstChild().FirstChild().NextSibling()
			},
			Expected: true,
		},
		{
			Name: "nesting HTML blocks in blockquotes",
			Source: `
> <!-- a -->
> 
> <!-- b -->
`,
			Node: func(n ast.Node) ast.Node {
				return n.FirstChild().FirstChild().NextSibling()
			},
			Expected: true,
		},
		{
			Name: "nesting HTML blocks in blockquotes",
			Source: `
> <!-- a -->
> <!-- b -->
`,
			Node: func(n ast.Node) ast.Node {
				return n.FirstChild().FirstChild().NextSibling()
			},
			Expected: false,
		},
		{
			Name: "nesting loose lists in blockquotes",
			Source: `
> - a
> 
> - b
`,
			Node: func(n ast.Node) ast.Node {
				return n.FirstChild().FirstChild().FirstChild().NextSibling()
			},
			Expected: true,
		},
		{
			Name: "nesting tight lists in blockquotes",
			Source: `
> - a
> - b
`,
			Node: func(n ast.Node) ast.Node {
				return n.FirstChild().FirstChild().FirstChild().NextSibling()
			},
			Expected: false,
		},
		{
			Name: "nesting paragraphs in lists",
			Source: `
- a

  b
`,
			Node: func(n ast.Node) ast.Node {
				return n.FirstChild().FirstChild().FirstChild().NextSibling()
			},
			Expected: true,
		},
	}
	md := New()
	for _, cs := range cases {
		t.Run(cs.Name, func(t *testing.T) {
			n := md.Parser().Parse(text.NewReader([]byte(cs.Source)))
			if cs.Node(n).HasBlankPreviousLines() != cs.Expected {
				t.Errorf("expected %v, got %v", cs.Expected, !cs.Expected)
			}
		})
	}
}
