package ast

import (
	"fmt"
	gast "github.com/yuin/goldmark/ast"
)

// A TaskCheckBox struct represents a checkbox of a task list.
type TaskCheckBox struct {
	gast.BaseInline
	IsChecked bool
}

// Dump impelemtns Node.Dump.
func (n *TaskCheckBox) Dump(source []byte, level int) {
	m := map[string]string{
		"Checked": fmt.Sprintf("%v", n.IsChecked),
	}
	gast.DumpHelper(n, source, level, "TaskCheckBox", m, nil)
}

// NewTaskCheckBox returns a new TaskCheckBox node.
func NewTaskCheckBox(checked bool) *TaskCheckBox {
	return &TaskCheckBox{
		IsChecked: checked,
	}
}
