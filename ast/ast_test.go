package ast

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/yuin/goldmark/text"
)

func TestRemoveChildren(t *testing.T) {
	root := NewDocument()

	node1 := NewDocument()

	node2 := NewDocument()

	root.AppendChild(root, node1)
	root.AppendChild(root, node2)

	root.RemoveChildren(root)

	t.Logf("%+v", node2.PreviousSibling())
}

func TestWalk(t *testing.T) {
	tests := []struct {
		name   string
		node   Node
		want   []NodeKind
		action map[NodeKind]WalkStatus
	}{
		{
			"visits all in depth first order",
			node(NewDocument(), node(NewHeading(1), NewText()), NewLink()),
			[]NodeKind{KindDocument, KindHeading, KindText, KindLink},
			map[NodeKind]WalkStatus{},
		},
		{
			"stops after heading",
			node(NewDocument(), node(NewHeading(1), NewText()), NewLink()),
			[]NodeKind{KindDocument, KindHeading},
			map[NodeKind]WalkStatus{KindHeading: WalkStop},
		},
		{
			"skip children",
			node(NewDocument(), node(NewHeading(1), NewText()), NewLink()),
			[]NodeKind{KindDocument, KindHeading, KindLink},
			map[NodeKind]WalkStatus{KindHeading: WalkSkipChildren},
		},
	}
	for _, tt := range tests {
		var kinds []NodeKind
		collectKinds := func(n Node, entering bool) (WalkStatus, error) {
			if entering {
				kinds = append(kinds, n.Kind())
			}
			if status, ok := tt.action[n.Kind()]; ok {
				return status, nil
			}
			return WalkContinue, nil
		}
		t.Run(tt.name, func(t *testing.T) {
			if err := Walk(tt.node, collectKinds); err != nil {
				t.Errorf("Walk() error = %v", err)
			} else if !reflect.DeepEqual(kinds, tt.want) {
				t.Errorf("Walk() expected = %v, got = %v", tt.want, kinds)
			}
		})
	}
}

func node(n Node, children ...Node) Node {
	for _, c := range children {
		n.AppendChild(n, c)
	}
	return n
}

func TestBaseBlock_Text(t *testing.T) {
	source := []byte(`# Heading

    code block here
	and also here

A paragraph

` + "```" + `somelang
fenced code block
` + "```" + `

The end`)

	t.Run("fetch text from code block", func(t *testing.T) {
		block := NewCodeBlock()
		block.lines = text.NewSegments()
		block.lines.Append(text.Segment{Start: 15, Stop: 31})
		block.lines.Append(text.Segment{Start: 32, Stop: 46})

		expected := []byte("code block here\nand also here\n")
		if !bytes.Equal(expected, block.Text(source)) {
			t.Errorf("Expected: %q, got: %q", string(expected), string(block.Text(source)))
		}
	})

	t.Run("fetch text from fenced code block", func(t *testing.T) {
		block := NewFencedCodeBlock(&Text{
			Segment: text.Segment{Start: 63, Stop: 71},
		})
		block.lines = text.NewSegments()
		block.lines.Append(text.Segment{Start: 72, Stop: 90})

		expectedLang := []byte("somelang")
		if !bytes.Equal(expectedLang, block.Language(source)) {
			t.Errorf("Expected: %q, got: %q", string(expectedLang), string(block.Language(source)))
		}

		expected := []byte("fenced code block\n")
		if !bytes.Equal(expected, block.Text(source)) {
			t.Errorf("Expected: %q, got: %q", string(expected), string(block.Text(source)))
		}
	})
}
