package ast_test

import (
	"fmt"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

func ExampleWalk() {
	// Extract links from markdown text
	src := []byte(`Some links: <https://golang.org/>, [goldmark repo](https://github.com/yuin/goldmark)`)

	fn := func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch n.Kind() {
		case ast.KindLink, ast.KindAutoLink:
		default:
			return ast.WalkContinue, nil
		}
		if l, ok := n.(*ast.AutoLink); ok {
			fmt.Printf("auto link: %s\n", l.URL(src))
		}
		if l, ok := n.(*ast.Link); ok {
			fmt.Printf("link: %s\n", l.Destination)
		}
		return ast.WalkContinue, nil
	}

	node := goldmark.DefaultParser().Parse(text.NewReader(src))
	_ = ast.Walk(node, fn)
	// Output:
	//
	// auto link: https://golang.org/
	// link: https://github.com/yuin/goldmark
}
