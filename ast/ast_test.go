package ast

import "testing"

func TestRemoveChildren(t *testing.T) {
	root := NewDocument()

	node1 := NewDocument()

	node2 := NewDocument()

	root.AppendChild(root, node1)
	root.AppendChild(root, node2)

	root.RemoveChildren(root)

	t.Logf("%+v", node2.PreviousSibling())
}
