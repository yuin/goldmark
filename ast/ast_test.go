package ast

import (
	"reflect"
	"testing"
)

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

func TestWalkAndReplace(t *testing.T) {
	doc := node(NewDocument(), node(NewHeading(1), NewLink(), NewLink()))
	want := []NodeKind{KindDocument, KindHeading, KindHeading, KindHeading}
	var got []NodeKind
	walkerReplace := func(n Node, entering bool) (WalkStatus, error) {
		// We replace any link by an heading
		if entering {
			n, ok := n.(*Link)
			if !ok {
				return WalkContinue, nil
			}
			parent := n.Parent()
			parent.ReplaceChild(parent, n, NewHeading(2))
		}
		return WalkContinue, nil
	}
	walkerCollect := func(n Node, entering bool) (WalkStatus, error) {
		if entering {
			got = append(got, n.Kind())
		}
		return WalkContinue, nil
	}
	if err := Walk(doc, walkerReplace); err != nil {
		t.Fatalf("Walk() error = %v", err)
	}
	if err := Walk(doc, walkerCollect); err != nil {
		t.Fatalf("Walk() error = %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Walk() expected = %v, got = %v", want, got)
	}
}

func node(n Node, children ...Node) Node {
	for _, c := range children {
		n.AppendChild(n, c)
	}
	return n
}
