package ast

import (
	"reflect"
	"testing"
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

func TestNodeKindString(t *testing.T) {
	tests := []struct {
		name string
		give NodeKind
		want string
	}{
		{
			name: "known",
			give: KindLink,
			want: "Link",
		},
		{
			name: "unknown",
			give: NodeKind(500),
			want: "NodeKind(500)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.give.String()
			if tt.want != got {
				t.Errorf("String() expected %q, got %q", tt.want, got)
			}
		})
	}
}
