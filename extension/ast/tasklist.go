package ast

import (
	"fmt"

	gast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// A TaskCheckBox struct represents a checkbox of a task list.
type TaskCheckBox struct {
	gast.BaseInline

	// Segment is a position in a source text.
	Segment text.Segment

	IsChecked bool
}

// Dump implements Node.Dump.
func (n *TaskCheckBox) Dump(source []byte, level int) {
	m := map[string]string{
		"Checked": fmt.Sprintf("%v", n.IsChecked),
	}
	gast.DumpHelper(n, source, level, m, nil)
}

// KindTaskCheckBox is a NodeKind of the TaskCheckBox node.
var KindTaskCheckBox = gast.NewNodeKind("TaskCheckBox")

// Kind implements Node.Kind.
func (n *TaskCheckBox) Kind() gast.NodeKind {
	return KindTaskCheckBox
}

// NewTaskCheckBox returns a new TaskCheckBox node.
func NewTaskCheckBox(checked bool) *TaskCheckBox {
	return &TaskCheckBox{
		IsChecked: checked,
	}
}

// NewTaskCheckBoxSegment returns a new TaskCheckBox node with the given source position.
func NewTaskCheckBoxSegment(checked bool, segment text.Segment) *TaskCheckBox {
	return &TaskCheckBox{
		IsChecked: checked,
		Segment:   segment,
	}
}
