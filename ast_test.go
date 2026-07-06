package goldmark_test

import (
	"testing"

	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/parser"
)

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
	for _, cs := range cases {
		t.Run(cs.Name, func(t *testing.T) {
			n := parser.New().Parse([]byte(cs.Source))
			if cs.Node(n).(ast.BlockNode).HasBlankPreviousLines() != cs.Expected {
				t.Errorf("expected %v, got %v", cs.Expected, !cs.Expected)
			}
		})
	}
}

func TestInlinePos(t *testing.T) {
	source := []byte(`[bar][]

[foo][bar]

[bar]

[foo](http://example.com)

aaaa **b** 

![aaa](http://example.com/foo.png "title")

[bar]: 
  /url "ti
  tle"
`)
	n := parser.New().Parse(source)
	if n.FirstChild().FirstChild().Pos() != 0 {
		t.Error("unexpected position for 1st link reference")
	}
	if n.FirstChild().NextSibling().FirstChild().Pos() != 9 {
		t.Error("unexpected position for 2nd link reference")
	}
	if n.FirstChild().NextSibling().NextSibling().FirstChild().Pos() != 21 {
		t.Error("unexpected position for 3rd link reference")
	}
	if n.FirstChild().NextSibling().NextSibling().NextSibling().FirstChild().Pos() != 28 {
		t.Error("unexpected position for 1st inline link ")
	}
	if n.FirstChild().NextSibling().NextSibling().NextSibling().NextSibling().FirstChild().NextSibling().Pos() != 60 {
		t.Error("unexpected position for 1st emphasis")
	}
	if n.FirstChild().NextSibling().NextSibling().NextSibling().NextSibling().NextSibling().FirstChild().Pos() != 68 {
		t.Error("unexpected position for 1st image")
	}
}

func TestBlockPos(t *testing.T) {
	source := []byte(`paragraph text

## heading

---

> blockquote

- list item

1. ordered

    indented code

` + "```go\nfenced code\n```" + `

<!-- html block -->

[bar]: /url`)
	n := parser.New().Parse(source)
	paragraph := n.FirstChild()
	heading := paragraph.NextSibling()
	thematicBreak := heading.NextSibling()
	blockquote := thematicBreak.NextSibling()
	unorderedList := blockquote.NextSibling()
	orderedList := unorderedList.NextSibling()
	codeBlock := orderedList.NextSibling()
	htmlBlock := codeBlock.NextSibling()
	linkRefDef := htmlBlock.NextSibling()
	if paragraph.Pos() != 0 {
		t.Error("unexpected position for paragraph")
	}
	if heading.Pos() != 16 {
		t.Error("unexpected position for heading")
	}
	if thematicBreak.Pos() != 28 {
		t.Error("unexpected position for thematic break")
	}
	if blockquote.Pos() != 33 {
		t.Error("unexpected position for blockquote")
	}
	if unorderedList.Pos() != 47 {
		t.Error("unexpected position for unordered list")
	}
	if orderedList.Pos() != 60 {
		t.Error("unexpected position for ordered list")
	}
	if codeBlock.Pos() != 91 {
		t.Error("unexpected position for fenced code block")
	}
	if htmlBlock.Pos() != 114 {
		t.Error("unexpected position for html block")
	}
	if linkRefDef.Pos() != 135 {
		t.Error("unexpected position for link reference definition")
	}
}
